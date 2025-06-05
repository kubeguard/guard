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

package rbac

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"go.kubeguard.dev/guard/auth/providers/azure/graph"
	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/stretchr/testify/assert"
	authzv1 "k8s.io/api/authorization/v1"
)

func adaptCheckAccessHTTPClientWithTestTLSConfig(
	checkAccessHTTPClient *http.Client,
	testServer *httptest.Server,
) *http.Client {
	testTransport := testServer.Client().Transport
	if tt, ok := testTransport.(*http.Transport); ok {
		if httpClientTransport, ok := checkAccessHTTPClient.Transport.(*http.Transport); ok {
			// Copy the TLS config from the test server to the checkAccess HTTP client
			httpClientTransport.TLSClientConfig = tt.TLSClientConfig
			checkAccessHTTPClient.Transport = httpClientTransport
		}
	}

	return checkAccessHTTPClient
}

func getAPIServerAndAccessInfo(returnCode int, body, clusterType, resourceId string) (*httptest.Server, *AccessInfo) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(returnCode)
		_, _ = w.Write([]byte(body))
	}))

	apiURL, _ := url.Parse(ts.URL)
	u := &AccessInfo{
		client:          adaptCheckAccessHTTPClientWithTestTLSConfig(httpclient.DefaultHTTPClient, ts),
		apiURL:          apiURL,
		headers:         http.Header{},
		expiresAt:       time.Now().Add(time.Hour),
		clusterType:     clusterType,
		azureResourceId: resourceId,
		armCallLimit:    0,
		lock:            sync.RWMutex{},
		auditSAR:        true,
	}
	return ts, u
}

// getAPIServerAndAccessInfoWithPaths allows custom status and body for /managedNamespaces/ and /namespaces/ paths
func getAPIServerAndAccessInfoWithPaths(
	defaultStatus int, defaultBody, clusterType, resourceId string,
	managedNamespacesStatus int, managedNamespacesBody string,
	namespacesStatus int, namespacesBody string,
) (*httptest.Server, *AccessInfo) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/managedNamespaces/") {
			w.WriteHeader(managedNamespacesStatus)
			_, _ = w.Write([]byte(managedNamespacesBody))
			return
		} else if strings.Contains(r.URL.Path, "/namespaces/") {
			w.WriteHeader(namespacesStatus)
			_, _ = w.Write([]byte(namespacesBody))
			return
		}
		w.WriteHeader(defaultStatus)
		_, _ = w.Write([]byte(defaultBody))
	}))
	apiURL, _ := url.Parse(ts.URL)
	u := &AccessInfo{
		client:          adaptCheckAccessHTTPClientWithTestTLSConfig(httpclient.DefaultHTTPClient, ts),
		apiURL:          apiURL,
		headers:         http.Header{},
		expiresAt:       time.Now().Add(time.Hour),
		clusterType:     clusterType,
		azureResourceId: resourceId,
		armCallLimit:    0,
		lock:            sync.RWMutex{},
		auditSAR:        true,
	}
	return ts, u
}

