package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// testServerAndClient creates a test HTTP server with the given handler and returns
// a Client configured to use it. The server is automatically closed when the test ends.
// testServerAndClient creates a test HTTP server with the given handler and returns
// a Client configured to use it. The server is automatically closed when the test ends.
// A default Content-Type of application/json is set on responses; handlers can override it.
func testServerAndClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	})
	server := httptest.NewServer(wrapped)
	t.Cleanup(server.Close)
	return NewClient("jwt-token", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
}
