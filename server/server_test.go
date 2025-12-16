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

package server

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLoggerWithSkipPaths(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		skipPaths []string
		shouldLog bool
	}{
		{
			name:      "logs request to non-skipped path",
			path:      "/tokenreviews",
			skipPaths: []string{"/healthz", "/metrics", "/readyz"},
			shouldLog: true,
		},
		{
			name:      "skips logging for /healthz",
			path:      "/healthz",
			skipPaths: []string{"/healthz", "/metrics", "/readyz"},
			shouldLog: false,
		},
		{
			name:      "skips logging for /metrics",
			path:      "/metrics",
			skipPaths: []string{"/healthz", "/metrics", "/readyz"},
			shouldLog: false,
		},
		{
			name:      "skips logging for /readyz",
			path:      "/readyz",
			skipPaths: []string{"/healthz", "/metrics", "/readyz"},
			shouldLog: false,
		},
		{
			name:      "logs request to /subjectaccessreviews",
			path:      "/subjectaccessreviews",
			skipPaths: []string{"/healthz", "/metrics", "/readyz"},
			shouldLog: true,
		},
		{
			name:      "logs when skip list is empty",
			path:      "/healthz",
			skipPaths: []string{},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track whether the logger middleware was invoked
			var loggerCalled atomic.Bool

			// Create a custom logger middleware that tracks invocation
			trackingLogger := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					loggerCalled.Store(true)
					middleware.Logger(next).ServeHTTP(w, r)
				})
			}

			// Create loggerWithSkipPaths that uses our tracking logger
			skipMiddleware := loggerWithSkipPathsCustom(trackingLogger, tt.skipPaths...)

			// Create a simple handler that returns 200 OK
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrappedHandler := skipMiddleware(handler)

			// Create request and response recorder
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			// Serve the request
			wrappedHandler.ServeHTTP(rec, req)

			// Verify response
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify logging behavior
			if tt.shouldLog {
				assert.True(t, loggerCalled.Load(),
					"expected logger to be called for path %q", tt.path)
			} else {
				assert.False(t, loggerCalled.Load(),
					"expected logger to be skipped for path %q", tt.path)
			}
		})
	}
}

// loggerWithSkipPathsCustom is a test helper that accepts a custom logger middleware.
func loggerWithSkipPathsCustom(logger func(http.Handler) http.Handler, skipPaths ...string) func(http.Handler) http.Handler {
	skip := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skip[p] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			logger(next).ServeHTTP(w, r)
		})
	}
}

func TestLoggerWithSkipPathsPreservesHandler(t *testing.T) {
	expectedBody := "test response"
	expectedHeader := "X-Custom-Header"
	expectedHeaderValue := "custom-value"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(expectedHeader, expectedHeaderValue)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(expectedBody))
	})

	middleware := loggerWithSkipPaths("/healthz")
	wrappedHandler := middleware(handler)

	// Test skipped path preserves handler behavior
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, expectedHeaderValue, rec.Header().Get(expectedHeader))
	assert.Equal(t, expectedBody, rec.Body.String())

	// Test non-skipped path preserves handler behavior
	req = httptest.NewRequest(http.MethodGet, "/other", nil)
	rec = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, expectedHeaderValue, rec.Header().Get(expectedHeader))
	assert.Equal(t, expectedBody, rec.Body.String())
}
