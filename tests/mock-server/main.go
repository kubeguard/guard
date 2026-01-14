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

// AuthorizationDecision matches Guard's expected response format.
type AuthorizationDecision struct {
	Decision       string `json:"accessDecision"`
	ActionID       string `json:"actionId"`
	IsDataAction   bool   `json:"isDataAction"`
	TimeToLiveInMs int    `json:"timeToLiveInMs"`
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

	// AKS token endpoint
	mux.HandleFunc("/authz/token", handleTokenRequest)

	// CheckAccess v1 API endpoint (wildcard for different resource paths)
	mux.HandleFunc("/", handleCheckAccess)

	// Metrics endpoint
	mux.HandleFunc("/mock-metrics", handleMockMetrics)

	// Health endpoint
	mux.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Mock Azure server starting on %s", addr)
	log.Printf("Configuration: latency=%d-%dms, allow_rate=%.2f, throttle_rate=%.2f",
		config.MinLatencyMS, config.MaxLatencyMS, config.AllowRate, config.ThrottleRate)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	var err error
	if config.UseTLS {
		err = server.ListenAndServeTLS(config.CertFile, config.KeyFile)
	} else {
		err = server.ListenAndServe()
	}

	if err != nil {
		log.Fatalf("Server failed: %v", err)
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

func handleCheckAccess(w http.ResponseWriter, r *http.Request) {
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
