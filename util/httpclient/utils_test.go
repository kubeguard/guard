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
		name                           string
		envVars                        map[string]string
		expectedHTTP2ClientPingEnabled bool
		expectedReadIdleTimeout        time.Duration
		expectedPingTimeout            time.Duration
	}{
		{
			name: "HTTP2 client ping enabled with custom timeouts",
			envVars: map[string]string{
				"HTTP2_TRANSPORT_PING_ENABLED":      "true",
				"HTTP2_TRANSPORT_READ_IDLE_TIMEOUT": "45s",
				"HTTP2_TRANSPORT_PING_TIMEOUT":      "10s",
			},
			expectedHTTP2ClientPingEnabled: true,
			expectedReadIdleTimeout:        45 * time.Second,
			expectedPingTimeout:            10 * time.Second,
		},
		{
			name: "HTTP2 client ping disabled",
			envVars: map[string]string{
				"HTTP2_TRANSPORT_PING_ENABLED": "false",
			},
			expectedHTTP2ClientPingEnabled: false,
			expectedReadIdleTimeout:        30 * time.Second,
			expectedPingTimeout:            5 * time.Second,
		},
		{
			name: "HTTP2 client ping enabled with default timeouts",
			envVars: map[string]string{
				"HTTP2_TRANSPORT_PING_ENABLED": "true",
			},
			expectedHTTP2ClientPingEnabled: true,
			expectedReadIdleTimeout:        30 * time.Second,
			expectedPingTimeout:            5 * time.Second,
		},
		{
			name: "HTTP2 client ping enabled with defaults values when using invalid timeouts",
			envVars: map[string]string{
				"HTTP2_TRANSPORT_PING_ENABLED":      "true",
				"HTTP2_TRANSPORT_READ_IDLE_TIMEOUT": "whatever",
				"HTTP2_TRANSPORT_PING_TIMEOUT":      "whatever",
			},
			expectedHTTP2ClientPingEnabled: true,
			expectedReadIdleTimeout:        30 * time.Second,
			expectedPingTimeout:            5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Reinitialize the HTTP client
			initOsEnv()

			// Check if the HTTP2 client ping is enabled
			assert.Equal(t, tt.expectedHTTP2ClientPingEnabled, IsHTTP2ClientPingEnabled())

			// Check the read idle timeout
			assert.Equal(t, tt.expectedReadIdleTimeout, GetHTTP2TransportReadIdleTimeout())

			// Check the ping timeout
			assert.Equal(t, tt.expectedPingTimeout, GetHTTP2TransportPingTimeout())

			// Unset environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
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
