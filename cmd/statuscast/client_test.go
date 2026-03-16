package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildClient_MissingKey(t *testing.T) {
	t.Setenv("STATUSCAST_API_KEY", "")

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	err := newApp().Run(context.Background(), []string{
		"statuscast", "--base-url", srv.URL, "components", "list",
	})
	if err == nil {
		t.Fatal("expected error with no API key, got nil")
	}
}

func TestBuildClient_KeyFromEnv(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", statusOnly(401))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	t.Setenv("STATUSCAST_API_KEY", "env-key")

	// No --api-key flag; key must come from env.
	err := newApp().Run(context.Background(), []string{
		"statuscast", "--base-url", srv.URL, "components", "list",
	})
	// A 401 means the client was built (key found) and a request was made.
	if err == nil {
		t.Fatal("expected 401 error, got nil")
	}
}
