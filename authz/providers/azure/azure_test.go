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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auth "go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/authz"
	"go.kubeguard.dev/guard/authz/providers/azure/data"
	authzOpts "go.kubeguard.dev/guard/authz/providers/azure/options"
	"go.kubeguard.dev/guard/authz/providers/azure/rbac"
	errutils "go.kubeguard.dev/guard/util/error"
	"go.kubeguard.dev/guard/util/httpclient/httpclienttesting"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	authzv1 "k8s.io/api/authorization/v1"
)

const (
	loginResp            = `{ "token_type": "Bearer", "expires_on": 1732881796, "access_token": "%v"}`
	httpClientRetryCount = 2
)

func init() {
	httpclienttesting.HijackDefaultHTTPClientTransportWithSelfSignedTLS()
}

func clientSetup(serverUrl, mode string) (*Authorizer, error) {
	c := &Authorizer{}
	opts := authzOpts.Options{
		AuthzMode:                      mode,
		ResourceId:                     "resourceId",
		ARMCallLimit:                   2000,
		SkipAuthzCheck:                 []string{"alpha, tango, charlie"},
		SkipAuthzForNonAADUsers:        true,
		AllowNonResDiscoveryPathAccess: true,
		// Used by the AKS/Fleet/AIManager OBO token provider.
		AKSAuthzTokenURL: serverUrl + "/login/token",
	}

	authOpts := auth.Options{
		ClientID:             "client_id",
		ClientSecret:         "client_secret",
		TenantID:             "tenant_id",
		HttpClientRetryCount: httpClientRetryCount,
	}

	authzInfo := rbac.AuthzInfo{
		AADEndpoint: serverUrl + "/login/",
		ARMEndPoint: serverUrl + "/arm/",
	}

	c.rbacClient, err = rbac.New(opts, authOpts, &authzInfo)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func serverSetup(loginResp, checkaccessResp string, loginStatus, checkaccessStatus int, sleepFor time.Duration) *httptest.Server {
	m := chi.NewRouter()

	m.Post("/login/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(loginStatus)
		_, _ = w.Write([]byte(loginResp))
	})

	m.Post("/arm/*", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleepFor)
		w.WriteHeader(checkaccessStatus)
		_, _ = w.Write([]byte(checkaccessResp))
	})

	srv := httptest.NewTLSServer(m)

	return srv
}

func getServerAndClient(t *testing.T, loginResp, checkaccessResp string, checkaccessStatus int, sleepFor time.Duration) (*httptest.Server, *Authorizer, authz.Store) {
	return getServerAndClientForMode(t, authzOpts.ARCAuthzMode, loginResp, checkaccessResp, checkaccessStatus, sleepFor)
}

func getServerAndClientForMode(t *testing.T, mode, loginResp, checkaccessResp string, checkaccessStatus int, sleepFor time.Duration) (*httptest.Server, *Authorizer, authz.Store) {
	srv := serverSetup(loginResp, checkaccessResp, http.StatusOK, checkaccessStatus, sleepFor)

	client, err := clientSetup(srv.URL, mode)
	if err != nil {
		t.Fatalf("Error when creatidng azure client. reason : %v", err)
	}

	testOptions := data.Options{
		HardMaxCacheSize:   1,
		Shards:             1,
		LifeWindow:         1 * time.Minute,
		CleanWindow:        1 * time.Minute,
		MaxEntriesInWindow: 10,
		MaxEntrySize:       5,
		Verbose:            false,
	}
	dataStore, _ := data.NewDataStore(testOptions)

	return srv, client, dataStore
}

