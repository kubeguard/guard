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
	"fmt"
	"net/http"
	"testing"
)

func TestAKSTokenProvider(t *testing.T) {
	const (
		inputAccessToken    = "inputAccessToken"
		oboAccessToken      = "oboAccessToken"
		tenantID            = "tenantID"
		oboResponse         = `{"token_type":"Bearer","expires_on":1732881796,"access_token":"%s"}`
		expectedContentType = "application/json"
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
			tokenRequest := struct {
				TenantID    string `json:"tenantID,omitempty"`
				AccessToken string `json:"accessToken,omitempty"`
			}{}

			if err := json.NewDecoder(req.Body).Decode(&tokenRequest); err != nil {
				t.Errorf("unable to decode request: %s", err)
			}
			if tokenRequest.TenantID != tenantID {
				t.Errorf("expected tenant ID: %s, actual: %s", tenantID, tokenRequest.TenantID)
			}
			if tokenRequest.AccessToken != inputAccessToken {
				t.Errorf("expected access token: %s, actual: %s", inputAccessToken, tokenRequest.AccessToken)
			}
			_, _ = rw.Write([]byte(fmt.Sprintf(oboResponse, oboAccessToken)))
		})

		defer stopTestServer(t, s)

		ctx := context.Background()
		r := NewAKSTokenProvider(s.URL, tenantID)
		resp, err := r.Acquire(ctx, inputAccessToken)
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

			tokenRequest := struct {
				TenantID    string `json:"tenantID,omitempty"`
				AccessToken string `json:"accessToken,omitempty"`
			}{}

			if err := json.NewDecoder(req.Body).Decode(&tokenRequest); err != nil {
				t.Errorf("unable to decode request: %s", err)
			}
			if tokenRequest.TenantID != tenantID {
				t.Errorf("expected tenant ID: %s, actual: %s", tenantID, tokenRequest.TenantID)
			}
			if tokenRequest.AccessToken != inputAccessToken {
				t.Errorf("expected access token: %s, actual: %s", inputAccessToken, tokenRequest.AccessToken)
			}
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte(`{"error":{"code":"Authorization_RequestDenied","message":"Insufficient privileges to complete the operation.","innerError":{"request-id":"6e73da71-96a3-4415-8c6a-a940cb1ba052","date":"2019-12-17T21:57:17"}}}`))
		})

		defer stopTestServer(t, s)

		ctx := context.Background()
		r := NewAKSTokenProvider(s.URL, tenantID)
		resp, err := r.Acquire(ctx, inputAccessToken)
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
