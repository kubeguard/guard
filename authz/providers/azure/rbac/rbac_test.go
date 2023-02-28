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
	"sync"
	"testing"
	"time"

	"go.kubeguard.dev/guard/auth/providers/azure/graph"
	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/stretchr/testify/assert"
	authzv1 "k8s.io/api/authorization/v1"
)

func getAPIServerAndAccessInfo(returnCode int, body, clusterType, resourceId string) (*httptest.Server, *AccessInfo) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(returnCode)
		_, _ = w.Write([]byte(body))
	}))
	apiURL, _ := url.Parse(ts.URL)
	u := &AccessInfo{
		client:          httpclient.DefaultHTTPClient,
		apiURL:          apiURL,
		headers:         http.Header{},
		expiresAt:       time.Now().Add(time.Hour),
		clusterType:     clusterType,
		azureResourceId: resourceId,
		armCallLimit:    0,
		lock:            sync.RWMutex{},
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
							"expires_in": 3599,
							"access_token": "%s"
						}`
		ts, u := getAuthServerAndAccessInfo(http.StatusOK, fmt.Sprintf(validBody, validToken), "jason", "bourne")
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
