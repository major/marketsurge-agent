package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// testServerAndClient creates a test HTTP server with the given handler and returns
// a Client configured to use it. The server is automatically closed when the test ends.
func testServerAndClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := NewClient("jwt-token")
	c.Endpoint = server.URL
	c.HTTPClient = server.Client()
	return c
}
