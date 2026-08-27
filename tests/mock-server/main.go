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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// AuthResponse matches Guard's expected token response format.
type AuthResponse struct {
	TokenType string `json:"token_type"`
	Token     string `json:"access_token"`
	ExpiresOn int64  `json:"expires_on"`
}

// CheckAccessRequest represents the incoming authorization request.
type CheckAccessRequest struct {
	Subject struct {
		Attributes struct {
			ObjectID []string `json:"ObjectId"`
		} `json:"Attributes"`
	} `json:"Subject"`
	Actions []struct {
		ID           string `json:"Id"`
		IsDataAction bool   `json:"IsDataAction"`
	} `json:"Actions"`
	Resource struct {
		ID string `json:"Id"`
	} `json:"Resource"`
}

// AuthorizationDecision matches Guard's expected v1 response format.
type AuthorizationDecision struct {
	Decision       string `json:"accessDecision"`
	ActionID       string `json:"actionId"`
	IsDataAction   bool   `json:"isDataAction"`
	TimeToLiveInMs int    `json:"timeToLiveInMs"`
}

// V2 response types matching the checkaccess-v2-go-sdk types.

type V2RoleAssignment struct {
	ID               string `json:"id,omitempty"`
	RoleDefinitionID string `json:"roleDefinitionId,omitempty"`
	PrincipalID      string `json:"principalId,omitempty"`
	PrincipalType    string `json:"principaltype,omitempty"`
	Scope            string `json:"scope,omitempty"`
}

type V2AuthorizationDecision struct {
	ActionID       string           `json:"actionId,omitempty"`
	AccessDecision string           `json:"accessDecision,omitempty"`
	IsDataAction   bool             `json:"isDataAction,omitempty"`
	RoleAssignment V2RoleAssignment `json:"roleAssignment,omitempty"`
	TimeToLiveInMs int              `json:"timeToLiveInMs,omitempty"`
}

type V2AuthorizationDecisionResponse struct {
	Value []V2AuthorizationDecision `json:"value"`
}

type V2AuthorizationRequest struct {
	Subject struct {
		Attributes struct {
			ObjectID  string   `json:"ObjectId"`
			Groups    []string `json:"Groups,omitempty"`
			ClaimName string   `json:"_claim_names,omitempty"`
		} `json:"Attributes"`
	} `json:"Subject"`
	Actions []struct {
		ID string `json:"Id"`
	} `json:"Actions"`
	Resource struct {
		ID string `json:"Id"`
	} `json:"Resource"`
}

// Config holds the mock server configuration.
type Config struct {
	Port           int
	MinLatencyMS   int
	MaxLatencyMS   int
	AllowRate      float64
	ThrottleRate   float64
	CertFile       string
	KeyFile        string
	UseTLS         bool
	VerboseLogging bool
}

var (
	config          Config
	requestCount    atomic.Int64
	throttleCount   atomic.Int64
	allowedCount    atomic.Int64
	deniedCount     atomic.Int64
	tokenIssueCount atomic.Int64
)

