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
package httpclienttesting

import (
	"net/http"
	"net/http/httptest"
	"sync"

	"go.kubeguard.dev/guard/util/httpclient"
)

var hijackOnce = new(sync.Once)

// HijackDefaultHTTPClientTransportWithSelfSignedTLS injects the self-signed TLS cert from httptest server
// into the global shared http client transport.
// This is necessary to allow the checkAccess HTTP client to communicate with the test server over HTTPS.
// This call is invoked in test init as the DefaultHTTPClient is shared across all modules,
// which we cannot mutate without causing data race.
func HijackDefaultHTTPClientTransportWithSelfSignedTLS() {
	hijackOnce.Do(func() {
		ts := httptest.NewTLSServer(http.NotFoundHandler())
		if tt, ok := ts.Client().Transport.(*http.Transport); ok {
			if httpClientTransport, ok := httpclient.DefaultHTTPClient.Transport.(*http.Transport); ok {
				// Copy the TLS config from the test server to the checkAccess HTTP client
				httpClientTransport.TLSClientConfig = tt.TLSClientConfig
			}
		}
	})
}
