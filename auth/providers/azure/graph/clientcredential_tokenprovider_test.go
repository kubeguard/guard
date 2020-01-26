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
	"testing"
)

func TestClientCredentialTokenProvider(t *testing.T) {
	const (
		inputAccessToken    = "inputAccessToken"
		oboAccessToken      = "oboAccessToken"
		clientID            = "fakeID"
		clientSecret        = "fakeSecret"
		scope               = "https://graph.microsoft.com/.default"
		oboResponse         = `{"token_type":"Bearer","expires_in":3599,"access_token":"%s"}`
		expectedContentType = "application/x-www-form-urlencoded"
		expectedGrantType   = "client_credentials"
		expectedTokneType   = "Bearer"
	)

	t.Run("Upon Success Response", func(t *testing.T) {
		s := startTestServer(t, func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("expected http method %s, actual %s", http.MethodPost, req.Method)
			}
			if req.Header.Get("Content-Type") != expectedContentType {
				t.Errorf("expected content type: %s, actual: %s", expectedContentType, req.Header.Get("Content-Type"))
			}
			if req.FormValue("client_id") != clientID {
				t.Errorf("expected client_id: %s, actual: %s", clientID, req.FormValue("client_id"))
			}
			if req.FormValue("client_secret") != clientSecret {
				t.Errorf("expected client_secret: %s, actual: %s", clientSecret, req.FormValue("client_secret"))
			}
			if req.FormValue("scope") != scope {
				t.Errorf("expected scope: %s, actual: %s", scope, req.FormValue("scope"))
			}
			if req.FormValue("grant_type") != expectedGrantType {
				t.Errorf("expected grant_type: %s, actual: %s", expectedGrantType, req.FormValue("grant_type"))
			}
			_, _ = rw.Write([]byte(fmt.Sprintf(oboResponse, oboAccessToken)))
		})

		defer stopTestServer(t, s)

		r := NewClientCredentialTokenProvider(clientID, clientSecret, s.URL, scope)
		resp, err := r.Acquire(inputAccessToken)
		if err != nil {
			t.Fatalf("refresh should not return error: %s", err)
		}

		if resp.Token != oboAccessToken {
			t.Errorf("returned obo token '%s' doesn't match expected '%s'", resp.Token, oboAccessToken)
		}
		if resp.TokenType != expectedTokneType {
			t.Errorf("expected token type: Bearer, actual: %s", resp.TokenType)
		}
	})

	t.Run("Upon Error Response", func(t *testing.T) {
		s := startTestServer(t, func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("expected http method %s, actual %s", http.MethodPost, req.Method)
			}
			if req.Header.Get("Content-Type") != expectedContentType {
				t.Errorf("expected content type: %s, actual: %s", expectedContentType, req.Header.Get("Content-Type"))
			}
			if req.FormValue("client_id") != clientID {
				t.Errorf("expected client_id: %s, actual: %s", clientID, req.FormValue("client_id"))
			}
			if req.FormValue("client_secret") != clientSecret {
				t.Errorf("expected client_secret: %s, actual: %s", clientSecret, req.FormValue("client_secret"))
			}
			if req.FormValue("scope") != scope {
				t.Errorf("expected scope: %s, actual: %s", scope, req.FormValue("scope"))
			}
			if req.FormValue("grant_type") != expectedGrantType {
				t.Errorf("expected grant_type: %s, actual: %s", expectedGrantType, req.FormValue("grant_type"))
			}

			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte(`{"error":{"code":"Authorization_RequestDenied","message":"Insufficient privileges to complete the operation.","innerError":{"request-id":"6e73da70-96f3-4415-8c6a-a940cb1ba0e2","date":"2019-12-17T21:57:17"}}}`))
		})

		defer stopTestServer(t, s)

		r := NewClientCredentialTokenProvider(clientID, clientSecret, s.URL, scope)
		resp, err := r.Acquire(inputAccessToken)
		if err == nil {
			t.Error("refresh should return error")
		}

		if resp.Token != "" {
			t.Errorf("returned obo token '%s' should be empty", resp.Token)
		}
		if resp.TokenType != "" {
			t.Errorf("expected token type: %s should be empty", resp.TokenType)
		}
	})
}
