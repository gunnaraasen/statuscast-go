package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	version = "v1.2.3"
	t.Cleanup(func() { version = "dev" })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	runErr := newApp().Run(context.Background(), []string{"statuscast", "version"})

	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	r.Close()

	if runErr != nil {
		t.Fatalf("unexpected error: %v", runErr)
	}
	if got := strings.TrimSpace(buf.String()); got != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %q", got)
	}
}

func TestVersionCommand_NoAPIKeyRequired(t *testing.T) {
	t.Setenv("STATUSCAST_API_KEY", "")

	err := newApp().Run(context.Background(), []string{"statuscast", "version"})
	if err != nil {
		t.Fatalf("version should not require an API key, got: %v", err)
	}
}
