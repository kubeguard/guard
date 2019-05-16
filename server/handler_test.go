package server

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/guard/auth/providers/azure"
	"github.com/appscode/guard/auth/providers/github"
	"github.com/appscode/guard/auth/providers/gitlab"
	"github.com/appscode/guard/auth/providers/google"
	"github.com/appscode/guard/auth/providers/ldap"
	fuzz "github.com/google/gofuzz"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
	auth "k8s.io/api/authentication/v1"
)

func TestServeHTTP(t *testing.T) {
	srv := Server{
		RecommendedOptions: NewRecommendedOptions(),
	}

	store, err := certstore.NewCertStore(afero.NewMemMapFs(), "/pki", "foo")
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
	f := fuzz.New()
	obj := TestData{}

	for i := 0; i < 1000; i++ {
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
		RecommendedOptions: NewRecommendedOptions(),
	}

	// https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-v2-protocols-oidc
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-protocols-oauth-code#jwt-token-claims
	s.RecommendedOptions.Azure.TenantID = "7fe81447-da57-4385-becb-6de57f21477e"

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {

			client, err := s.getAuthProviderClient(test.authProvider, "")

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
