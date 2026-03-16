package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// runCmd starts an httptest.Server backed by handler, runs the CLI with the
// given args (global --api-key and --base-url are prepended automatically),
// captures stdout, and returns the output along with any error returned by Run.
func runCmd(t *testing.T, handler http.Handler, args ...string) (string, error) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fullArgs := append(
		[]string{"statuscast", "--api-key", "test-key", "--base-url", srv.URL},
		args...,
	)
	runErr := newApp().Run(context.Background(), fullArgs)

	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	r.Close()

	return buf.String(), runErr
}

// jsonBody returns an http.HandlerFunc that writes a JSON body with status code.
func jsonBody(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// statusOnly returns an http.HandlerFunc that writes only the given status code.
func statusOnly(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}
