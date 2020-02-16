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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func getAuthServerAndUserInfo(returnCode int, body, clientID, clientSecret string) (*httptest.Server, *UserInfo) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(returnCode)
		_, _ = w.Write([]byte(body))
	}))
	u := &UserInfo{
		client:        http.DefaultClient,
		headers:       http.Header{},
		groupsPerCall: expandedGroupsPerCall,
	}
	u.tokenProvider = NewClientCredentialTokenProvider(clientID, clientSecret, ts.URL, "")
	return ts, u
}

func TestLogin(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		var validToken = "blackbriar"
		var validBody = `{
  "token_type": "Bearer",
  "expires_in": 3599,
  "access_token": "%s"
}`
		ts, u := getAuthServerAndUserInfo(http.StatusOK, fmt.Sprintf(validBody, validToken), "jason", "bourne")
		defer ts.Close()

		err := u.RefreshToken("")
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

		err := u.RefreshToken("")
		if err == nil {
			t.Error("Should have gotten error")
		}
	})

	t.Run("request error", func(t *testing.T) {
		badURL := "https://127.0.0.1:34567"
		u := &UserInfo{
			client:        http.DefaultClient,
			headers:       http.Header{},
			groupsPerCall: expandedGroupsPerCall,
		}
		u.tokenProvider = NewClientCredentialTokenProvider("CIA", "outcome", badURL, "")

		err := u.RefreshToken("")
		if err == nil {
			t.Error("Should have gotten error")
		}
	})

	t.Run("bad response body", func(t *testing.T) {
		ts, u := getAuthServerAndUserInfo(http.StatusOK, "{bad_json", "CIA", "treadstone")
		defer ts.Close()

		err := u.RefreshToken("")
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
		client:        http.DefaultClient,
		apiURL:        apiURL,
		headers:       http.Header{},
		expires:       time.Now().Add(time.Hour),
		groupsPerCall: expandedGroupsPerCall,
	}
	return ts, u
}

func TestGetGroupIDs(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var validBody = `{
  "value": [
      "f36ec2c5-fa5t-4f05-b87f-deadbeef"
  ]
}`
		ts, u := getAPIServerAndUserInfo(http.StatusOK, validBody)
		defer ts.Close()

		groups, err := u.getGroupIDs("john.michael.kane@yacht.io")
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

		groups, err := u.getGroupIDs("alexander.conklin@cia.gov")
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
			client:        http.DefaultClient,
			apiURL:        badURL,
			headers:       http.Header{},
			expires:       time.Now().Add(time.Hour),
			groupsPerCall: expandedGroupsPerCall,
		}

		groups, err := u.getGroupIDs("richard.webb@cia.gov")
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

		groups, err := u.getGroupIDs("nicky.parsons@cia.gov")
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}
	})
}

func TestGetExpandedGroups(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var validBody = `{
  "value": [
    {
      "@odata.type": "#microsoft.graph.group",
      "displayName": "Treadstone.Assets.All"
    }
  ]
}`
		ts, u := getAPIServerAndUserInfo(http.StatusOK, validBody)
		defer ts.Close()

		groups, err := u.getExpandedGroups([]string{"f36ec2c5-fa5t-4f05-b87f-deadbeef"})
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

		groups, err := u.getExpandedGroups([]string{"f36ec2c5-fa5t-4f05-b87f-deadbeef"})
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
			client:        http.DefaultClient,
			apiURL:        badURL,
			headers:       http.Header{},
			expires:       time.Now().Add(time.Hour),
			groupsPerCall: expandedGroupsPerCall,
		}

		groups, err := u.getExpandedGroups([]string{"f36ec2c5-fa5t-4f05-b87f-deadbeef"})
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

		groups, err := u.getExpandedGroups([]string{"f36ec2c5-fa5t-4f05-b87f-deadbeef"})
		if err == nil {
			t.Error("Should have gotten error")
		}
		if groups != nil {
			t.Error("Group list should be nil")
		}
	})
}

// This is only testing the full function run, error cases are handled in the tests above
func TestGetGroups(t *testing.T) {
	var validBody1 = `
{
    "value": [
        "f36ec2c5-fa5t-4f05-b87f-deadbeef"
    ]
}`
	var validBody2 = `{
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
		client:        http.DefaultClient,
		apiURL:        apiURL,
		headers:       http.Header{},
		expires:       time.Now().Add(time.Hour),
		groupsPerCall: expandedGroupsPerCall,
	}
	defer ts.Close()

	groups, err := u.GetGroups("blackbriar@cia.gov")
	if err != nil {
		t.Errorf("Should not have gotten error: %s", err)
	}
	if len(groups) != 1 {
		t.Errorf("Should have gotten a list of groups with 1 entry. Got: %d", len(groups))
	}

	uWithGroupID := &UserInfo{
		client:        http.DefaultClient,
		apiURL:        apiURL,
		headers:       http.Header{},
		expires:       time.Now().Add(time.Hour),
		groupsPerCall: expandedGroupsPerCall,
		useGroupUID:   true,
	}
	defer ts.Close()

	groups, err = uWithGroupID.GetGroups("blackbriar@cia.gov")
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
	var validBody1 = `
{
    "value": [
		"f36ec2c5-fa5t-4f05-b87f-deadbeef",
		"f26ec2c5-fa5t-4f05-b87f-deadbeef",
		"f16ec2c5-fa5t-4f05-b87f-deadbeef"
    ]
}`
	var validBody2 = `{
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

	var validBody3 = `{
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

		var objectQuery = &ObjectQuery{}
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
		client:        http.DefaultClient,
		apiURL:        apiURL,
		headers:       http.Header{},
		expires:       time.Now().Add(time.Hour),
		groupsPerCall: 2,
	}
	defer ts.Close()

	groups, err := u.GetGroups("blackbriar@cia.gov")
	if err != nil {
		t.Errorf("Should not have gotten error: %s", err)
	}

	if len(groups) != 3 {
		t.Errorf("Should have gotten a list of groups with 3 entries. Got: %d", len(groups))
	}
}
