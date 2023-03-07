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

package google

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/coreos/go-oidc"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	gdir "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
	"gopkg.in/square/go-jose.v2"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	userEmail   = "nahid@domain.com"
	adminEmail  = "admin@domain.com"
	domain      = "domain"
	googleToken = `{ "iss" : "%s", "email" : "%s", "aud" : "%s", "hd" : "%s"}`
	badToken    = "bad_token"
)

type signingKey struct {
	keyID string // optional
	priv  interface{}
	pub   interface{}
	alg   jose.SignatureAlgorithm
}

func (s *signingKey) sign(payload []byte) (string, error) {
	privKey := &jose.JSONWebKey{Key: s.priv, Algorithm: string(s.alg), KeyID: ""}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: s.alg, Key: privKey}, nil)
	if err != nil {
		return "", err
	}
	jws, err := signer.Sign(payload)
	if err != nil {
		return "", err
	}

	data, err := jws.CompactSerialize()
	if err != nil {
		return "", err
	}
	return data, nil
}

// jwk returns the public part of the signing key.
func (s *signingKey) jwk() jose.JSONWebKeySet {
	k := jose.JSONWebKey{Key: s.pub, Use: "sig", Algorithm: string(s.alg), KeyID: s.keyID}
	kset := jose.JSONWebKeySet{}
	kset.Keys = append(kset.Keys, k)
	return kset
}

func newRSAKey() (*signingKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 1028)
	if err != nil {
		return nil, err
	}
	return &signingKey{"", priv, priv.Public(), jose.RS256}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type googleGroupResp func(u *url.URL) (int, []byte)

// group email format: group[groupNo]@domain.com
func googleGetGroupEmail(groupNo int) string {
	return "group" + strconv.Itoa(groupNo) + "@domain.com"
}

func numberOfPage(total, perPage int) int {
	x := float64(total)
	y := float64(perPage)
	return int(math.Ceil(x / y))
}

func googleGetGroups(pageToken string, groupStartNo, groupPerPage int) gdir.Groups {
	groups := gdir.Groups{
		NextPageToken: pageToken,
	}
	for i := 1; i <= groupPerPage; i++ {
		groups.Groups = append(groups.Groups, &gdir.Group{
			Email: googleGetGroupEmail(groupStartNo + i - 1),
		})
	}
	return groups
}

func googleGetGroupResp(totalGroups, groupPerPage, totalPage int) googleGroupResp {
	groupData := map[string]gdir.Groups{}
	for i := 1; i <= totalPage; i++ {
		nextPageToken := strconv.Itoa(i + 1)
		if i == totalPage {
			// empty next page token for last page
			nextPageToken = ""
		}
		startNo := (i-1)*groupPerPage + 1
		groupSizeInPage := min(groupPerPage, totalGroups-startNo+1)

		if i == 1 {
			// for first page, pageToken is empty
			groupData[""] = googleGetGroups(nextPageToken, startNo, groupSizeInPage)
		} else {
			groupData[strconv.Itoa(i)] = googleGetGroups(nextPageToken, startNo, groupSizeInPage)
		}
	}

	return func(u *url.URL) (int, []byte) {
		if pgToken, ok := u.Query()["pageToken"]; ok {
			if len(pgToken) != 1 {
				return http.StatusBadRequest, []byte("invalid query page parameter")
			}
			data, err := json.Marshal(groupData[pgToken[0]])
			if err != nil {
				return http.StatusBadRequest, []byte(err.Error())
			}
			return http.StatusOK, data
		}

		return http.StatusBadRequest, []byte("invalid query page parameter")
	}
}

// groups search url parameter
// domain, pageToken, userKey
//
// domain, userKey must be non-empty
func googleVerifyUrlParams(u *url.URL) error {
	urlParams := u.Query()
	queryParams := []string{"domain", "pageToken", "userKey"}

	for _, key := range queryParams {
		if val, ok := urlParams[key]; ok {
			if len(val) != 1 {
				return fmt.Errorf("invalid query %v parameter", key)
			}
			if key != "pageToken" && len(val[0]) == 0 {
				return fmt.Errorf("invalid query %v parameter value", key)
			}
		}
	}
	return nil
}

func googleClientSetup(serverUrl string) (*Authenticator, error) {
	g := &Authenticator{
		ctx: context.Background(),
		Options: Options{
			AdminEmail:             adminEmail,
			ServiceAccountJsonFile: "sa.json",
		},
	}
	p, err := oidc.NewProvider(g.ctx, serverUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider for azure. Reason: %v", err)
	}

	g.verifier = p.Verifier(&oidc.Config{
		ClientID:        GoogleOauth2ClientID,
		SkipExpiryCheck: true,
	})

	g.service, err = gdir.NewService(context.Background(), option.WithHTTPClient(httpclient.DefaultHTTPClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create google service. Reason: %v", err)
	}
	g.service.BasePath = serverUrl
	g.service.Groups = gdir.NewGroupsService(g.service)

	return g, nil
}

func googleServerSetup(jwkResp []byte, groupResp googleGroupResp) (*httptest.Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	addr := listener.Addr().String()

	m := chi.NewRouter()

	m.Get("/admin/directory/v1/groups", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// have to work on error response
		err := googleVerifyUrlParams(r.URL)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		status, resp := groupResp(r.URL)
		w.WriteHeader(status)
		_, _ = w.Write(resp)
	}))

	m.Get("/.well-known/openid-configuration", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := `{"issuer" : "http://%v", "jwks_uri" : "http://%v/jwk"}`
		_, _ = w.Write([]byte(fmt.Sprintf(resp, addr, addr)))
	}))

	m.Get("/jwk", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwkResp)
	}))

	srv := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: m},
	}
	srv.Start()

	return srv, nil
}

