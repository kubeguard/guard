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

/*
Package azure implements the Azure RBAC authorization provider for Guard.

# Two-Layer Caching Architecture

Authorization decisions are cached at two layers to minimize latency and reduce
load on the Azure CheckAccess API:

  - Layer 1: Guard's internal cache (BigCache)
    Configured via --azure.cache-size-mb and --azure.cache-ttl-minutes flags.
    Default: 50MB cache size, 3 minute TTL.

  - Layer 2: Kubernetes API server webhook cache
    Configured via kube-apiserver flags:
      --authorization-webhook-cache-authorized-ttl (default: 5m)
      --authorization-webhook-cache-unauthorized-ttl (default: 5m)

# How Caching Works

When a SubjectAccessReview request arrives:

 1. Guard checks its internal cache first
 2. On cache miss, Guard calls Azure CheckAccess API
 3. The response is cached in Guard's BigCache
 4. Guard returns SubjectAccessReviewStatus to the API server
 5. The API server caches the response based on the Allowed field

# Error Handling and Caching

The Kubernetes API server only caches HTTP 200 responses. To enable caching of
error scenarios, Guard handles errors as follows:

  - Deterministic errors (4xx from Azure): Returned as HTTP 200 with
    {Allowed: false, Denied: true}. This allows both Guard and the API server
    to cache the denial, preventing repeated failing calls to Azure.

  - Transient errors (5xx, network errors): Returned as HTTP 5xx error.
    These are NOT cached because they may resolve on retry.

# Metrics

Cache behavior can be monitored via Prometheus metrics:
  - guard_azure_authz_cache_hits_total
  - guard_azure_authz_cache_misses_total
  - guard_azure_authz_cache_entries
  - guard_azure_authz_cache_errors_cached_total (4xx errors cached as denied)

# References

  - Kubernetes Authorization Webhook: https://kubernetes.io/docs/reference/access-authn-authz/webhook/
  - Azure RBAC CheckAccess API: https://learn.microsoft.com/en-us/azure/role-based-access-control/
*/
package azure

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	auth "go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/authz"
	"go.kubeguard.dev/guard/authz/providers/azure/data"
	authzOpts "go.kubeguard.dev/guard/authz/providers/azure/options"
	"go.kubeguard.dev/guard/authz/providers/azure/rbac"
	azureutils "go.kubeguard.dev/guard/util/azure"
	errutils "go.kubeguard.dev/guard/util/error"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/google/uuid"
	authzv1 "k8s.io/api/authorization/v1"
	"k8s.io/klog/v2"
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
		return nil, fmt.Errorf("Error in getAuthzInfo: %w", err)
	}

	c.rbacClient, err = rbac.New(opts, authopts, authzInfoVal)
	if err != nil {
		return nil, fmt.Errorf("failed to create ms rbac client: %w", err)
	}

	return c, nil
}

