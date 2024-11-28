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
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestMSITokenProvider(t *testing.T) {
	const (
		inputAccessToken    = "inputAccessToken"
		msiAccessToken      = "msiAccessToken"
		tokenResponse       = `{"access_token":"%s","expires_in":"86700","refresh_token":"","expires_on":"1732881796","not_before":"1732795096","resource":"https://management.azure.com","token_type":"Bearer"}`
		expectedContentType = "application/json"
		expectedTokenType   = "Bearer"
	)

	t.Run("Upon Success Response", func(t *testing.T) {
		s := startMSITestServer(t, func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodGet {
				t.Errorf("expected http method %s, actual %s", http.MethodGet, req.Method)
			}
			if req.Header.Get("Content-Type") != expectedContentType {
				t.Errorf("expected content type: %s, actual: %s", expectedContentType, req.Header.Get("Content-Type"))
			}
			_, _ = rw.Write([]byte(fmt.Sprintf(tokenResponse, msiAccessToken)))
		})

		defer stopMSITestServer(t, s)

		ctx := context.Background()
		r := NewMSITokenProvider("http://management.azure.com", s.URL)
		resp, err := r.Acquire(ctx, inputAccessToken)
		if err != nil {
			t.Fatalf("refresh should not return error: %s", err)
		}
		t.Logf("Response: %v", resp)

		var expectedResp TokenResponse
		err = json.Unmarshal([]byte(tokenResponse), &expectedResp)
		if err != nil {
			t.Fatalf("failed to unmarshal token response: %v", err)
		}

		if resp.Token != msiAccessToken {
			t.Errorf("returned obo token '%s' doesn't match expected '%s'", resp.Token, msiAccessToken)
		}
		if resp.TokenType != expectedTokenType {
			t.Errorf("expected token type: Bearer, actual: %s", resp.TokenType)
		}

		expectedExpiresOn, _ := strconv.Atoi(expectedResp.ExpiresOn)
		if resp.ExpiresOn != expectedExpiresOn {
			t.Errorf("expected expires on: %s, actual: %d", expectedResp.ExpiresOn, resp.ExpiresOn)
		}
	})

	t.Run("Upon Error Response", func(t *testing.T) {
		s := startMSITestServer(t, func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodGet {
				t.Errorf("expected http method %s, actual %s", http.MethodGet, req.Method)
			}
			if req.Header.Get("Content-Type") != expectedContentType {
				t.Errorf("expected content type: %s, actual: %s", expectedContentType, req.Header.Get("Content-Type"))
			}

			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte(`{"error":{"code":"Authorization_RequestDenied","message":"Insufficient privileges to complete the operation.","innerError":{"request-id":"6e73da70-96f3-4415-8c6a-a940cb1ba0e2","date":"2019-12-17T21:57:17"}}}`))
		})

		defer stopMSITestServer(t, s)

		ctx := context.Background()
		r := NewMSITokenProvider("http://management.azure.com", s.URL)
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

func startMSITestServer(t *testing.T, handler func(rw http.ResponseWriter, req *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(handler))
}

func stopMSITestServer(t *testing.T, server *httptest.Server) {
	t.Helper()
	server.Close()
}
