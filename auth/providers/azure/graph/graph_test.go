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

package graph

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.kubeguard.dev/guard/util/httpclient"

	"gopkg.in/square/go-jose.v2"
)

const (
	accessTokenWithOverageClaim       = `{ "aud": "client", "iss" : "%v", "exp" : "%v",  "upn": "arc", "_claim_names": {"groups": "src1"}, "_claim_sources": {"src1": {"endpoint": "https://foobar" }} }`
	accessTokenWithOverageClaimForApp = `{ "aud": "client", "iss" : "%v", "exp" : "%v", "idtyp" : "app", "upn": "arc", "_claim_names": {"groups": "src1"}, "_claim_sources": {"src1": {"endpoint": "https://foobar" }} }`
	location                          = "eastus"
)

type swKey struct {
	keyID  string
	pKey   *rsa.PrivateKey
	pubKey interface{}
}

func (swk *swKey) Alg() string {
	return "RS256"
}

func (swk *swKey) KeyID() string {
	return ""
}

func NewSwkKey() (*swKey, error) {
	rsa, err := rsa.GenerateKey(rand.Reader, 1028)
	if err != nil {
		return nil, err
	}
	return &swKey{"", rsa, rsa.Public()}, nil
}

func (swk *swKey) GenerateToken(payload []byte) (string, error) {
	pKey := &jose.JSONWebKey{Key: swk.pKey, Algorithm: swk.Alg(), KeyID: swk.KeyID()}

	// create a Square.jose RSA signer, used to sign the JWT
	signerOpts := jose.SignerOptions{}
	signerOpts.WithType("JWT")
	rsaSigner, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: pKey}, &signerOpts)
	if err != nil {
		return "", err
	}
	jws, err := rsaSigner.Sign(payload)
	if err != nil {
		return "", err
	}

	token, err := jws.CompactSerialize()
	if err != nil {
		return "", err
	}
	return token, nil
}

func getAuthServerAndUserInfo(returnCode int, body, clientID, clientSecret string) (*httptest.Server, *UserInfo) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(returnCode)
		_, _ = w.Write([]byte(body))
	}))
	u := &UserInfo{
		client:        httpclient.DefaultHTTPClient,
		headers:       http.Header{},
		groupsPerCall: expandedGroupsPerCall,
	}
	u.tokenProvider = NewClientCredentialTokenProvider(clientID, clientSecret, ts.URL, "")
	return ts, u
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		validToken := "blackbriar"
		validBody := `{
  "token_type": "Bearer",
  "expires_in": 3599,
  "access_token": "%s"
}`
		ts, u := getAuthServerAndUserInfo(http.StatusOK, fmt.Sprintf(validBody, validToken), "jason", "bourne")
		defer ts.Close()

		err := u.RefreshToken(ctx, "")
		if err != nil {
			t.Errorf("Error when trying to log in: %s", err)
		}
		if u.headers.Get("Authorization") != fmt.Sprintf("Bearer %s", validToken) {
			t.Errorf("Authorization header should be set. Expected: %q. Got: %q", fmt.Sprintf("Bearer %s", validToken), u.headers.Get("Authorization"))
		}
		if !time.Now().Before(u.expires) {
			t.Errorf("Expiry not set properly. Expected it to be after the current time. Actual: %v", u.expires)
		}
	})

	t.Run("unsuccessful login", func(t *testing.T) {
		ts, u := getAuthServerAndUserInfo(http.StatusUnauthorized, "Unauthorized", "CIA", "treadstone")
		defer ts.Close()

		err := u.RefreshToken(ctx, "")
		if err == nil {
			t.Error("Should have gotten error")
		}
	})

	t.Run("request error", func(t *testing.T) {
		badURL := "https://127.0.0.1:34567"
		u := &UserInfo{
			client:        httpclient.DefaultHTTPClient,
			headers:       http.Header{},
			groupsPerCall: expandedGroupsPerCall,
		}
		u.tokenProvider = NewClientCredentialTokenProvider("CIA", "outcome", badURL, "")

		err := u.RefreshToken(ctx, "")
		if err == nil {
			t.Error("Should have gotten error")
		}
	})

	t.Run("bad response body", func(t *testing.T) {
		ts, u := getAuthServerAndUserInfo(http.StatusOK, "{bad_json", "CIA", "treadstone")
		defer ts.Close()

		err := u.RefreshToken(ctx, "")
		if err == nil {
			t.Error("Should have gotten error")
		}
	})
}

