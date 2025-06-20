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
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/http2"
	"k8s.io/klog/v2"
)

var (
	DefaultHTTPClient             *http.Client
	http2TransportPingEnabled     bool
	http2TransportReadIdleTimeout time.Duration
	http2TransportPingTimeout     time.Duration
)

const (
	readIdleTimeout = 30 * time.Second
	pingTimeout     = 5 * time.Second
)

// initOsEnv initializes the environment variables
// We are enabling 3 env variables to control the HTTP2 client ping feature:
// 1. HTTP2_TRANSPORT_PING_ENABLED: If set to true, the transport will setup http2 ping
// 2. HTTP2_TRANSPORT_READ_IDLE_TIMEOUT: The duration after which the connection will be closed if it has been idle
// 3. HTTP2_TRANSPORT_PING_TIMEOUT: The duration after which the connection will be closed if there's no response to the ping
func initOsEnv() {
	http2TransportPingEnabled, _ = strconv.ParseBool(os.Getenv("HTTP2_TRANSPORT_PING_ENABLED"))
	http2TransportReadIdleTimeout = readIdleTimeout
	if val, err := time.ParseDuration(os.Getenv("HTTP2_TRANSPORT_READ_IDLE_TIMEOUT")); err == nil && val > 0 {
		http2TransportReadIdleTimeout = val
	}
	http2TransportPingTimeout = pingTimeout
	if val, err := time.ParseDuration(os.Getenv("HTTP2_TRANSPORT_PING_TIMEOUT")); err == nil && val > 0 {
		http2TransportPingTimeout = val
	}
}

func init() {
	defaultTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	initOsEnv()
	// If HTTP2_TRANSPORT_PING_ENABLED env var is set, the transport will setup http2 ping
	if http2TransportPingEnabled {
		tr2 := defaultTransport.Clone()
		if http2Transport, err := http2.ConfigureTransports(tr2); err == nil {
			klog.V(10).Infof("HTTP2 transport ping enabled with read idle timeout: %v and ping timeout: %v", http2TransportReadIdleTimeout, http2TransportPingTimeout)
			// if the connection has been idle for 30 seconds, send a ping frame for a health check
			http2Transport.ReadIdleTimeout = http2TransportReadIdleTimeout
			// if there's no response to the ping within the timeout, the connection will be closed
			http2Transport.PingTimeout = http2TransportPingTimeout
			DefaultHTTPClient = &http.Client{
				Transport: tr2,
			}
			return
		}
	}
	DefaultHTTPClient = &http.Client{
		Transport: defaultTransport,
	}
}

// Getter functions to access the read-only variables
func IsHTTP2TransportPingEnabled() bool {
	return http2TransportPingEnabled
}

func GetHTTP2TransportReadIdleTimeout() time.Duration {
	return http2TransportReadIdleTimeout
}

func GetHTTP2TransportPingTimeout() time.Duration {
	return http2TransportPingTimeout
}
