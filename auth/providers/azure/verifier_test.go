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
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntraSDKTokenVerifierVerify(t *testing.T) {
	t.Run("calls Validate endpoint and returns claims", func(t *testing.T) {
		var method, path, query, authHeader, acceptHeader, hostHeader string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
			path = r.URL.Path
			query = r.URL.RawQuery
			authHeader = r.Header.Get("Authorization")
			acceptHeader = r.Header.Get("Accept")
			hostHeader = r.Host
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"protocol":"Bearer","token":"entra-token","claims":{"aud":["client_id","extra"],"upn":"nahid","oid":"abc-123d4"}}`))
		}))
		defer srv.Close()

		verifier, err := newEntraSDKTokenVerifier(srv.URL+"/?api-version=1", "client_id", false, httpClientRetryCount)
		if !assert.NoError(t, err) {
			return
		}
		verifiedToken, err := verifier.Verify(context.Background(), "entra-token")
		if assert.NoError(t, err) && assert.NotNil(t, verifiedToken) {
			tokenClaims, err := verifiedToken.Claims()
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, "nahid", tokenClaims["upn"])
			assert.Equal(t, "abc-123d4", tokenClaims["oid"])
			assert.True(t, audienceContains(tokenClaims["aud"], "client_id"))
		}

		assert.Equal(t, http.MethodGet, method)
		assert.Equal(t, "/Validate", path)
		assert.Empty(t, query)
		assert.Equal(t, "Bearer entra-token", authHeader)
		assert.Equal(t, "application/json", acceptHeader)
		assert.Equal(t, "localhost", hostHeader)
	})

	t.Run("normalizes loopback hosts to localhost", func(t *testing.T) {
		var hostHeader string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hostHeader = r.Host
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"protocol":"Bearer","token":"entra-token","claims":{"aud":"client_id","oid":"abc-123d4"}}`))
		}))
		defer srv.Close()

		verifier, err := newEntraSDKTokenVerifier(srv.URL, "client_id", false, httpClientRetryCount)
		if !assert.NoError(t, err) {
			return
		}

		_, err = verifier.Verify(context.Background(), "entra-token")
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(hostHeader, "localhost"))
	})

	t.Run("returns SDK error detail on non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"type":"about:blank","title":"Unauthorized","status":401,"detail":"token rejected"}`))
		}))
		defer srv.Close()

		verifier, err := newEntraSDKTokenVerifier(srv.URL, "client_id", false, httpClientRetryCount)
		assert.NoError(t, err)
		verifiedToken, err := verifier.Verify(context.Background(), "bad-token")
		assert.Nil(t, verifiedToken)
		assert.EqualError(t, err, "token failed validation: token rejected")
	})

	t.Run("errors when claims are missing from SDK response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"protocol":"Bearer","token":"entra-token"}`))
		}))
		defer srv.Close()

		verifier, err := newEntraSDKTokenVerifier(srv.URL, "client_id", false, httpClientRetryCount)
		assert.NoError(t, err)
		verifiedToken, err := verifier.Verify(context.Background(), "entra-token")
		assert.Nil(t, verifiedToken)
		assert.EqualError(t, err, "Entra SDK validate response did not include claims")
	})
}

func TestNewEntraSDKTokenVerifier(t *testing.T) {
	t.Run("returns error when endpoint is empty", func(t *testing.T) {
		verifier, err := newEntraSDKTokenVerifier("", "", false, 0)
		assert.Nil(t, verifier)
		assert.EqualError(t, err, "Entra SDK endpoint is empty")
	})

	t.Run("returns error when endpoint is invalid", func(t *testing.T) {
		verifier, err := newEntraSDKTokenVerifier(string([]byte{0x7f}), "", false, 0)
		assert.Nil(t, verifier)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Entra SDK endpoint")
	})

	t.Run("strips user path query and fragment details from the base URL", func(t *testing.T) {
		verifier, err := newEntraSDKTokenVerifier("http://user:pass@localhost:8080/?api-version=1#fragment", "", false, 0)
		if !assert.NoError(t, err) || !assert.NotNil(t, verifier) {
			return
		}
		assert.Equal(t, "http", verifier.baseURL.Scheme)
		assert.Equal(t, "localhost:8080", verifier.baseURL.Host)
		assert.Nil(t, verifier.baseURL.User)
		assert.Empty(t, verifier.baseURL.Path)
		assert.Empty(t, verifier.baseURL.RawQuery)
		assert.Empty(t, verifier.baseURL.Fragment)
	})

	t.Run("rejects endpoint with a non-root path", func(t *testing.T) {
		verifier, err := newEntraSDKTokenVerifier("http://localhost:8080/Validate", "", false, 0)
		assert.Nil(t, verifier)
		assert.EqualError(t, err, "Entra SDK endpoint must be a base URL")
	})
}

func TestParseEntraSDKError(t *testing.T) {
	t.Run("uses JSON title when detail is missing", func(t *testing.T) {
		err := parseEntraSDKError(http.StatusUnauthorized, []byte(`{"title":"Unauthorized"}`))
		assert.EqualError(t, err, "token failed validation: Unauthorized")
	})

	t.Run("uses trimmed non-JSON response body", func(t *testing.T) {
		err := parseEntraSDKError(http.StatusUnauthorized, []byte("  upstream failure  \n"))
		assert.EqualError(t, err, "token failed validation: upstream failure")
	})

	t.Run("handles empty validation response body", func(t *testing.T) {
		err := parseEntraSDKError(http.StatusUnauthorized, nil)
		assert.EqualError(t, err, "token failed validation")
	})

	t.Run("keeps SDK-specific message for unexpected failures", func(t *testing.T) {
		err := parseEntraSDKError(http.StatusBadGateway, []byte("  upstream failure  \n"))
		assert.EqualError(t, err, "Entra SDK validate request failed with status 502: upstream failure")
	})

	t.Run("handles empty unexpected response body", func(t *testing.T) {
		err := parseEntraSDKError(http.StatusBadGateway, nil)
		assert.EqualError(t, err, "Entra SDK validate request failed with status 502")
	})
}

func TestEntraSDKRequestHost(t *testing.T) {
	t.Run("returns empty when base URL is nil", func(t *testing.T) {
		assert.Empty(t, entraSDKRequestHost(nil))
	})

	t.Run("normalizes loopback ip to localhost", func(t *testing.T) {
		assert.Equal(t, "localhost", entraSDKRequestHost(&url.URL{Host: "127.0.0.1:8080"}))
	})

	t.Run("preserves non-loopback hostname without port", func(t *testing.T) {
		assert.Equal(t, "entra-sdk.example.com", entraSDKRequestHost(&url.URL{Host: "entra-sdk.example.com:8080"}))
	})
}
