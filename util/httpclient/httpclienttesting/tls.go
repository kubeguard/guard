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
