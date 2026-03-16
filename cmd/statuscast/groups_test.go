package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const groupListJSON = `{
	"items": [
		{"id":1,"name":"Engineering","dateCreated":"2024-01-01T00:00:00Z"},
		{"id":2,"name":"Support","dateCreated":"2024-01-02T00:00:00Z"}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

func TestGroupsList_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/groups", jsonBody(200, groupListJSON))

	out, err := runCmd(t, mux, "groups", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Engineering") {
		t.Errorf("output missing group name; got:\n%s", out)
	}
	if !strings.Contains(out, "Support") {
		t.Errorf("output missing second group; got:\n%s", out)
	}
}

func TestGroupsList_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/groups", jsonBody(200, groupListJSON))

	out, err := runCmd(t, mux, "--json", "groups", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	items, ok := result["items"].([]any)
	if !ok || len(items) != 2 {
		t.Errorf("expected 2 items; got %v", result["items"])
	}
}

func TestGroupsList_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/groups", statusOnly(401))

	_, err := runCmd(t, mux, "groups", "list")
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestGroupsList_Pagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/groups", jsonBody(200, `{"items":[],"totalItems":0,"page":2,"pages":5}`))

	_, err := runCmd(t, mux, "groups", "list", "--page", "2", "--per-page", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
