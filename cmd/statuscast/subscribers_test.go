package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

const subscriberJSON = `{"id":10,"emailAddress":"user@example.com","subscribeToIncidentPosts":true}`

const subscriberListJSON = `{
	"items": [
		{"id":10,"emailAddress":"user@example.com","subscribeToIncidentPosts":true},
		{"id":11,"emailAddress":"other@example.com","subscribeToIncidentPosts":true}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

func TestSubscribersList_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscribers/search", jsonBody(200, subscriberListJSON))

	out, err := runCmd(t, mux, "subscribers", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("output missing email; got:\n%s", out)
	}
	if !strings.Contains(out, "other@example.com") {
		t.Errorf("output missing second email; got:\n%s", out)
	}
}

func TestSubscribersList_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscribers/search", jsonBody(200, subscriberListJSON))

	out, err := runCmd(t, mux, "--json", "subscribers", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if result["total_count"].(float64) != 2 {
		t.Errorf("total_count = %v; want 2", result["total_count"])
	}
}

func TestSubscribersList_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscribers/search", statusOnly(401))

	_, err := runCmd(t, mux, "subscribers", "list")
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestSubscribersGet_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/subscriber/{id}", jsonBody(200, subscriberJSON))

	out, err := runCmd(t, mux, "subscribers", "get", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("output missing email; got:\n%s", out)
	}
}

func TestSubscribersGet_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "subscribers", "get")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestSubscribersAdd_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", jsonBody(200, subscriberJSON))

	out, err := runCmd(t, mux, "subscribers", "add", "--email", "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "10") {
		t.Errorf("output missing subscriber id; got:\n%s", out)
	}
}

func TestSubscribersAdd_MissingEmail(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "subscribers", "add")
	if err == nil {
		t.Fatal("expected error for missing --email, got nil")
	}
}

func TestSubscribersUpdate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/subscriber", jsonBody(200, subscriberJSON))

	out, err := runCmd(t, mux, "subscribers", "update", "10", "--group", "5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "10") {
		t.Errorf("output missing subscriber id; got:\n%s", out)
	}
}

func TestSubscribersUpdate_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "subscribers", "update", "--group", "5")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestSubscribersRemove_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/subscriber/{id}", statusOnly(200))

	out, err := runCmd(t, mux, "subscribers", "remove", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "10") {
		t.Errorf("output missing subscriber id; got:\n%s", out)
	}
}

func TestSubscribersRemove_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "subscribers", "remove")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestSubscribersBulkImport_Success(t *testing.T) {
	// Register handler for each POST (one per subscriber row).
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", jsonBody(200, subscriberJSON))

	// Write a temp CSV file.
	f := t.TempDir() + "/subscribers.csv"
	if err := writeFile(t, f, "email\nuser@example.com\nother@example.com\n"); err != nil {
		t.Fatal(err)
	}

	out, err := runCmd(t, mux, "subscribers", "bulk-import", "--file", f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("expected imported count of 2 in output; got:\n%s", out)
	}
}

func TestSubscribersBulkImport_MissingFile(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "subscribers", "bulk-import", "--file", "/nonexistent/file.csv")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSubscribersBulkImport_NoEmailColumn(t *testing.T) {
	mux := http.NewServeMux()

	f := t.TempDir() + "/bad.csv"
	if err := writeFile(t, f, "name,phone\nAlice,555-1234\n"); err != nil {
		t.Fatal(err)
	}

	_, err := runCmd(t, mux, "subscribers", "bulk-import", "--file", f)
	if err == nil {
		t.Fatal("expected error for missing email column, got nil")
	}
}

// writeFile writes content to path, returning any error.
func writeFile(t *testing.T, path, content string) error {
	t.Helper()
	return os.WriteFile(path, []byte(content), 0o600)
}
