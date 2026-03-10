package statuscast_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	statuscast "statuscast-go"
)

// newMockClient starts an httptest.Server with handler and returns a Client
// configured to use it. The server is shut down when the test ends.
func newMockClient(t *testing.T, handler http.Handler) *statuscast.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := statuscast.New(
		statuscast.WithAPIKey("test-key"),
		statuscast.WithBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

// jsonHandler returns an http.HandlerFunc that writes the given JSON with
// the given status code and Content-Type: application/json.
func jsonHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// statusHandler returns an http.HandlerFunc that writes only the given status code.
func statusHandler(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}