func getAPIServerAndUserInfo(returnCode int, body string) (*httptest.Server, *UserInfo) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(returnCode)
		_, _ = w.Write([]byte(body))
	}))
	apiURL, _ := url.Parse(ts.URL)
	u := &UserInfo{
		client:        httpclient.DefaultHTTPClient,
		apiURL:        apiURL,
		headers:       http.Header{},
		expires:       time.Now().Add(time.Hour),
		groupsPerCall: expandedGroupsPerCall,
	}
	return ts, u
}

func TestGetGroupIDs(t *testing.T) {
	ctx := context.Background()

	t.Run("successful request", func(t *testing.T) {
		validBody := `{
  "value": [
      "f36ec2c5-fa5t-4f05-b87f-deadbeef"
  ]
}`
		ts, u := getAPIServerAndUserInfo(http.StatusOK, validBody)
		defer ts.Close()

		groups, err := u.getGroupIDs(ctx, "john.michael.kane@yacht.io")
		if err != nil {
			t.Errorf("Should not have gotten error: %s", err)
		}
		if len(groups) != 1 {
			t.Errorf("Should have gotten a list of group IDs with 1 entry. Got: %d", len(groups))
		}
	})
	t.Run("bad server response", func(t *testing.T) {
		ts, u := getAPIServerAndUserInfo(http.StatusInternalServerError, "shutdown")
		defer ts.Close()

		groups, err := u.getGroupIDs(ctx, "alexander.conklin@cia.gov")
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}
	})
	t.Run("request error", func(t *testing.T) {
		badURL, _ := url.Parse("https://127.0.0.1:34567")
		u := &UserInfo{
			client:        httpclient.DefaultHTTPClient,
			apiURL:        badURL,
			headers:       http.Header{},
			expires:       time.Now().Add(time.Hour),
			groupsPerCall: expandedGroupsPerCall,
		}

		groups, err := u.getGroupIDs(ctx, "richard.webb@cia.gov")
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}
	})
	t.Run("bad response body", func(t *testing.T) {
		ts, u := getAPIServerAndUserInfo(http.StatusOK, "{bad_json")
		defer ts.Close()

		groups, err := u.getGroupIDs(ctx, "nicky.parsons@cia.gov")
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}
	})
}

func TestGetExpandedGroups(t *testing.T) {
	ctx := context.Background()

	t.Run("successful request", func(t *testing.T) {
		validBody := `{
  "value": [
    {
      "@odata.type": "#microsoft.graph.group",
      "displayName": "Treadstone.Assets.All"
    }
  ]
}`
		ts, u := getAPIServerAndUserInfo(http.StatusOK, validBody)
		defer ts.Close()

		groups, err := u.getExpandedGroups(ctx, []string{"f36ec2c5-fa5t-4f05-b87f-deadbeef"})
		if err != nil {
			t.Errorf("Should not have gotten error: %s", err)
		}
		if len(groups.Value) != 1 {
			t.Errorf("Should have gotten a list of groups with 1 entry. Got: %d", len(groups.Value))
		}
	})
	t.Run("bad server response", func(t *testing.T) {
		ts, u := getAPIServerAndUserInfo(http.StatusInternalServerError, "shutdown")
		defer ts.Close()

		groups, err := u.getExpandedGroups(ctx, []string{"f36ec2c5-fa5t-4f05-b87f-deadbeef"})
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}
	})
	t.Run("request error", func(t *testing.T) {
		badURL, _ := url.Parse("https://127.0.0.1:34567")
		u := &UserInfo{
			client:        httpclient.DefaultHTTPClient,
			apiURL:        badURL,
			headers:       http.Header{},
			expires:       time.Now().Add(time.Hour),
			groupsPerCall: expandedGroupsPerCall,
		}

		groups, err := u.getExpandedGroups(ctx, []string{"f36ec2c5-fa5t-4f05-b87f-deadbeef"})
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}
	})
	t.Run("bad response body", func(t *testing.T) {
		ts, u := getAPIServerAndUserInfo(http.StatusOK, "{bad_json")
		defer ts.Close()

		groups, err := u.getExpandedGroups(ctx, []string{"f36ec2c5-fa5t-4f05-b87f-deadbeef"})
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}
	})
}

