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

//nolint:unparam
package azure

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"go.kubeguard.dev/guard/auth/providers/azure/graph"

	"github.com/appscode/pat"
	oidc "github.com/coreos/go-oidc"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	jose "gopkg.in/square/go-jose.v2"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var jsonLib = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	username                       = "nahid"
	objectID                       = "abc-123d4"
	loginResp                      = `{ "token_type": "Bearer", "expires_in": 8459, "access_token": "%v"}`
	accessToken                    = `{ "aud": "client_id", "iss" : "%v", "upn": "nahid", "groups": [ "1", "2", "3"] }`
	accessTokenWithOid             = `{ "aud": "client_id","iss" : "%v", "oid": "abc-123d4", "groups": [ "1", "2", "3"] }`
	accessTokenWithUpnAndOid       = `{ "aud": "client_id","iss" : "%v", "upn": "nahid", "oid": "abc-123d4", "groups": [ "1", "2", "3"] }`
	emptyUpn                       = `{ "aud": "client_id","iss" : "%v",	"groups": [ "1", "2", "3"] }`
	emptyGroup                     = `{	"aud": "client_id","iss" : "%v", "upn": "nahid" }`
	emptyGroupOid                  = `{	"aud": "client_id","iss" : "%v","oid": "abc-123d4" }`
	accessTokenWithOverageClaim    = `{ "aud": "client_id", "iss" : "%v", "upn": "nahid", "_claim_names": {"groups": "src1"}, "_claim_sources": {"src1": {"endpoint": "https://foobar" }} }`
	accessTokenWithNoGroups        = `{ "aud": "client_id", "iss" : "%v", "oid": "abc-123d4" }`
	accessTokenWithoutOverageClaim = `{ "aud": "client_id", "iss" : "%v", "upn": "nahid", "_claim_names": {"foo": "src1"}, "_claim_sources": {"src1": {"endpoint": "https://foobar" }} }`
	badToken                       = "bad_token"
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