func (s Authorizer) Check(ctx context.Context, request *authzv1.SubjectAccessReviewSpec, store authz.Store) (*authzv1.SubjectAccessReviewStatus, error) {
	requestID := uuid.New().String()

	log := klog.FromContext(ctx).WithValues("requestID", requestID)
	ctx = klog.NewContext(ctx, log)
	ctx = azureutils.WithRequestID(ctx, requestID)

	if request == nil {
		return nil, errutils.WithCode(fmt.Errorf("Authorization request failed (requestID: %s): subject access review is nil", requestID), http.StatusBadRequest)
	}

	// check if user is system accounts
	if strings.HasPrefix(strings.ToLower(request.User), "system:") {
		log.V(10).Info("Returning no op to system accounts")
		return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NoOpinionVerdict}, nil
	}

	if s.rbacClient.SkipAuthzCheck(request) {
		log.V(3).Info("User is part of skip authz list, returning no op", "user", request.User)
		return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NoOpinionVerdict}, nil
	}

	if _, ok := request.Extra["oid"]; !ok {
		if s.rbacClient.ShouldSkipAuthzCheckForNonAADUsers() {
			log.V(5).Info("Non-AAD user, returning no op", "user", request.User)
			return &authzv1.SubjectAccessReviewStatus{Allowed: false, Reason: rbac.NonAADUserNoOpVerdict}, nil
		} else {
			log.Info("Non-AAD user denied", "user", request.User)
			return &authzv1.SubjectAccessReviewStatus{Allowed: false, Denied: true, Reason: rbac.NonAADUserNotAllowedVerdict}, nil
		}
	}

	exist, result := s.rbacClient.GetResultFromCache(ctx, request, store)

	if exist {
		log.V(5).Info("Cache hit", "allowed", result, "resourceAttributes", request.ResourceAttributes)
		if result {
			return &authzv1.SubjectAccessReviewStatus{Allowed: result, Reason: rbac.AccessAllowedVerdict}, nil
		} else {
			return &authzv1.SubjectAccessReviewStatus{Allowed: result, Denied: true, Reason: rbac.AccessNotAllowedVerdict}, nil
		}
	}

	// if set true, webhook will allow access to discovery APIs for authenticated users. If false, access check will be performed on Azure.
	if s.rbacClient.AllowNonResPathDiscoveryAccess(request) {
		log.V(5).Info("Allowing user access for discovery check")
		if err := s.rbacClient.SetResultInCache(ctx, request, true, store); err != nil {
			log.Error(err, "Failed to cache discovery access result")
		}
		return &authzv1.SubjectAccessReviewStatus{Allowed: true, Reason: rbac.AccessAllowedVerdict}, nil
	}

	ctx = azureutils.WithRetryableHttpClient(ctx, s.httpClientRetryCount)

	if s.rbacClient.IsTokenExpired() {
		if err := s.rbacClient.RefreshToken(ctx); err != nil {
			return nil, errutils.WithCode(fmt.Errorf("Failed to refresh token (requestID: %s): %w", requestID, err), http.StatusInternalServerError)
		}
	}

	response, checkErr := s.rbacClient.CheckAccess(ctx, request)
	if checkErr == nil {
		log.Info("Authorization check completed", "allowed", response.Allowed, "reason", response.Reason, "resourceAttributes", request.ResourceAttributes)
		if cacheErr := s.rbacClient.SetResultInCache(ctx, request, response.Allowed, store); cacheErr != nil {
			log.Error(cacheErr, "Failed to cache authorization result")
		}
		return response, nil
	}

	code := http.StatusInternalServerError
	if v, ok := checkErr.(errutils.HttpStatusCode); ok {
		code = v.Code()
	}

	// For deterministic errors (4xx), return a denied response instead of an error.
	// This allows both Guard and API server to cache the result.
	// Transient errors (5xx, network errors) are returned as errors since they may resolve.
	if code >= 400 && code < 500 {
		log.Info("Returning denied for 4xx error", "statusCode", code, "error", checkErr.Error(), "resourceAttributes", request.ResourceAttributes)
		if cacheErr := s.rbacClient.SetResultInCache(ctx, request, false, store); cacheErr != nil {
			log.Error(cacheErr, "Failed to cache error result")
		} else {
			data.IncErrorsCached()
		}
		// Return a denied response (no error) so API server returns HTTP 200 and caches it
		return &authzv1.SubjectAccessReviewStatus{
			Allowed: false,
			Denied:  true,
			Reason:  fmt.Sprintf("%s (requestID: %s, statusCode: %d): %s", rbac.CheckAccessErrorVerdict, requestID, code, checkErr.Error()),
		}, nil
	}

	// For transient errors (5xx), return the error so it's not cached
	return nil, errutils.WithCode(fmt.Errorf("Authorization check failed (requestID: %s, statusCode: %d): %w", requestID, code, checkErr), code)
}

func getAuthzInfo(environment string) (*rbac.AuthzInfo, error) {
	var err error
	env := azure.PublicCloud
	if environment != "" {
		env, err = azure.EnvironmentFromName(environment)
		if err != nil {
			return nil, fmt.Errorf("failed to parse environment for azure: %w", err)
		}
	}

	return &rbac.AuthzInfo{
		AADEndpoint: env.ActiveDirectoryEndpoint,
		ARMEndPoint: env.ResourceManagerEndpoint,
	}, nil
}