func TestGetMemberGroupsUsingARCOboService(t *testing.T) {
	ctx := context.Background()
	key, err := NewSwkKey()
	if err != nil {
		t.Fatalf("Failed to generate SF key. Error:%+v", err)
	}

	t.Run("successful request", func(t *testing.T) {
		validBody := `{
			"value": [
				"f36ec2c5-fa5t-4f05-b87f-deadbeef"
			]
		  }`

		ts, u := getAPIServerAndUserInfo(http.StatusOK, validBody)
		defer ts.Close()

		getOBORegionalEndpoint = func(location string, resourceID string) (string, error) {
			return ts.URL, nil
		}

		u.headers.Set("Authorization", "Bearer msitoken")

		tokenstring, err := key.GenerateToken([]byte(fmt.Sprintf(accessTokenWithOverageClaim, ts.URL, time.Now().Add(time.Minute*5).Unix())))
		if err != nil {
			t.Fatalf("Error when generating token. Error:%+v", err)
		}

		groups, err := u.GetMemberGroupsUsingARCOboService(ctx, "tenantId", ts.URL, location, tokenstring)
		if err != nil {
			t.Errorf("Should not have gotten error: %s", err)
		}
		if len(groups) != 1 {
			t.Errorf("Should have gotten a list of groups with 1 entry. Got: %d", len(groups))
		}
	})
	t.Run("bad server response", func(t *testing.T) {
		ts, u := getAPIServerAndUserInfo(http.StatusInternalServerError, "shutdown")
		defer ts.Close()

		getOBORegionalEndpoint = func(location string, resourceID string) (string, error) {
			return ts.URL, nil
		}

		u.headers.Set("Authorization", "Bearer msitoken")

		tokenstring, err := key.GenerateToken([]byte(fmt.Sprintf(accessTokenWithOverageClaim, ts.URL, time.Now().Add(time.Minute*5).Unix())))
		if err != nil {
			t.Fatalf("Error when generating token. Error:%+v", err)
		}

		groups, err := u.GetMemberGroupsUsingARCOboService(ctx, "tenantId", ts.URL, location, tokenstring)
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}

		if !strings.Contains(err.Error(), "Failed to fetch group info with status code: 500") {
			t.Errorf("Expected: Failed to fetch group info with status code: 500, Got: %s", err.Error())
		}
	})
	t.Run("applications not supported error", func(t *testing.T) {
		ts, u := getAPIServerAndUserInfo(http.StatusBadRequest, "")
		defer ts.Close()
		getOBORegionalEndpoint = func(location string, resourceID string) (string, error) {
			return ts.URL, nil
		}

		u.headers.Set("Authorization", "Bearer msitoken")

		tokenstring, err := key.GenerateToken([]byte(fmt.Sprintf(accessTokenWithOverageClaimForApp, ts.URL, time.Now().Add(time.Minute*5).Unix())))
		if err != nil {
			t.Fatalf("Error when generating token. Error:%+v", err)
		}

		groups, err := u.GetMemberGroupsUsingARCOboService(ctx, "tenantId", ts.URL, location, tokenstring)
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}

		if !strings.Contains(err.Error(), "Obo.GetMemberGroups call is not supported for applications.") {
			t.Errorf("Expected: Obo.GetMemberGroups call is not supported for applications., Got: %s", err.Error())
		}
	})
	t.Run("bad response body", func(t *testing.T) {
		ts, u := getAPIServerAndUserInfo(http.StatusOK, "{bad_json")
		defer ts.Close()

		getOBORegionalEndpoint = func(location string, resourceID string) (string, error) {
			return ts.URL, nil
		}

		u.headers.Set("Authorization", "Bearer msitoken")

		tokenstring, err := key.GenerateToken([]byte(fmt.Sprintf(accessTokenWithOverageClaim, ts.URL, time.Now().Add(time.Minute*5).Unix())))
		if err != nil {
			t.Fatalf("Error when generating token. Error:%+v", err)
		}

		groups, err := u.GetMemberGroupsUsingARCOboService(ctx, "tenantId", ts.URL, location, tokenstring)
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}

		if !strings.Contains(err.Error(), "Failed to decode response for request") {
			t.Errorf("Expected: Failed to decode response for request, Got: %s", err.Error())
		}
	})
}