func main() {
	flag.IntVar(&config.Port, "port", 8080, "Server port")
	flag.IntVar(&config.MinLatencyMS, "min-latency", 50, "Minimum response latency in ms")
	flag.IntVar(&config.MaxLatencyMS, "max-latency", 200, "Maximum response latency in ms")
	flag.Float64Var(&config.AllowRate, "allow-rate", 0.9, "Rate of allowed responses (0.0-1.0)")
	flag.Float64Var(&config.ThrottleRate, "throttle-rate", 0.01, "Rate of throttled responses (0.0-1.0)")
	flag.StringVar(&config.CertFile, "cert", "", "TLS certificate file")
	flag.StringVar(&config.KeyFile, "key", "", "TLS key file")
	flag.BoolVar(&config.UseTLS, "tls", false, "Enable TLS")
	flag.BoolVar(&config.VerboseLogging, "verbose", false, "Enable verbose logging")
	flag.Parse()

	if config.UseTLS && (config.CertFile == "" || config.KeyFile == "") {
		log.Fatal("TLS enabled but cert or key file not specified")
	}

	mux := http.NewServeMux()

	// AKS OBO authz token endpoint (production pattern: /v1/<ccpid>/authztoken)
	mux.HandleFunc("/v1/", handleOBOToken)

	// OBO token endpoint (legacy path, kept for backward compatibility)
	mux.HandleFunc("/authz/token", handleTokenRequest)

	// CheckAccess v2 PDP endpoint (SDK posts directly to root)
	mux.HandleFunc("/checkaccess/v2", handleCheckAccessV2)

	// CheckAccess v1 API endpoint (wildcard for ARM resource paths)
	mux.HandleFunc("/", handleCheckAccessRouter)

	// Metrics endpoint
	mux.HandleFunc("/mock-metrics", handleMockMetrics)

	// Health endpoint
	mux.HandleFunc("/health", handleHealth)

	log.Printf("Configuration: latency=%d-%dms, allow_rate=%.2f, throttle_rate=%.2f",
		config.MinLatencyMS, config.MaxLatencyMS, config.AllowRate, config.ThrottleRate)

	if config.UseTLS {
		// Dual mode: HTTP on 8080 (OBO tokens) + HTTPS on configured port (PDP)
		go func() {
			httpAddr := ":8080"
			log.Printf("Mock Azure server starting HTTP on %s (OBO token endpoints)", httpAddr)
			httpServer := &http.Server{
				Addr:         httpAddr,
				Handler:      mux,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
			}
			if err := httpServer.ListenAndServe(); err != nil {
				log.Fatalf("HTTP server failed: %v", err)
			}
		}()

		tlsAddr := fmt.Sprintf(":%d", config.Port)
		log.Printf("Mock Azure server starting HTTPS on %s (PDP endpoint)", tlsAddr)
		tlsServer := &http.Server{
			Addr:         tlsAddr,
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
		if err := tlsServer.ListenAndServeTLS(config.CertFile, config.KeyFile); err != nil {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	} else {
		addr := fmt.Sprintf(":%d", config.Port)
		log.Printf("Mock Azure server starting on %s", addr)
		server := &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}

func handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenIssueCount.Add(1)

	// Simulate some latency for token acquisition
	simulateLatency(10, 50)

	response := AuthResponse{
		TokenType: "Bearer",
		Token:     fmt.Sprintf("mock-pdp-token-%d-%d", time.Now().UnixNano(), rand.Int63()),
		ExpiresOn: time.Now().Add(1 * time.Hour).Unix(),
	}

	if config.VerboseLogging {
		log.Printf("[TOKEN] Issued token (total: %d)", tokenIssueCount.Load())
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode token response: %v", err)
	}
}

func handleCheckAccessRouter(w http.ResponseWriter, r *http.Request) {
	// Only handle POST requests to checkaccess endpoint
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify this is a checkaccess request
	if !strings.Contains(r.URL.Path, "/providers/Microsoft.Authorization/checkaccess") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	reqCount := requestCount.Add(1)

	// Simulate realistic Azure API latency
	simulateLatency(config.MinLatencyMS, config.MaxLatencyMS)

	// Random throttling simulation
	if rand.Float64() < config.ThrottleRate {
		throttleCount.Add(1)
		w.Header().Set("Retry-After", "1")
		w.Header().Set("x-ms-ratelimit-remaining-subscription-reads", "0")
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		if config.VerboseLogging {
			log.Printf("[THROTTLE] Request #%d throttled", reqCount)
		}
		return
	}

	// Parse the request
	var checkReq CheckAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&checkReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate decisions for each action
	decisions := make([]AuthorizationDecision, len(checkReq.Actions))
	allAllowed := true

	for i, action := range checkReq.Actions {
		allowed := rand.Float64() < config.AllowRate
		decision := "Denied"
		if allowed {
			decision = "Allowed"
		} else {
			allAllowed = false
		}

		decisions[i] = AuthorizationDecision{
			Decision:       decision,
			ActionID:       action.ID,
			IsDataAction:   action.IsDataAction,
			TimeToLiveInMs: 300000, // 5 minutes
		}
	}

	if allAllowed {
		allowedCount.Add(1)
	} else {
		deniedCount.Add(1)
	}

	// Set rate limit headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-ms-ratelimit-remaining-subscription-reads", "11999")
	w.Header().Set("x-ms-request-id", fmt.Sprintf("mock-%d", reqCount))

	if config.VerboseLogging {
		userID := ""
		if len(checkReq.Subject.Attributes.ObjectID) > 0 {
			userID = checkReq.Subject.Attributes.ObjectID[0]
		}
		log.Printf("[CHECKACCESS] Request #%d: user=%s actions=%d allowed=%v",
			reqCount, userID, len(decisions), allAllowed)
	}

	if err := json.NewEncoder(w).Encode(decisions); err != nil {
		log.Printf("[ERROR] Failed to encode checkaccess response: %v", err)
	}
}

// OBOTokenRequest matches the request format from aksTokenProvider.Acquire()
type OBOTokenRequest struct {
	TenantID    string `json:"tenantID,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
	Resource    string `json:"resource,omitempty"`
}

func handleOBOToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OBOTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokenIssueCount.Add(1)
	simulateLatency(10, 50)

	response := AuthResponse{
		TokenType: "Bearer",
		Token:     fmt.Sprintf("mock-obo-pdp-token-%d-%d", time.Now().UnixNano(), rand.Int63()),
		ExpiresOn: time.Now().Add(1 * time.Hour).Unix(),
	}

	resource := req.Resource
	if resource == "" {
		resource = "https://management.azure.com"
	}

	if config.VerboseLogging {
		log.Printf("[OBO-TOKEN] Issued token for path %s resource=%s (total: %d)", r.URL.Path, resource, tokenIssueCount.Load())
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode OBO token response: %v", err)
	}
}

func handleCheckAccessV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reqCount := requestCount.Add(1)
	simulateLatency(config.MinLatencyMS, config.MaxLatencyMS)

	if rand.Float64() < config.ThrottleRate {
		throttleCount.Add(1)
		w.Header().Set("Retry-After", "1")
		http.Error(w, `{"error":{"code":"TooManyRequests","message":"Rate limit exceeded"}}`, http.StatusTooManyRequests)
		if config.VerboseLogging {
			log.Printf("[V2-THROTTLE] Request #%d throttled", reqCount)
		}
		return
	}

	var req V2AuthorizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"BadRequest","message":"Invalid request body"}}`, http.StatusBadRequest)
		return
	}

	decisions := make([]V2AuthorizationDecision, len(req.Actions))
	allAllowed := true

	for i, action := range req.Actions {
		allowed := rand.Float64() < config.AllowRate

		decision := V2AuthorizationDecision{
			ActionID:       action.ID,
			IsDataAction:   true,
			TimeToLiveInMs: 300000,
		}

		if allowed {
			decision.AccessDecision = "Allowed"
			decision.RoleAssignment = V2RoleAssignment{
				ID:               fmt.Sprintf("/subscriptions/test-sub/providers/Microsoft.Authorization/roleAssignments/%s", generateUUID()),
				RoleDefinitionID: fmt.Sprintf("/subscriptions/test-sub/providers/Microsoft.Authorization/roleDefinitions/%s", generateUUID()),
				PrincipalID:      req.Subject.Attributes.ObjectID,
				PrincipalType:    "User",
				Scope:            req.Resource.ID,
			}
		} else {
			allAllowed = false
			decision.AccessDecision = "NotAllowed"
		}

		decisions[i] = decision
	}

	if allAllowed {
		allowedCount.Add(1)
	} else {
		deniedCount.Add(1)
	}

	resp := V2AuthorizationDecisionResponse{Value: decisions}

	if config.VerboseLogging {
		log.Printf("[V2-CHECKACCESS] Request #%d: user=%s actions=%d resource=%s allowed=%v",
			reqCount, req.Subject.Attributes.ObjectID, len(decisions), req.Resource.ID, allAllowed)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-ms-request-id", fmt.Sprintf("mock-v2-%d", reqCount))
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] Failed to encode v2 checkaccess response: %v", err)
	}
}

