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

	auth "github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/authz"
	"github.com/appscode/guard/authz/providers/azure/rbac"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	authzv1 "k8s.io/api/authorization/v1"
)

const (
	OrgType = "azure"
)

func init() {
	authz.SupportedOrgs = append(authz.SupportedOrgs, OrgType)
}

type Authorizer struct {
	rbacClient *rbac.AccessInfo
}

type authzInfo struct {
	AADEndpoint string
	ARMEndPoint string
}

func New(opts Options, authopts auth.Options, dataStore authz.Store) (authz.Interface, error) {
	c := &Authorizer{}

	authzInfoVal, err := getAuthzInfo(authopts.Environment)
	if err != nil {
		return nil, errors.Wrap(err, "Error in getAuthzInfo %s")
	}

	switch opts.AuthzMode {
	case ARCAuthzMode:
		c.rbacClient, err = rbac.New(authopts.ClientID, authopts.ClientSecret, authopts.TenantID, authzInfoVal.AADEndpoint, authzInfoVal.ARMEndPoint, opts.AuthzMode, opts.ResourceId, opts.ARMCallLimit, dataStore, opts.SkipAuthzCheck, opts.AuthzResolveGroupMemberships, opts.SkipAuthzForNonAADUsers)
	case AKSAuthzMode:
		c.rbacClient, err = rbac.NewWithAKS(opts.AKSAuthzURL, authopts.TenantID, authzInfoVal.ARMEndPoint, opts.AuthzMode, opts.ResourceId, opts.ARMCallLimit, dataStore, opts.SkipAuthzCheck, opts.AuthzResolveGroupMemberships, opts.SkipAuthzForNonAADUsers)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to create ms rbac client")
	}
	return c, nil
}

func (s Authorizer) Check(request *authzv1.SubjectAccessReviewSpec) (*authzv1.SubjectAccessReviewStatus, error) {
	if request == nil {
		return nil, errors.New("subject access review is nil")
	}

	// check if user is service account
	if strings.HasPrefix(strings.ToLower(request.User), "system") {
		glog.V(3).Infof("returning no op to service accounts")
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

	if s.rbacClient.SkipAuthzCheck(request) {
		glog.V(3).Infof("user %s is part of skip authz list. returning no op.", request.User)
		return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NoOpinionVerdict}, nil
	}

	exist, result := s.rbacClient.GetResultFromCache(request)
	if exist {
		if result {
			glog.V(3).Infof("cache hit: returning allowed to user")
			return &authzv1.SubjectAccessReviewStatus{Allowed: result, Reason: rbac.AccessAllowedVerdict}, nil
		} else {
			glog.V(3).Infof("cache hit: returning denied to user")
			return &authzv1.SubjectAccessReviewStatus{Allowed: result, Denied: true, Reason: rbac.AccessNotAllowedVerdict}, nil
		}
	}

	if s.rbacClient.IsTokenExpired() {
		_ = s.rbacClient.RefreshToken()
	}
	return s.rbacClient.CheckAccess(request)
}

func getAuthzInfo(environment string) (*authzInfo, error) {
	var err error
	env := azure.PublicCloud
	if environment != "" {
		env, err = azure.EnvironmentFromName(environment)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse environment for azure")
		}
	}

	return &authzInfo{
		AADEndpoint: env.ActiveDirectoryEndpoint,
		ARMEndPoint: env.ResourceManagerEndpoint,
	}, nil
}
