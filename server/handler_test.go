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

package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/auth/providers/github"
	"go.kubeguard.dev/guard/auth/providers/gitlab"
	"go.kubeguard.dev/guard/auth/providers/google"
	"go.kubeguard.dev/guard/auth/providers/ldap"
	errutils "go.kubeguard.dev/guard/util/error"

	fuzz "github.com/google/gofuzz"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
	auth "k8s.io/api/authentication/v1"
)

func TestServeHTTP(t *testing.T) {
	srv := Server{
		AuthRecommendedOptions: NewAuthRecommendedOptions(),
	}

	store, err := certstore.New(blobfs.NewInMemoryFS(), "/pki", "foo")
	if err != nil {
		t.Fatal(err)
	}
	err = store.InitCA()
	if err != nil {
		t.Fatal(err)
	}

	pemCertsWithOrg, _, err := store.NewClientCertPairBytes(cert.AltNames{DNSNames: []string{"guard"}}, "foo")
	if err != nil {
		t.Fatal(err)
	}
	clientCertWithOrg, err := cert.ParseCertsPEM(pemCertsWithOrg)
	if err != nil {
		t.Fatal(err)
	}

	pemCertsWithoutOrg, _, err := store.NewClientCertPairBytes(cert.AltNames{DNSNames: []string{"guard"}})
	if err != nil {
		t.Fatal(err)
	}
	clientCertWithoutOrg, err := cert.ParseCertsPEM(pemCertsWithoutOrg)
	if err != nil {
		t.Fatal(err)
	}

	type TestData struct {
		TokenReview      auth.TokenReview
		UseClientCert    bool
		IncludeClientOrg bool
	}
	f := fuzz.New().MaxDepth(3)

	for i := 0; i < 1000; i++ {
		obj := TestData{}
		f.Fuzz(&obj)

		review := new(bytes.Buffer)
		err := json.NewEncoder(review).Encode(obj.TokenReview)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "http://guard.test/tokenreviews", review)
		if obj.UseClientCert && obj.IncludeClientOrg {
			req.TLS = &tls.ConnectionState{
				PeerCertificates: clientCertWithOrg,
			}
		} else if obj.UseClientCert && !obj.IncludeClientOrg {
			req.TLS = &tls.ConnectionState{
				PeerCertificates: clientCertWithoutOrg,
			}
		}

		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "unexpected response status code")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "unexpected response content-type")
		err = json.NewDecoder(resp.Body).Decode(&auth.TokenReview{})
		assert.Nil(t, err, "response body must be of kind TokenReview")
	}
}

func TestGetAuthProviderClient(t *testing.T) {
	const invalideAuthProvider = "invalid_auth_provider"

	testData := []struct {
		testName     string
		authProvider string
		expectedErr  error
	}{
		{
			"get github client",
			github.OrgType,
			nil,
		},
		{
			"get google client",
			google.OrgType,
			nil,
		},
		{
			"get gitlab client",
			gitlab.OrgType,
			nil,
		},
		{
			"get azure client",
			azure.OrgType,
			nil,
		},
		{
			"get LDAP client",
			ldap.OrgType,
			nil,
		},
		{
			"unknown auth providername",
			invalideAuthProvider,
			errors.Errorf("Client is using unknown organization %s", invalideAuthProvider),
		},
	}
	s := Server{
		AuthRecommendedOptions: NewAuthRecommendedOptions(),
	}
	ctx := context.Background()

	// https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-v2-protocols-oidc
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-protocols-oauth-code#jwt-token-claims
	s.AuthRecommendedOptions.Azure.TenantID = "7fe81447-da57-4385-becb-6de57f21477e"

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			client, err := s.getAuthProviderClient(ctx, test.authProvider, "")

			if test.expectedErr == nil {
				assert.Nil(t, err, "expected error nil")

				if err == nil {
					assert.Equal(t, test.authProvider, client.UID())
				}
			} else {
				assert.NotNil(t, err)

				if err != nil {
					assert.EqualError(t, err, test.expectedErr.Error())
				}
			}
		})
	}
}

func TestWriteAuthDecisionError(t *testing.T) {
	t.Run("authentication decision error returns HTTP 200 with Status.Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		authErr := fmt.Errorf("group membership overage claim detected for a service principal token")
		write(w, nil, authErr)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "auth decision errors should return HTTP 200 per TokenReview webhook contract")

		var review auth.TokenReview
		err := json.NewDecoder(resp.Body).Decode(&review)
		assert.Nil(t, err)
		assert.False(t, review.Status.Authenticated)
		assert.Contains(t, review.Status.Error, "group membership overage claim")
	})

	t.Run("infrastructure error with WithCode returns its HTTP status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		infraErr := errutils.WithCode(errors.New("Missing client certificate"), http.StatusBadRequest)
		write(w, nil, infraErr)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "infrastructure errors should retain their explicit HTTP status code")

		var review auth.TokenReview
		err := json.NewDecoder(resp.Body).Decode(&review)
		assert.Nil(t, err)
		assert.False(t, review.Status.Authenticated)
		assert.Contains(t, review.Status.Error, "Missing client certificate")
	})
}