// This is only testing the full function run, error cases are handled in the tests above
func TestGetGroups(t *testing.T) {
	ctx := context.Background()
	validBody1 := `
{
    "value": [
        "f36ec2c5-fa5t-4f05-b87f-deadbeef"
    ]
}`
	validBody2 := `{
	"value": [
		{
		    "@odata.type": "#microsoft.graph.group",
		    "displayName": "Treadstone.Assets.All"
		}
	]
}`
	mux := http.NewServeMux()
	mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{ "token_type": "Bearer", "expires_in": 8459, "access_token": "secret"}`))
	}))
	mux.Handle("/users/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(validBody1))
	}))
	mux.Handle("/directoryObjects/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(validBody2))
	}))
	ts := httptest.NewServer(mux)
	apiURL, _ := url.Parse(ts.URL)

	u := &UserInfo{
		client:        httpclient.DefaultHTTPClient,
		apiURL:        apiURL,
		headers:       http.Header{},
		expires:       time.Now().Add(time.Hour),
		groupsPerCall: expandedGroupsPerCall,
	}
	defer ts.Close()

	groups, err := u.GetGroups(ctx, "blackbriar@cia.gov")
	if err != nil {
		t.Errorf("Should not have gotten error: %s", err)
	}
	if len(groups) != 1 {
		t.Errorf("Should have gotten a list of groups with 1 entry. Got: %d", len(groups))
	}

	uWithGroupID := &UserInfo{
		client:        httpclient.DefaultHTTPClient,
		apiURL:        apiURL,
		headers:       http.Header{},
		expires:       time.Now().Add(time.Hour),
		groupsPerCall: expandedGroupsPerCall,
		useGroupUID:   true,
	}
	defer ts.Close()

	groups, err = uWithGroupID.GetGroups(ctx, "blackbriar@cia.gov")
	if err != nil {
		t.Errorf("Should not have gotten error: %s", err)
	}
	if len(groups) != 1 {
		t.Errorf("Should have gotten a list of groups with 1 entry. Got: %d", len(groups))
	}
	if groups[0] != "f36ec2c5-fa5t-4f05-b87f-deadbeef" {
		t.Errorf("Should have gotten one group ID in the list. Got: %s", groups[0])
	}
}

func TestGetGroupsPaging(t *testing.T) {
	validBody1 := `
{
    "value": [
		"f36ec2c5-fa5t-4f05-b87f-deadbeef",
		"f26ec2c5-fa5t-4f05-b87f-deadbeef",
		"f16ec2c5-fa5t-4f05-b87f-deadbeef"
    ]
}`
	validBody2 := `{
	"value": [
		{
		    "@odata.type": "#microsoft.graph.group",
		    "displayName": "Treadstone.Assets.All"
		},
		{
		    "@odata.type": "#microsoft.graph.group",
		    "displayName": "Treadstone.Assets.Finance"
		}
	]
}`

	validBody3 := `{
	"value": [
		{
		    "@odata.type": "#microsoft.graph.group",
		    "displayName": "Treadstone.Assets.HR"
		}
	]
}`

	mux := http.NewServeMux()
	mux.Handle("/users/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(validBody1))
	}))
	mux.Handle("/directoryObjects/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)

		objectQuery := &ObjectQuery{}
		err := json.NewDecoder(r.Body).Decode(objectQuery)
		if err != nil {
			t.Errorf("Error decoding request body: %s", err)
		}

		if len(objectQuery.IDs) == 2 {
			_, _ = w.Write([]byte(validBody2))
		} else {
			_, _ = w.Write([]byte(validBody3))
		}
	}))
	ts := httptest.NewServer(mux)
	apiURL, _ := url.Parse(ts.URL)

	u := &UserInfo{
		client:        httpclient.DefaultHTTPClient,
		apiURL:        apiURL,
		headers:       http.Header{},
		expires:       time.Now().Add(time.Hour),
		groupsPerCall: 2,
	}
	defer ts.Close()

	ctx := context.Background()
	groups, err := u.GetGroups(ctx, "blackbriar@cia.gov")
	if err != nil {
		t.Errorf("Should not have gotten error: %s", err)
	}

	if len(groups) != 3 {
		t.Errorf("Should have gotten a list of groups with 3 entries. Got: %d", len(groups))
	}
}
