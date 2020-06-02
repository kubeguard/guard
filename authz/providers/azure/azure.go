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
	"strings"
	"sync"

	auth "github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/authz"
	authzOpts "github.com/appscode/guard/authz/providers/azure/options"
	"github.com/appscode/guard/authz/providers/azure/rbac"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	authzv1 "k8s.io/api/authorization/v1"
)

const (
	OrgType = "azure"
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
	rbacClient *rbac.AccessInfo
}

func New(opts authzOpts.Options, authopts auth.Options) (authz.Interface, error) {
	once.Do(func() {
		glog.Info("Creating Azure global authz client")
		client, err = newAuthzClient(opts, authopts)
		if client == nil || err != nil {
			glog.Fatalf("Authz RBAC client creation failed. Error: %s", err)
		}
	})
	return client, err
}

func newAuthzClient(opts authzOpts.Options, authopts auth.Options) (authz.Interface, error) {
	c := &Authorizer{}

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

func (s Authorizer) Check(request *authzv1.SubjectAccessReviewSpec, store authz.Store) (*authzv1.SubjectAccessReviewStatus, error) {
	if request == nil {
		return nil, errors.New("subject access review is nil")
	}

	// check if user is system accounts
	if strings.HasPrefix(strings.ToLower(request.User), "system:") {
		glog.V(3).Infof("returning no op to system accounts")
		return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NoOpinionVerdict}, nil
	}

	if s.rbacClient.SkipAuthzCheck(request) {
		glog.V(3).Infof("user %s is part of skip authz list. returning no op.", request.User)
		return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NoOpinionVerdict}, nil
	}

	if _, ok := request.Extra["oid"]; !ok {
		if s.rbacClient.ShouldSkipAuthzCheckForNonAADUsers() {
			glog.V(3).Infof("Skip RBAC is set for non AAD users. Returning no opinion for user %s. You may observe this for AAD users for 'can-i' requests.", request.User)
			return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NoOpinionVerdict}, nil
		} else {
			glog.V(3).Infof("Skip RBAC for non AAD user is not set. Returning deny access for non AAD user %s. You may observe this for AAD users for 'can-i' requests.", request.User)
			return &authzv1.SubjectAccessReviewStatus{Allowed: false, Denied: true, Reason: rbac.NotAllowedForNonAADUsers}, nil
		}
	}

	exist, result := s.rbacClient.GetResultFromCache(request, store)
	if exist {
		if result {
			glog.V(3).Infof("cache hit: returning allowed to user %s", request.User)
			return &authzv1.SubjectAccessReviewStatus{Allowed: result, Reason: rbac.AccessAllowedVerdict}, nil
		} else {
			glog.V(3).Infof("cache hit: returning denied to user %s", request.User)
			return &authzv1.SubjectAccessReviewStatus{Allowed: result, Denied: true, Reason: rbac.AccessNotAllowedVerdict}, nil
		}
	}

	// if set true, webhook will allow access to discovery APIs for authenticated users. If false, access check will be performed on Azure.
	if s.rbacClient.AllowNonResPathDiscoveryAccess(request) {
		glog.V(3).Infof("Allowing user %s access for discovery check.", request.User)
		_ = s.rbacClient.SetResultInCache(request, true, store)
		return &authzv1.SubjectAccessReviewStatus{Allowed: true, Reason: rbac.AccessAllowedVerdict}, nil
	}

	if s.rbacClient.IsTokenExpired() {
		_ = s.rbacClient.RefreshToken()
	}

	response, err := s.rbacClient.CheckAccess(request)
	if err == nil {
		_ = s.rbacClient.SetResultInCache(request, response.Allowed, store)
	} else {
		_ = s.rbacClient.SetResultInCache(request, false, store)
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