func assertGroups(t *testing.T, groupNames []string, expectedSize int) {
	if len(groupNames) != expectedSize {
		t.Errorf("expected group size: %v, got %v", expectedSize, len(groupNames))
	}

	groups := sets.NewString(groupNames...)
	for i := 1; i <= expectedSize; i++ {
		group := googleGetGroupEmail(i)
		if !groups.Has(group) {
			t.Errorf("group %v is missing", group)
		}
	}
}

func assertUserInfo(t *testing.T, info *authv1.UserInfo, groupSize int) {
	if info.Username != userEmail {
		t.Errorf("expected username %v, got %v", userEmail, info.Username)
	}
	assertGroups(t, info.Groups, groupSize)
}

func TestCheckGoogleAuthenticationSuccess(t *testing.T) {
	signKey, err := newRSAKey()
	if err != nil {
		t.Fatalf("Error when creating signing key. reason : %v", err)
	}

	jwkSet := signKey.jwk()
	jwkResp, err := json.Marshal(jwkSet)
	if err != nil {
		t.Fatalf("Error when generating JSONWebKeySet. reason: %v", err)
	}

	dataset := []struct {
		groupSize    int
		groupPerPage int
	}{
		{0, 5},
		{1, 5},
		{10, 5},
		{13, 5},
	}

	for _, test := range dataset {
		t.Run(fmt.Sprintf("authentication successful, group size: %v, group per page: %v", test.groupSize, test.groupPerPage), func(t *testing.T) {
			srv, err := googleServerSetup(jwkResp, googleGetGroupResp(test.groupSize, test.groupPerPage, numberOfPage(test.groupSize, test.groupPerPage)))
			if err != nil {
				t.Fatalf("Error when creating server, reason: %v", err)
			}
			defer srv.Close()

			client, err := googleClientSetup(srv.URL)
			if err != nil {
				t.Fatalf("Error when creatidng google client. reason : %v", err)
			}

			token, err := signKey.sign([]byte(fmt.Sprintf(googleToken, srv.URL, userEmail, GoogleOauth2ClientID, domain)))
			if err != nil {
				t.Fatalf("Error when signing token. reason: %v", err)
			}
			// set client domain
			client.domainName = domain

			resp, err := client.Check(token)
			assert.Nil(t, err)
			assertUserInfo(t, resp, test.groupSize)
		})
	}
}

func TestCheckGoogleAuthenticationFailed(t *testing.T) {
	var (
		badIssuer        = fmt.Sprintf(`{ "iss":"%s", "email":"%s", "aud":"%s", "hd":"%s"}`, "https://bad", userEmail, GoogleOauth2ClientID, domain)
		emptyHDToken     = fmt.Sprintf(`{ "iss":"_ISSUER_", "email":"%s", "aud":"%s"}`, userEmail, GoogleOauth2ClientID)
		badClientIDToken = fmt.Sprintf(`{ "iss":"_ISSUER_", "email":"%s", "aud":"bad_client_id", "hd":"%s"}`, userEmail, domain)
		badDomainIDToken = fmt.Sprintf(`{ "iss":"_ISSUER_", "email":"%s", "aud":"%s", "hd":"bad_domain"}`, userEmail, GoogleOauth2ClientID)
		goodIDToken      = fmt.Sprintf(`{ "iss":"_ISSUER_", "email":"%s", "aud":"%s", "hd":"%s"}`, userEmail, GoogleOauth2ClientID, domain)
	)

	signKey, err := newRSAKey()
	if err != nil {
		t.Fatalf("Error when creating signing key. reason : %v", err)
	}

	jwkSet := signKey.jwk()
	jwkResp, err := json.Marshal(jwkSet)
	if err != nil {
		t.Fatalf("Error when generating JSONWebKeySet. reason: %v", err)
	}

	groupErrResp := func(u *url.URL) (int, []byte) {
		return http.StatusInternalServerError, []byte(`{"message" : "error"}`)
	}

	dataset := []struct {
		testName  string
		token     string
		groupResp googleGroupResp
	}{
		{
			"authentication unsuccessful, reason invalid issuer",
			badIssuer,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason invalid token",
			badToken,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason bad client ID",
			badClientIDToken,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason empty hd(host domain)",
			emptyHDToken,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason bad domain",
			badDomainIDToken,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason error occurred in Page 1 (first page)",
			goodIDToken,
			groupErrResp,
		},
		{
			"authentication unsuccessful, reason error occurred in Page 2 (last page)",
			goodIDToken,
			groupErrResp,
		},
	}

	for _, test := range dataset {
		t.Run(test.testName, func(t *testing.T) {
			srv, err := googleServerSetup(jwkResp, test.groupResp)
			if err != nil {
				t.Fatalf("Error when creating server, reason: %v", err)
			}
			defer srv.Close()

			client, err := googleClientSetup(srv.URL)
			if err != nil {
				t.Fatalf("Error when creatidng google client. reason : %v", err)
			}

			test.token = strings.Replace(test.token, "_ISSUER_", srv.URL, -1)
			token, err := signKey.sign([]byte(test.token))
			if err != nil {
				t.Fatalf("Error when signing token. reason: %v", err)
			}

			// set client domain
			client.domainName = domain

			resp, err := client.Check(token)
			// t.Log(test)
			assert.NotNil(t, err)
			assert.Nil(t, resp)
		})
	}
}
