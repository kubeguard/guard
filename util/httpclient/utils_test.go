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

package httpclient

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClientInitialization(t *testing.T) {
	tests := []struct {
		name                    string
		envVars                 map[string]string
		expectedPingEnabled     bool
		expectedReadIdleTimeout time.Duration
		expectedPingTimeout     time.Duration
	}{
		{
			name: "HTTP2 client ping enabled with custom timeouts",
			envVars: map[string]string{
				"HTTP2_TRANSPORT_PING_ENABLED":      "true",
				"HTTP2_TRANSPORT_READ_IDLE_TIMEOUT": "45s",
				"HTTP2_TRANSPORT_PING_TIMEOUT":      "10s",
			},
			expectedPingEnabled:     true,
			expectedReadIdleTimeout: 45 * time.Second,
			expectedPingTimeout:     10 * time.Second,
		},
		{
			name: "HTTP2 client ping disabled",
			envVars: map[string]string{
				"HTTP2_TRANSPORT_PING_ENABLED": "false",
			},
			expectedPingEnabled:     false,
			expectedReadIdleTimeout: 30 * time.Second,
			expectedPingTimeout:     5 * time.Second,
		},
		{
			name: "HTTP2 client ping enabled with default timeouts",
			envVars: map[string]string{
				"HTTP2_TRANSPORT_PING_ENABLED": "true",
			},
			expectedPingEnabled:     true,
			expectedReadIdleTimeout: 30 * time.Second,
			expectedPingTimeout:     5 * time.Second,
		},
		{
			name: "HTTP2 client ping enabled with defaults values when using invalid timeouts",
			envVars: map[string]string{
				"HTTP2_TRANSPORT_PING_ENABLED":      "true",
				"HTTP2_TRANSPORT_READ_IDLE_TIMEOUT": "whatever",
				"HTTP2_TRANSPORT_PING_TIMEOUT":      "whatever",
			},
			expectedPingEnabled:     true,
			expectedReadIdleTimeout: 30 * time.Second,
			expectedPingTimeout:     5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Reinitialize the HTTP client
			initOsEnv()

			// Check if the ping is enabled
			assert.Equal(t, tt.expectedPingEnabled, IsHTTP2TransportPingEnabled())

			// Check the read idle timeout
			assert.Equal(t, tt.expectedReadIdleTimeout, GetHTTP2TransportReadIdleTimeout())

			// Check the ping timeout
			assert.Equal(t, tt.expectedPingTimeout, GetHTTP2TransportPingTimeout())

			// Unset environment variables
			for key := range tt.envVars {
				_ = os.Unsetenv(key)
			}
		})
	}
}

func TestDefaultHTTPClient(t *testing.T) {
	assert.NotNil(t, DefaultHTTPClient)
	assert.IsType(t, &http.Client{}, DefaultHTTPClient)
	assert.Equal(t, 100, DefaultHTTPClient.Transport.(*http.Transport).MaxIdleConns)
	assert.Equal(t, 100, DefaultHTTPClient.Transport.(*http.Transport).MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, DefaultHTTPClient.Transport.(*http.Transport).IdleConnTimeout)
	assert.Equal(t, 10*time.Second, DefaultHTTPClient.Transport.(*http.Transport).TLSHandshakeTimeout)
	assert.Equal(t, 1*time.Second, DefaultHTTPClient.Transport.(*http.Transport).ExpectContinueTimeout)
}
