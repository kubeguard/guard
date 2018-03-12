package server

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/kutil/tools/certstore"
	"github.com/google/gofuzz"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	auth "k8s.io/api/authentication/v1"
	"k8s.io/client-go/util/cert"
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

	pemCertsWithOrg, _, err := store.NewClientCertPair("guard", "foo")
	if err != nil {
		t.Fatal(err)
	}
	clientCertWithOrg, err := cert.ParseCertsPEM(pemCertsWithOrg)
	if err != nil {
		t.Fatal(err)
	}

	pemCertsWithoutOrg, _, err := store.NewClientCertPair("guard")
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
