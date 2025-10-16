/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"context"
	"net/http"
	"strings"
	"sync"

	auth "go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/authz"
	authzOpts "go.kubeguard.dev/guard/authz/providers/azure/options"
	"go.kubeguard.dev/guard/authz/providers/azure/rbac"
	azureutils "go.kubeguard.dev/guard/util/azure"
	errutils "go.kubeguard.dev/guard/util/error"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	authzv1 "k8s.io/api/authorization/v1"
	"k8s.io/klog/v2"
)

const (
	OrgType = "azure"
)

type contextKey string

const (
	requestIDContextKey contextKey = "request-id"
)

var (
	once   sync.Once
	client authz.Interface
	err    error
)

func init() {
	authz.SupportedOrgs = append(authz.SupportedOrgs, OrgType)
}

type Authorizer struct {
	rbacClient           *rbac.AccessInfo
	httpClientRetryCount int
}

func New(opts authzOpts.Options, authopts auth.Options) (authz.Interface, error) {
	once.Do(func() {
		klog.Info("Creating Azure global authz client")
		client, err = newAuthzClient(opts, authopts)
		if client == nil || err != nil {
			klog.Fatalf("Authz RBAC client creation failed. Error: %s", err)
		}
	})
	return client, err
}

func newAuthzClient(opts authzOpts.Options, authopts auth.Options) (authz.Interface, error) {
	c := &Authorizer{
		httpClientRetryCount: authopts.HttpClientRetryCount,
	}

	authzInfoVal, err := getAuthzInfo(authopts.Environment)
	if err != nil {
		return nil, errors.Wrap(err, "Error in getAuthzInfo %s")
	}

	c.rbacClient, err = rbac.New(opts, authopts, authzInfoVal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ms rbac client")
	}

	return c, nil
}

func (s Authorizer) Check(ctx context.Context, request *authzv1.SubjectAccessReviewSpec, store authz.Store) (*authzv1.SubjectAccessReviewStatus, error) {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDContextKey, requestID)

	if request == nil {
		klog.ErrorS(errors.New("subject access review is nil"), "Authorization request failed", "requestID", requestID)
		return nil, errutils.WithCode(errors.New("subject access review is nil"), http.StatusBadRequest)
	}

	klog.InfoS("Authorization check started", "requestID", requestID)

	// check if user is system accounts
	if strings.HasPrefix(strings.ToLower(request.User), "system:") {
		klog.V(10).InfoS("Returning no op to system accounts", "requestID", requestID)
		return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NoOpinionVerdict}, nil
	}

	if s.rbacClient.SkipAuthzCheck(request) {
		klog.V(3).InfoS("User is part of skip authz list, returning no op", "requestID", requestID)
		return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NoOpinionVerdict}, nil
	}

	if _, ok := request.Extra["oid"]; !ok {
		if s.rbacClient.ShouldSkipAuthzCheckForNonAADUsers() {
			klog.V(5).InfoS("Non-AAD user, returning no op", "requestID", requestID)
			return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NonAADUserNoOpVerdict}, nil
		} else {
			klog.InfoS("Non-AAD user denied", "requestID", requestID)
			return &authzv1.SubjectAccessReviewStatus{Allowed: false, Denied: true, Reason: rbac.NonAADUserNotAllowedVerdict}, nil
		}
	}

	exist, result := s.rbacClient.GetResultFromCache(request, store)

	if exist {
		klog.V(5).InfoS("Cache hit", "requestID", requestID, "allowed", result)
		if result {
			return &authzv1.SubjectAccessReviewStatus{Allowed: result, Reason: rbac.AccessAllowedVerdict}, nil
		} else {
			return &authzv1.SubjectAccessReviewStatus{Allowed: result, Denied: true, Reason: rbac.AccessNotAllowedVerdict}, nil
		}
	}

	// if set true, webhook will allow access to discovery APIs for authenticated users. If false, access check will be performed on Azure.
	if s.rbacClient.AllowNonResPathDiscoveryAccess(request) {
		klog.V(10).InfoS("Allowing user access for discovery check", "requestID", requestID)
		_ = s.rbacClient.SetResultInCache(request, true, store)
		return &authzv1.SubjectAccessReviewStatus{Allowed: true, Reason: rbac.AccessAllowedVerdict}, nil
	}

	ctx = azureutils.WithRetryableHttpClient(ctx, s.httpClientRetryCount)

	if s.rbacClient.IsTokenExpired() {
		klog.V(5).InfoS("Token expired, refreshing", "requestID", requestID)
		if err := s.rbacClient.RefreshToken(ctx, requestID); err != nil {
			klog.ErrorS(err, "Token refresh failed", "requestID", requestID)
			return nil, errutils.WithCode(err, http.StatusInternalServerError)
		}
		klog.V(5).InfoS("Token refreshed successfully", "requestID", requestID)
	}

	response, err := s.rbacClient.CheckAccess(ctx, request)
	if err == nil {
		klog.InfoS("Authorization check completed", "requestID", requestID, "allowed", response.Allowed, "reason", response.Reason)
		_ = s.rbacClient.SetResultInCache(request, response.Allowed, store)
	} else {
		code := http.StatusInternalServerError
		if v, ok := err.(errutils.HttpStatusCode); ok {
			code = v.Code()
		}
		klog.ErrorS(err, "Authorization check failed", "requestID", requestID, "statusCode", code)
		err = errutils.WithCode(errors.Errorf(rbac.CheckAccessErrorFormat, err), code)
	}

	return response, err
}

func getAuthzInfo(environment string) (*rbac.AuthzInfo, error) {
	var err error
	env := azure.PublicCloud
	if environment != "" {
		env, err = azure.EnvironmentFromName(environment)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse environment for azure")
		}
	}

	return &rbac.AuthzInfo{
		AADEndpoint: env.ActiveDirectoryEndpoint,
		ARMEndPoint: env.ResourceManagerEndpoint,
	}, nil
}