func clientSetup(clientID, clientSecret, tenantID, serverUrl string, useGroupUID, verifyClientID bool) (*Authenticator, error) {
	c := &Authenticator{
		Options: Options{
			Environment:    "",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			TenantID:       tenantID,
			UseGroupUID:    useGroupUID,
			AuthMode:       ClientCredentialAuthMode,
			AKSTokenURL:    "",
			VerifyClientID: verifyClientID,
		},
		ctx: context.Background(),
	}

	p, err := oidc.NewProvider(c.ctx, serverUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider for azure. Reason: %v", err)
	}

	c.verifier = p.Verifier(&oidc.Config{
		SkipClientIDCheck: !verifyClientID,
		SkipExpiryCheck:   true,
		ClientID:          clientID,
	})

	c.graphClient, err = graph.TestUserInfo(clientID, clientSecret, serverUrl+"/login", serverUrl+"/api", useGroupUID)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func serverSetup(loginResp string, loginStatus int, jwkResp, groupIds, groupList []byte, groupStatus ...int) (*httptest.Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	addr := listener.Addr().String()

	m := pat.New()

	m.Post("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(loginStatus)
		_, _ = w.Write([]byte(loginResp))
	}))

	m.Post("/api/users/nahid/getMemberGroups", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(groupStatus) > 0 {
			w.WriteHeader(groupStatus[0])
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_, _ = w.Write(groupIds)
	}))

	m.Post("/api/users/abc-123d4/getMemberGroups", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(groupStatus) > 0 {
			w.WriteHeader(groupStatus[0])
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_, _ = w.Write(groupIds)
	}))

	m.Post("/api/directoryObjects/getByIds", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(groupList)
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

/*
goups id format:
{
   "value":[
      "1"
   ]
}

groupList formate:
{
   "value":[
      {
         "displayName":"group1",
         "id":"1"
      }
   ]
}
*/
func getGroupsAndIds(t *testing.T, groupSz int) ([]byte, []byte) {
	groupId := struct {
		Value []string `json:"value"`
	}{}

	type group struct {
		DisplayName string `json:"displayName"`
		Id          string `json:"id"`
	}
	groupList := struct {
		Value []group `json:"value"`
	}{}

	for i := 1; i <= groupSz; i++ {
		th := strconv.Itoa(i)
		groupId.Value = append(groupId.Value, th)
		groupList.Value = append(groupList.Value, group{"group" + th, th})
	}
	gId, err := jsonLib.Marshal(groupId)
	if err != nil {
		t.Fatalf("Error when generating groups id and group List. reason %v", err)
	}
	gList, err := jsonLib.Marshal(groupList)
	if err != nil {
		t.Fatalf("Error when generating groups id and group List. reason %v", err)
	}
	return gId, gList
}

func assertUserInfo(t *testing.T, info *authv1.UserInfo, groupSize int, useGroupUID bool) {
	if info.Username != username && info.Username != objectID {
		t.Errorf("expected username %v, got %v", username, info.Username)
	}

	assert.Equal(t, len(info.Groups), groupSize, "group size should be equal")

	groups := sets.NewString(info.Groups...)
	for i := 1; i <= groupSize; i++ {
		var group string
		if useGroupUID {
			group = strconv.Itoa(i)
		} else {
			group = "group" + strconv.Itoa(i)
		}

		if !groups.Has(group) {
			t.Errorf("group %v is missing", group)
		}
	}
}

func getServerAndClient(t *testing.T, signKey *signingKey, loginResp string, groupSize int, useGroupUID, verifyClientID bool, groupStatus ...int) (*httptest.Server, *Authenticator) {
	jwkSet := signKey.jwk()
	jwkResp, err := jsonLib.Marshal(jwkSet)
	if err != nil {
		t.Fatalf("Error when generating JSONWebKeySet. reason: %v", err)
	}

	// groupSize := sz
	groupIds, groupList := getGroupsAndIds(t, groupSize)

	srv, err := serverSetup(loginResp, http.StatusOK, jwkResp, groupIds, groupList, groupStatus...)
	if err != nil {
		t.Fatalf("Error when creating server, reason: %v", err)
	}

	client, err := clientSetup("client_id", "client_secret", "tenant_id", srv.URL, useGroupUID, verifyClientID)
	if err != nil {
		t.Fatalf("Error when creatidng azure client. reason : %v", err)
	}
	return srv, client
}

func TestCheckAzureAuthenticationSuccess(t *testing.T) {
	signKey, err := newRSAKey()
	if err != nil {
		t.Fatalf("Error when creating signing key. reason : %v", err)
	}

	dataset := []struct {
		groupSize int
		token     string
	}{
		{0, accessToken},
		{1, accessToken},
		{11, accessToken},
		{233, accessToken},
		{0, accessTokenWithOid},
		{1, accessTokenWithOid},
		{15, accessTokenWithOid},
		{0, accessTokenWithUpnAndOid},
		{1, accessTokenWithUpnAndOid},
		{115, accessTokenWithUpnAndOid},
		{5, emptyGroup},
		{5, emptyGroupOid},
	}

	for _, test := range dataset {
		// authenticated : true
		t.Run(fmt.Sprintf("authentication successful, group size %v", test.groupSize), func(t *testing.T) {

			srv, client := getServerAndClient(t, signKey, loginResp, test.groupSize, false, false)
			defer srv.Close()

			token, err := signKey.sign([]byte(fmt.Sprintf(test.token, srv.URL)))
			if err != nil {
				t.Fatalf("Error when signing token. reason: %v", err)
			}

			resp, err := client.Check(token)
			assert.Nil(t, err)
			assertUserInfo(t, resp, test.groupSize, client.UseGroupUID)
		})
	}

	for _, test := range dataset {
		// authenticated : true
		t.Run(fmt.Sprintf("authentication (with group IDs) successful, group size %v", test.groupSize), func(t *testing.T) {

			srv, client := getServerAndClient(t, signKey, loginResp, test.groupSize, true, false)
			defer srv.Close()

			token, err := signKey.sign([]byte(fmt.Sprintf(test.token, srv.URL)))
			if err != nil {
				t.Fatalf("Error when signing token. reason: %v", err)
			}

			resp, err := client.Check(token)
			assert.Nil(t, err)
			assertUserInfo(t, resp, test.groupSize, client.UseGroupUID)
		})
	}
}

func TestCheckAzureAuthenticationWithOverageCheckOption(t *testing.T) {
	signKey, err := newRSAKey()
	if err != nil {
		t.Fatalf("Error when creating signing key. reason : %v", err)
	}

	datasetForSuccess := []struct {
		groupSize int
		token     string
	}{
		{3, accessToken},
		{3, accessTokenWithOverageClaim},
		{0, accessTokenWithNoGroups},
		{0, accessTokenWithoutOverageClaim},
	}
	verifyClientIDs := []bool{true, false}

	for _, verifyClientID := range verifyClientIDs {
		for _, test := range datasetForSuccess {
			t.Run(fmt.Sprintf("authentication successful, verifyClientID: %t, group size %v", verifyClientID, test.groupSize), func(t *testing.T) {

				srv, client := getServerAndClient(t, signKey, loginResp, test.groupSize, true, verifyClientID)
				client.Options.ResolveGroupMembershipOnlyOnOverageClaim = true
				client.Options.UseGroupUID = true
				defer srv.Close()

				token, err := signKey.sign([]byte(fmt.Sprintf(test.token, srv.URL)))
				if err != nil {
					t.Fatalf("Error when signing token. reason: %v", err)
				}

				resp, err := client.Check(token)
				assert.Nil(t, err)
				assertUserInfo(t, resp, test.groupSize, client.UseGroupUID)
			})
		}
	}
}

func TestCheckAzureAuthenticationFailed(t *testing.T) {
	signKey, err := newRSAKey()
	if err != nil {
		t.Fatalf("Error when creating signing key. reason : %v", err)
	}

	dataset := []struct {
		testName        string
		token           string
		groupRespStatus []int
	}{
		{
			"authentication unsuccessful, reason bad token",
			badToken,
			nil,
		},
		{
			"error when getting groups",
			accessToken,
			[]int{http.StatusInternalServerError},
		},
		{
			"error when getting groups",
			accessTokenWithOid,
			[]int{http.StatusInternalServerError},
		},
		{
			"authentication unsuccessful, reason empty username claim",
			emptyUpn,
			nil,
		},
	}

	for _, test := range dataset {
		t.Run(test.testName, func(t *testing.T) {
			t.Log(test)
			srv, client := getServerAndClient(t, signKey, loginResp, 5, false, false, test.groupRespStatus...)
			defer srv.Close()

			var token string
			if test.token != badToken {
				token, err = signKey.sign([]byte(fmt.Sprintf(test.token, srv.URL)))
				if err != nil {
					t.Fatalf("Error when signing token. reason: %v", err)
				}
			} else {
				token = test.token
			}

			resp, err := client.Check(token)
			assert.NotNil(t, err)
			assert.Nil(t, resp)
		})
	}
}

var testClaims = claims{
	"upn":     username,
	"oid":     objectID,
	"bad_upn": 1204,
}

func TestReviewFromClaims(t *testing.T) {
	// valid user claim
	t.Run("valid user claim", func(t *testing.T) {
		var validUserInfo = &authv1.UserInfo{
			Username: username,
			Extra:    map[string]authv1.ExtraValue{"oid": {"abc-123d4"}},
		}

		resp, err := testClaims.getUserInfo("upn", "oid")
		assert.Nil(t, err)
		assert.Equal(t, validUserInfo, resp)
	})

	// invalid claim should error
	t.Run("invalid claim should error", func(t *testing.T) {
		resp, err := testClaims.getUserInfo("bad_upn", "")
		assert.NotNil(t, err)
		assert.Nil(t, resp)
	})
}

func TestString(t *testing.T) {
	// valid claim key should return value
	t.Run("valid claim key should return value", func(t *testing.T) {
		v, err := testClaims.string("upn")
		assert.Nil(t, err)
		assert.Equal(t, username, v)
	})

	// valid claim key should return value
	t.Run("valid claim key should return value", func(t *testing.T) {
		v, err := testClaims.string("oid")
		assert.Nil(t, err)
		assert.Equal(t, objectID, v)
	})

	// non-existent claim key should error
	t.Run("non-existent claim key should error", func(t *testing.T) {
		v, err := testClaims.string("claim_don't_exist")
		assert.NotNil(t, err)
		assert.Empty(t, v, "expected empty")
	})

	//non-string claim should error
	t.Run("non-string claim should error", func(t *testing.T) {
		v, err := testClaims.string("bad_upn")
		assert.NotNil(t, err)
		assert.Empty(t, v, "expected empty")
	})
}

func TestGetAuthInfo(t *testing.T) {
	authInfo, err := getAuthInfo("AzurePublicCloud", "testTenant", localGetMetadata)
	assert.NoError(t, err)
	assert.Contains(t, authInfo.AADEndpoint, "login.microsoftonline.com")

	authInfo, err = getAuthInfo("AzureChinaCloud", "testTenant", localGetMetadata)
	assert.NoError(t, err)
	assert.Contains(t, authInfo.AADEndpoint, "login.chinacloudapi.cn")
}

func localGetMetadata(string, string) (*metadataJSON, error) {
	return &metadataJSON{
		Issuer:      "testIssuer",
		MsgraphHost: "testHost",
	}, nil
}