func TestCheck(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		validBody := `[{"accessDecision":"Allowed",
		"actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete",
		"isDataAction":true,"roleAssignment":null,"denyAssignment":null,"timeToLiveInMs":300000}]`

		srv, client, store := getServerAndClient(t, loginResp, validBody, http.StatusOK, 1*time.Second)
		defer srv.Close()
		defer func() {
			_ = store.Close()
		}()

		request := &authzv1.SubjectAccessReviewSpec{
			User: "beta@bing.com",
			ResourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "dev", Group: "", Resource: "pods",
				Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
			}, Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
		}

		ctx := context.Background()
		resp, err := client.Check(ctx, request, store)
		assert.Nilf(t, err, "Should not have got error")
		assert.NotNil(t, resp)
		assert.Equal(t, resp.Allowed, true)
		assert.Equal(t, resp.Denied, false)
		if v, ok := err.(errutils.HttpStatusCode); ok {
			assert.Equal(t, v.Code(), http.StatusOK)
		}
	})

	t.Run("unsuccessful request", func(t *testing.T) {
		validBody := `""`
		srv, client, store := getServerAndClient(t, loginResp, validBody, http.StatusInternalServerError, 1*time.Second)
		defer srv.Close()
		defer func() {
			_ = store.Close()
		}()

		request := &authzv1.SubjectAccessReviewSpec{
			User: "beta@bing.com",
			ResourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "dev", Group: "", Resource: "pods",
				Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
			}, Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
		}

		ctx := context.Background()
		resp, err := client.Check(ctx, request, store)
		assert.Nilf(t, resp, "response should be nil")
		assert.NotNilf(t, err, "should get error")
		assert.Contains(t, err.Error(), "Authorization check failed")
		assert.Contains(t, err.Error(), fmt.Sprintf("giving up after %d attempt", httpClientRetryCount+1))
		if v, ok := err.(errutils.HttpStatusCode); ok {
			assert.Equal(t, v.Code(), http.StatusInternalServerError)
		}
	})

	t.Run("context timeout request", func(t *testing.T) {
		validBody := `""`
		srv, client, store := getServerAndClient(t, loginResp, validBody, http.StatusInternalServerError, 25*time.Second)
		defer srv.Close()
		defer func() {
			_ = store.Close()
		}()

		request := &authzv1.SubjectAccessReviewSpec{
			User: "beta@bing.com",
			ResourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "dev", Group: "", Resource: "pods",
				Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
			}, Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
		}

		ctx := context.Background()
		resp, err := client.Check(ctx, request, store)
		assert.Nilf(t, resp, "response should be nil")
		assert.NotNilf(t, err, "should get error")
		assert.Contains(t, err.Error(), "context deadline exceeded")
		if v, ok := err.(errutils.HttpStatusCode); ok {
			assert.Equal(t, v.Code(), http.StatusInternalServerError)
		}
	})
}

