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

// token-proxy acts as an OBO replacement for testing Guard against real PDP.
// It receives Guard's OBO-style token requests and fulfills them by calling
// Azure IMDS to get real PDP-audience tokens via a managed identity.
//
// Production flow:  Guard -> OBO -> PDP
// Test flow:        Guard -> token-proxy (IMDS) -> PDP
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type config struct {
	Port     int
	ClientID string
}

type oboRequest struct {
	TenantID    string `json:"tenantID,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
	Resource    string `json:"resource,omitempty"`
}

type oboResponse struct {
	TokenType string `json:"token_type"`
	Token     string `json:"access_token"`
	ExpiresOn int64  `json:"expires_on"`
}

type imdsTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresOn   string `json:"expires_on"`
	Resource    string `json:"resource"`
	TokenType   string `json:"token_type"`
}

var (
	cfg          config
	tokensIssued atomic.Int64
)

func handleOBOToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %v", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req oboRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("[ERROR] Failed to parse request: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resource := req.Resource
	if resource == "" {
		resource = "https://management.azure.com"
	}

	log.Printf("[OBO] Received token request: path=%s tenantID=%s resource=%s", r.URL.Path, req.TenantID, resource)

	imdsToken, err := getIMDSToken(resource, cfg.ClientID)
	if err != nil {
		log.Printf("[ERROR] IMDS token acquisition failed: resource=%s err=%v", resource, err)
		http.Error(w, fmt.Sprintf("IMDS token error: %v", err), http.StatusInternalServerError)
		return
	}

	expiresOn, err := strconv.ParseInt(imdsToken.ExpiresOn, 10, 64)
	if err != nil {
		log.Printf("[ERROR] Failed to parse expires_on=%s: %v", imdsToken.ExpiresOn, err)
		http.Error(w, "invalid expiry from IMDS", http.StatusInternalServerError)
		return
	}

	resp := oboResponse{
		TokenType: "Bearer",
		Token:     imdsToken.AccessToken,
		ExpiresOn: expiresOn,
	}

	count := tokensIssued.Add(1)
	log.Printf("[OBO] Token issued: path=%s resource=%s expires=%s total=%d",
		r.URL.Path, resource, time.Unix(expiresOn, 0).Format(time.RFC3339), count)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
	}
}

func getIMDSToken(resource, clientID string) (*imdsTokenResponse, error) {
	imdsURL := fmt.Sprintf(
		"http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=%s&client_id=%s",
		resource, clientID,
	)

	req, err := http.NewRequest(http.MethodGet, imdsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create IMDS request: %w", err)
	}
	req.Header.Set("Metadata", "true")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("IMDS request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("IMDS returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp imdsTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode IMDS response: %w", err)
	}

	return &tokenResp, nil
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"tokens_issued": tokensIssued.Load(),
		"client_id":     cfg.ClientID,
		"port":          cfg.Port,
	})
}

func main() {
	flag.IntVar(&cfg.Port, "port", 8080, "HTTP listen port")
	flag.StringVar(&cfg.ClientID, "client-id", "", "Managed identity client ID for IMDS token requests (required)")
	flag.Parse()

	if cfg.ClientID == "" {
		log.Fatal("--client-id is required (managed identity client ID)")
	}

	mux := http.NewServeMux()

	// OBO-compatible endpoints (Guard sends requests to these)
	mux.HandleFunc("/v1/", handleOBOToken)
	mux.HandleFunc("/authz/token", handleOBOToken)

	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/metrics", handleMetrics)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Token proxy starting on %s (client-id=%s)", addr, cfg.ClientID)
	log.Printf("Routes: /v1/<ccpid>/authztoken (OBO), /authz/token (legacy), /health, /metrics")

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
