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

	"github.com/appscode/guard/auth/providers/azure/graph"
	"github.com/appscode/pat"
	"github.com/coreos/go-oidc"
	"github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	username    = "nahid"
	loginResp   = `{ "token_type": "Bearer", "expires_in": 8459, "access_token": "%v"}`
	accessToken = `{ "iss" : "%v", "upn": "nahid", "groups": [ "1", "2", "3"] }`
	emptyUpn    = `{ "iss" : "%v",	"groups": [ "1", "2", "3"] }`
	emptyGroup  = `{	"iss" : "%v", "upn": "nahid" }`
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

func newRSAKey(t *testing.T) (*signingKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 1028)
	if err != nil {
		t.Fatal(err)
	}
	return &signingKey{"", priv, priv.Public(), jose.RS256}, nil
}

func clientSetup(clientID, clientSecret, tenantID, serverUrl string, useGroupUID bool) (*Authenticator, error) {
	c := &Authenticator{
		Options: Options{clientID, clientSecret, tenantID, useGroupUID},
		ctx:     context.Background(),
	}

	p, err := oidc.NewProvider(c.ctx, serverUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider for azure. Reason: %v", err)
	}

	c.verifier = p.Verifier(&oidc.Config{
		SkipClientIDCheck: true,
		SkipExpiryCheck:   true,
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
		w.Write([]byte(loginResp))
	}))

	m.Post("/api/users/nahid/getMemberGroups", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(groupStatus) > 0 {
			w.WriteHeader(groupStatus[0])
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write(groupIds)
	}))

	m.Post("/api/directoryObjects/getByIds", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(groupList)
	}))

	m.Get("/.well-known/openid-configuration", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := `{"issuer" : "http://%v", "jwks_uri" : "http://%v/jwk"}`
		w.Write([]byte(fmt.Sprintf(resp, addr, addr)))
	}))

	m.Get("/jwk", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(jwkResp)
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
	gId, err := json.Marshal(groupId)
	if err != nil {
		t.Fatalf("Error when generating groups id and group List. reason %v", err)
	}
	gList, err := json.Marshal(groupList)
	if err != nil {
		t.Fatalf("Error when generating groups id and group List. reason %v", err)
	}
	return gId, gList
}

func assertUserInfo(t *testing.T, info *authv1.UserInfo, groupSize int, useGroupUID bool) {
	if info.Username != username {
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

func getServerAndClient(t *testing.T, signKey *signingKey, loginResp string, groupSize int, useGroupUID bool, groupStatus ...int) (*httptest.Server, *Authenticator) {
	jwkSet := signKey.jwk()
	jwkResp, err := json.Marshal(jwkSet)
	if err != nil {
		t.Fatalf("Error when generating JSONWebKeySet. reason: %v", err)
	}

	// groupSize := sz
	groupIds, groupList := getGroupsAndIds(t, groupSize)

	srv, err := serverSetup(loginResp, http.StatusOK, jwkResp, groupIds, groupList, groupStatus...)
	if err != nil {
		t.Fatalf("Error when creating server, reason: %v", err)
	}

	client, err := clientSetup("client_id", "client_secret", "tenant_id", srv.URL, useGroupUID)
	if err != nil {
		t.Fatalf("Error when creatidng azure client. reason : %v", err)
	}
	return srv, client
}

func TestCheckAzureAuthenticationSuccess(t *testing.T) {
	signKey, err := newRSAKey(t)
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
		{5, emptyGroup},
	}

	for _, test := range dataset {
		// authenticated : true
		t.Run(fmt.Sprintf("authentication successful, group size %v", test.groupSize), func(t *testing.T) {

			srv, client := getServerAndClient(t, signKey, loginResp, test.groupSize, false)
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

			srv, client := getServerAndClient(t, signKey, loginResp, test.groupSize, true)
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

func TestCheckAzureAuthenticationFailed(t *testing.T) {
	signKey, err := newRSAKey(t)
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
			"authentication unsuccessful, reason empty username claim",
			emptyUpn,
			nil,
		},
	}

	for _, test := range dataset {
		t.Run(test.testName, func(t *testing.T) {
			t.Log(test)
			srv, client := getServerAndClient(t, signKey, loginResp, 5, false, test.groupRespStatus...)
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
	"bad_upn": 1204,
}

func TestReviewFromClaims(t *testing.T) {
	// valid user claim
	t.Run("valid user claim", func(t *testing.T) {
		var validUserInfo = &authv1.UserInfo{
			Username: username,
		}

		resp, err := testClaims.getUserInfo("upn")
		assert.Nil(t, err)
		assert.Equal(t, validUserInfo, resp)
	})

	// invalid claim should error
	t.Run("invalid claim should error", func(t *testing.T) {
		resp, err := testClaims.getUserInfo("bad_upn")
		assert.NotNil(t, err)
		assert.Nil(t, resp)
	})
}

func TestString(t *testing.T) {
	// valid claim key should return value
	t.Run("valid claim key should return value", func(t *testing.T) {
		v, err := testClaims.String("upn")
		assert.Nil(t, err)
		assert.Equal(t, username, v)
	})

	// non-existent claim key should error
	t.Run("non-existent claim key should error", func(t *testing.T) {
		v, err := testClaims.String("claim_don't_exist")
		assert.NotNil(t, err)
		assert.Empty(t, v, "expected empty")
	})

	//non-string claim should error
	t.Run("non-string claim should error", func(t *testing.T) {
		v, err := testClaims.String("bad_upn")
		assert.NotNil(t, err)
		assert.Empty(t, v, "expected empty")
	})
}