func TestCheck_FleetHubClusterReservedResources(t *testing.T) {
	allowedBody := `[{"accessDecision":"Allowed","actionId":"Microsoft.ContainerService/fleets/pods/write","isDataAction":true}]`

	tests := []struct {
		name               string
		mode               string
		resourceAttributes *authzv1.ResourceAttributes
		expectedAllowed    bool
		expectedDenied     bool
		expectedReason     string
	}{
		{
			name: "write access to a resource in a fleet reserved namespace",
			mode: authzOpts.FleetAuthzMode,
			resourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "fleet-system", Group: "apps", Resource: "deployments", Version: "v1", Verb: "create",
			},
			expectedAllowed: false,
			expectedDenied:  true,
			expectedReason:  rbac.FleetReservedNSAccessNotAllowedVerdict,
		},
		{
			name: "read access to secrets in a fleet reserved namespace",
			mode: authzOpts.FleetAuthzMode,
			resourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "kube-system", Group: "", Resource: "secrets", Version: "v1", Verb: "get",
			},
			expectedAllowed: false,
			expectedDenied:  true,
			expectedReason:  rbac.FleetReservedNSAccessNotAllowedVerdict,
		},
		{
			name: "read access to secrets across all namespaces",
			mode: authzOpts.FleetAuthzMode,
			resourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "", Group: "", Resource: "secrets", Version: "v1", Verb: "list",
			},
			expectedAllowed: false,
			expectedDenied:  true,
			expectedReason:  rbac.FleetReservedNSAccessNotAllowedVerdict,
		},
		{
			name: "read access to a node proxy subresource",
			mode: authzOpts.FleetAuthzMode,
			resourceAttributes: &authzv1.ResourceAttributes{
				Group: "", Resource: "nodes", Subresource: "proxy", Version: "v1", Name: "node-1", Verb: "get",
			},
			expectedAllowed: false,
			expectedDenied:  true,
			expectedReason:  rbac.FleetNodeProxyAccessDeniedVerdict,
		},
		{
			name: "write access to a Fleet built-in validating admission policy",
			mode: authzOpts.FleetAuthzMode,
			resourceAttributes: &authzv1.ResourceAttributes{
				Group: "admissionregistration.k8s.io", Resource: "validatingadmissionpolicies", Version: "v1", Name: "fleet-builtin-policy-1", Verb: "update",
			},
			expectedAllowed: false,
			expectedDenied:  true,
			expectedReason:  rbac.FleetBuiltInValidatingAdmissionControlAccessDeniedVerdict,
		},
		{
			name: "write access to a resource in a non-reserved namespace",
			mode: authzOpts.FleetAuthzMode,
			resourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "dev", Group: "apps", Resource: "deployments", Version: "v1", Verb: "create",
			},
			expectedAllowed: true,
			expectedDenied:  false,
		},
		{
			name: "read access to a node status subresource",
			mode: authzOpts.FleetAuthzMode,
			resourceAttributes: &authzv1.ResourceAttributes{
				Group: "", Resource: "nodes", Subresource: "status", Version: "v1", Name: "node-1", Verb: "get",
			},
			expectedAllowed: true,
			expectedDenied:  false,
		},
		{
			name: "reserved resources check is skipped for non-fleet cluster types",
			mode: authzOpts.ARCAuthzMode,
			resourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "fleet-system", Group: "apps", Resource: "deployments", Version: "v1", Verb: "create",
			},
			expectedAllowed: true,
			expectedDenied:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, client, store := getServerAndClientForMode(t, tc.mode, loginResp, allowedBody, http.StatusOK, 0)
			defer srv.Close()
			defer func() {
				_ = store.Close()
			}()

			request := &authzv1.SubjectAccessReviewSpec{
				User:               "beta@bing.com",
				ResourceAttributes: tc.resourceAttributes,
				Extra:              map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
			}

			resp, err := client.Check(context.Background(), request, store)
			if err != nil || resp == nil {
				t.Fatalf("Check() = %v, %v; want non-nil status and no error", resp, err)
			}
			assert.Equal(t, tc.expectedAllowed, resp.Allowed, "Allowed mismatch")
			assert.Equal(t, tc.expectedDenied, resp.Denied, "Denied mismatch")
			if tc.expectedReason != "" {
				assert.Equal(t, tc.expectedReason, resp.Reason, "Reason mismatch")
			}
		})
	}
}

// Denied fleet guardrail results must not be written to the cache, so that a later
// request that the guardrail does not cover is still evaluated against Azure.
func TestCheck_FleetHubClusterReservedResourcesNotCached(t *testing.T) {
	allowedBody := `[{"accessDecision":"Allowed","actionId":"Microsoft.ContainerService/fleets/pods/write","isDataAction":true}]`

	srv, client, store := getServerAndClientForMode(t, authzOpts.FleetAuthzMode, loginResp, allowedBody, http.StatusOK, 0)
	defer srv.Close()
	defer func() {
		_ = store.Close()
	}()

	newRequest := func(namespace string) *authzv1.SubjectAccessReviewSpec {
		return &authzv1.SubjectAccessReviewSpec{
			User: "beta@bing.com",
			ResourceAttributes: &authzv1.ResourceAttributes{
				Namespace: namespace, Group: "apps", Resource: "deployments", Version: "v1", Verb: "create",
			},
			Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
		}
	}

	ctx := context.Background()

	denied, err := client.Check(ctx, newRequest("fleet-system"), store)
	if err != nil || denied == nil {
		t.Fatalf("Check() = %v, %v; want non-nil status and no error", denied, err)
	}
	assert.False(t, denied.Allowed)
	assert.True(t, denied.Denied)
	assert.Equal(t, rbac.FleetReservedNSAccessNotAllowedVerdict, denied.Reason)

	allowed, err := client.Check(ctx, newRequest("dev"), store)
	if err != nil || allowed == nil {
		t.Fatalf("Check() = %v, %v; want non-nil status and no error", allowed, err)
	}
	assert.True(t, allowed.Allowed)
	assert.False(t, allowed.Denied)
}
