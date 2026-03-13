package statuscast_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	statuscast "statuscast-go"
)

// mustNew creates a Client from opts, panicking on error. Using panic (rather
// than t.Fatal) lets SA5011 model this as a definite non-nil return path.
func mustNew(t *testing.T, opts ...statuscast.Option) *statuscast.Client {
	t.Helper()
	c, err := statuscast.New(opts...)
	if err != nil {
		panic("statuscast.New: " + err.Error())
	}
	return c
}

// newMockClient starts an httptest.Server with handler and returns a Client
// configured to use it. The server is shut down when the test ends.
func newMockClient(t *testing.T, handler http.Handler) *statuscast.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return mustNew(t,
		statuscast.WithAPIKey("test-key"),
		statuscast.WithBaseURL(srv.URL),
	)
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