func TestCheckAccess(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		validBody := `[{"accessDecision":"Allowed",
		"actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete",
		"isDataAction":true,"roleAssignment":null,"denyAssignment":null,"timeToLiveInMs":300000}]`

		ts, u := getAPIServerAndAccessInfo(http.StatusOK, validBody, "arc", "resourceid")
		defer ts.Close()

		request := &authzv1.SubjectAccessReviewSpec{
			User: "alpha@bing.com",
			ResourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "dev", Group: "", Resource: "pods",
				Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
			}, Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
		}

		response, err := u.CheckAccess(request)

		assert.Nilf(t, err, "Should not have got error")
		assert.NotNil(t, response)
		assert.Equal(t, response.Allowed, true)
		assert.Equal(t, response.Denied, false)
	})

	t.Run("too many requests", func(t *testing.T) {
		validBody := `""`

		ts, u := getAPIServerAndAccessInfo(http.StatusTooManyRequests, validBody, "arc", "resourceid")
		defer ts.Close()

		request := &authzv1.SubjectAccessReviewSpec{
			User: "alpha@bing.com",
			ResourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "dev", Group: "", Resource: "pods",
				Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
			}, Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
		}

		response, err := u.CheckAccess(request)

		assert.Nilf(t, response, "response should be nil")
		assert.NotNilf(t, err, "should get error")
	})

	t.Run("check acess not available", func(t *testing.T) {
		validBody := `""`

		ts, u := getAPIServerAndAccessInfo(http.StatusInternalServerError, validBody,
			"arc", "resourceid")
		defer ts.Close()

		request := &authzv1.SubjectAccessReviewSpec{
			User: "alpha@bing.com",
			ResourceAttributes: &authzv1.ResourceAttributes{
				Namespace: "dev", Group: "", Resource: "pods",
				Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
			}, Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
		}

		response, err := u.CheckAccess(request)

		assert.Nilf(t, response, "response should be nil")
		assert.NotNilf(t, err, "should get error")
	})

	t.Run("concurrent access to CheckAccess method", func(t *testing.T) {
		validBody := `[{"accessDecision":"Allowed",
		"actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete",
		"isDataAction":true,"roleAssignment":null,"denyAssignment":null,"timeToLiveInMs":300000}]`

		ts, u := getAPIServerAndAccessInfo(http.StatusOK, validBody, "aks", "aks-managed-cluster")
		defer ts.Close()

		requestTimes := 5
		requests := []*authzv1.SubjectAccessReviewSpec{}
		for i := 0; i < requestTimes; i++ {
			requests = append(
				requests,
				&authzv1.SubjectAccessReviewSpec{
					User: fmt.Sprintf("user%d@bing.com", i),
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "dev", Group: "", Resource: "pods",
						Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
					}, Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
				},
			)
		}

		wg := new(sync.WaitGroup)
		for _, request := range requests {
			wg.Add(1)
			go func(request *authzv1.SubjectAccessReviewSpec) {
				defer wg.Done()
				response, err := u.CheckAccess(request)
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.True(t, response.Allowed)
				assert.False(t, response.Denied)
			}(request)
		}

		wg.Wait()
	})

	t.Run("differing responses for managedNamespaces and namespaces", func(t *testing.T) {
		t.Parallel()
		type testCase struct {
			name                       string
			managedStatus              int
			managedBody                string
			namespaceStatus            int
			namespaceBody              string
			expectedAllowed            bool
			expectedDenied             bool
			enableManagedNamespaceRBAC bool
			clusterType                string
		}

		tests := []testCase{
			{
				name:                       "allowed for managedNamespaces, denied for namespaces",
				managedStatus:              http.StatusOK,
				managedBody:                `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus:            http.StatusOK,
				namespaceBody:              `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            true,
				expectedDenied:             false,
				enableManagedNamespaceRBAC: true,
				clusterType:                managedClusters,
			},
			{
				name:                       "denied for managedNamespaces, allowed for namespaces",
				managedStatus:              http.StatusOK,
				managedBody:                `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus:            http.StatusOK,
				namespaceBody:              `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            true,
				expectedDenied:             false,
				enableManagedNamespaceRBAC: true,
				clusterType:                managedClusters,
			},
			{
				name:                       "denied for both",
				managedStatus:              http.StatusOK,
				managedBody:                `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus:            http.StatusOK,
				namespaceBody:              `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            false,
				expectedDenied:             true,
				enableManagedNamespaceRBAC: true,
				clusterType:                managedClusters,
			},
			{
				name:                       "allowed for both",
				managedStatus:              http.StatusOK,
				managedBody:                `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus:            http.StatusOK,
				namespaceBody:              `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            true,
				expectedDenied:             false,
				enableManagedNamespaceRBAC: true,
				clusterType:                managedClusters,
			},
			{
				name:                       "allowed for both flag off",
				managedStatus:              http.StatusOK,
				managedBody:                `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus:            http.StatusOK,
				namespaceBody:              `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            true,
				expectedDenied:             false,
				enableManagedNamespaceRBAC: false,
				clusterType:                managedClusters,
			},
			{
				name:                       "denied for both flag off",
				managedStatus:              http.StatusOK,
				managedBody:                `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus:            http.StatusOK,
				namespaceBody:              `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            false,
				expectedDenied:             true,
				enableManagedNamespaceRBAC: false,
				clusterType:                managedClusters,
			},
			{
				name:            "allowed for managedNamespaces, denied for namespaces flag off",
				managedStatus:   http.StatusOK,
				managedBody:     `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus: http.StatusOK,

				namespaceBody:              `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            false,
				expectedDenied:             true,
				enableManagedNamespaceRBAC: false,
				clusterType:                managedClusters,
			},
			{
				name:                       "denied for managedNamespaces, allowed for namespaces flag off",
				managedStatus:              http.StatusOK,
				managedBody:                `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus:            http.StatusOK,
				namespaceBody:              `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            true,
				expectedDenied:             false,
				enableManagedNamespaceRBAC: false,
				clusterType:                managedClusters,
			},
			{
				name:                       "allowed for managedNamespaces, denied for namespaces non-mc",
				managedStatus:              http.StatusOK,
				managedBody:                `[{"accessDecision":"Allowed","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				namespaceStatus:            http.StatusOK,
				namespaceBody:              `[{"accessDecision":"Denied","actionId":"Microsoft.Kubernetes/connectedClusters/pods/delete","isDataAction":true}]`,
				expectedAllowed:            false,
				expectedDenied:             true,
				enableManagedNamespaceRBAC: true,
				clusterType:                connectedClusters,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ts, u := getAPIServerAndAccessInfoWithPaths(
					http.StatusOK, tc.namespaceBody, "aks", "resourceid",
					tc.managedStatus, tc.managedBody,
					tc.namespaceStatus, tc.namespaceBody,
				)
				u.useManagedNamespaceResourceScopeFormat = tc.enableManagedNamespaceRBAC
				u.clusterType = tc.clusterType
				defer ts.Close()

				request := &authzv1.SubjectAccessReviewSpec{
					User: "test@bing.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "dev", Group: "", Resource: "pods",
						Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
					},
					Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
				}

				response, err := u.CheckAccess(request)
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tc.expectedAllowed, response.Allowed)
				assert.Equal(t, tc.expectedDenied, response.Denied)
			})
		}
	})
}

func TestCheckAccess_ClusterScoped(t *testing.T) {
	tests := []struct {
		name          string
		returnCode    int
		body          string
		expectedAllow bool
		expectedDeny  bool
	}{
		{
			name:          "cluster-scoped empty namespace → allowed",
			returnCode:    http.StatusOK,
			body:          `[{"accessDecision":"Allowed","actionId":"foo","isDataAction":false}]`,
			expectedAllow: true,
			expectedDeny:  false,
		},
		{
			name:          "cluster-scoped empty namespace → denied",
			returnCode:    http.StatusOK,
			body:          `[{"accessDecision":"Denied","actionId":"foo","isDataAction":false}]`,
			expectedAllow: false,
			expectedDeny:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// set up a fake ARM endpoint that always returns tc.body
			ts, u := getAPIServerAndAccessInfo(
				tc.returnCode,
				tc.body,
				managedClusters,
				"resourceid",
			)
			defer ts.Close()

			// build a SAR with Namespace = ""
			request := &authzv1.SubjectAccessReviewSpec{
				User: "cluster-admin@company.com",
				ResourceAttributes: &authzv1.ResourceAttributes{
					Namespace:   "", // empty = cluster-scoped
					Group:       "",
					Resource:    "pods",
					Version:     "v1",
					Verb:        "get",
					Name:        "my-pod",
					Subresource: "",
				},
				Extra: map[string]authzv1.ExtraValue{"oid": {"00000000-0000-0000-0000-000000000000"}},
			}

			u.useManagedNamespaceResourceScopeFormat = true

			resp, err := u.CheckAccess(request)

			assert.NoError(t, err, "CheckAccess should not return error")
			assert.NotNil(t, resp, "response should always be non-nil")
			assert.Equal(t, tc.expectedAllow, resp.Allowed, "Allowed mismatch")
			assert.Equal(t, tc.expectedDeny, resp.Denied, "Denied mismatch")
		})
	}
}

func getAuthServerAndAccessInfo(returnCode int, body, clientID, clientSecret string) (*httptest.Server, *AccessInfo) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(returnCode)
		_, _ = w.Write([]byte(body))
	}))
	u := &AccessInfo{
		client:  httpclient.DefaultHTTPClient,
		headers: http.Header{},
		lock:    sync.RWMutex{},
	}
	u.tokenProvider = graph.NewClientCredentialTokenProvider(clientID, clientSecret, ts.URL, "")
	return ts, u
}

func TestLogin(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		validToken := "blackbriar"
		validBody := `{
							"token_type": "Bearer",
							"access_token": "%s",
							"expires_on": %d
						}`
		expiresOn := time.Now().Add(time.Second * 3599)
		ts, u := getAuthServerAndAccessInfo(http.StatusOK, fmt.Sprintf(validBody, validToken, expiresOn.Unix()), "jason", "bourne")
		defer ts.Close()

		ctx := context.Background()
		err := u.RefreshToken(ctx)
		if err != nil {
			t.Errorf("Error when trying to log in: %s", err)
		}
		if u.headers.Get("Authorization") != fmt.Sprintf("Bearer %s", validToken) {
			t.Errorf("Authorization header should be set. Expected: %q. Got: %q", fmt.Sprintf("Bearer %s", validToken), u.headers.Get("Authorization"))
		}
		if !time.Now().Before(u.expiresAt) {
			t.Errorf("Expiry not set properly. Expected it to be after the current time. Actual: %v", u.expiresAt)
		}

		// Normalize to second precision for comparison
		expectedExpiresOn := expiresOn.Add(-tokenExpiryDelta).Truncate(time.Second)
		actualExpires := u.expiresAt.Truncate(time.Second)
		if !expectedExpiresOn.Equal(actualExpires) {
			t.Errorf("Expiry not set properly. Expected it to be %v equal to expiresOn. Actual: %v", expectedExpiresOn, actualExpires)
		}
	})

	t.Run("unsuccessful login", func(t *testing.T) {
		ts, u := getAuthServerAndAccessInfo(http.StatusUnauthorized, "Unauthorized", "CIA", "treadstone")
		defer ts.Close()

		ctx := context.Background()
		err := u.RefreshToken(ctx)
		assert.NotNilf(t, err, "Should have gotten error")
	})

	t.Run("request error", func(t *testing.T) {
		badURL := "https://127.0.0.1:34567"
		u := &AccessInfo{
			client:  httpclient.DefaultHTTPClient,
			headers: http.Header{},
			lock:    sync.RWMutex{},
		}
		u.tokenProvider = graph.NewClientCredentialTokenProvider("CIA", "outcome", badURL, "")

		ctx := context.Background()
		err := u.RefreshToken(ctx)
		assert.NotNilf(t, err, "Should have gotten error")
	})

	t.Run("bad response body", func(t *testing.T) {
		ts, u := getAuthServerAndAccessInfo(http.StatusOK, "{bad_json", "CIA", "treadstone")
		defer ts.Close()

		ctx := context.Background()
		err := u.RefreshToken(ctx)
		assert.NotNilf(t, err, "Should have gotten error")
	})
}

func Test_auditSARIfNeeded(t *testing.T) {
	createSAR := func(mu ...func(*authzv1.SubjectAccessReviewSpec)) *authzv1.SubjectAccessReviewSpec {
		sar := &authzv1.SubjectAccessReviewSpec{}
		for _, m := range mu {
			m(sar)
		}
		return sar
	}

	cases := []struct {
		name       string
		accessInfo *AccessInfo
		request    *authzv1.SubjectAccessReviewSpec
	}{
		{
			name:       "audit disabled",
			accessInfo: &AccessInfo{auditSAR: false},
			request:    createSAR(),
		},
		{
			name:       "audit enabled, nil request",
			accessInfo: &AccessInfo{auditSAR: true},
			request:    nil,
		},
		{
			name:       "audit enabled, empty request",
			accessInfo: &AccessInfo{auditSAR: true},
			request:    createSAR(),
		},
		{
			name:       "audit enabled, request with ResourceAttributes",
			accessInfo: &AccessInfo{auditSAR: true},
			request: createSAR(func(sar *authzv1.SubjectAccessReviewSpec) {
				sar.ResourceAttributes = &authzv1.ResourceAttributes{
					Namespace:   "test-namespace",
					Group:       "test-group",
					Resource:    "test-resource",
					Version:     "v1",
					Verb:        "get",
					Name:        "test-name",
					Subresource: "status",
				}
			}),
		},
		{
			name:       "audit enabled, request with NonResourceAttributes",
			accessInfo: &AccessInfo{auditSAR: true},
			request: createSAR(func(sar *authzv1.SubjectAccessReviewSpec) {
				sar.NonResourceAttributes = &authzv1.NonResourceAttributes{
					Path: "/api/v1/test",
					Verb: "get",
				}
			}),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.accessInfo.auditSARIfNeeded(c.request)
		})
	}
}