func generateUUID() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		rand.Int31(), rand.Int31n(0xffff), rand.Int31n(0xffff),
		rand.Int31n(0xffff), rand.Int63n(0xffffffffffff))
}

func handleMockMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]int64{
		"total_requests":  requestCount.Load(),
		"throttled":       throttleCount.Load(),
		"allowed":         allowedCount.Load(),
		"denied":          deniedCount.Load(),
		"tokens_issued":   tokenIssueCount.Load(),
		"uptime_seconds":  time.Since(startTime).Milliseconds() / 1000,
		"config_port":     int64(config.Port),
		"config_min_lat":  int64(config.MinLatencyMS),
		"config_max_lat":  int64(config.MaxLatencyMS),
		"config_allow":    int64(config.AllowRate * 100),
		"config_throttle": int64(config.ThrottleRate * 100),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Printf("[ERROR] Failed to encode metrics: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("[ERROR] Failed to write health response: %v", err)
	}
}

func simulateLatency(minMS, maxMS int) {
	if minMS <= 0 && maxMS <= 0 {
		return
	}
	latency := minMS
	if maxMS > minMS {
		latency = minMS + rand.Intn(maxMS-minMS)
	}
	time.Sleep(time.Duration(latency) * time.Millisecond)
}

var startTime = time.Now()

func init() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}
