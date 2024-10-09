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
)

var DefaultHTTPClient *http.Client

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
	// If HTTP2_CLIENT_PING_ENABLED env var is set, the transport will setup http2 ping
	v, _ := strconv.ParseBool(os.Getenv("HTTP2_CLIENT_PING_ENABLED"))
	if v {
		tr2 := defaultTransport.Clone()
		if http2Transport, err := http2.ConfigureTransports(tr2); err == nil {
			// if the connection has been idle for 3 seconds, send a ping frame for a health check
			http2Transport.ReadIdleTimeout = 3 * time.Second
			// if there's no response to the ping within the timeout, the connection will be closed
			http2Transport.PingTimeout = 2 * time.Second
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
