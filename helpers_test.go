package bitculator

// Test doubles: every test runs against a local net/http/httptest server —
// no network, no mocks. newTestClient wires a Client to a per-test server.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.Handler, opts ...Option) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	base := append([]Option{WithBaseURL(server.URL), WithUserAgent("bitculator-go/test")}, opts...)
	return New("test-key", base...)
}

func respond(w http.ResponseWriter, status int, body string, headers map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	for key, value := range headers {
		w.Header().Set(key, value)
	}
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

func errorBody(code, message string) string {
	return fmt.Sprintf(`{"error":{"code":%q,"message":%q,"details":null}}`, code, message)
}
