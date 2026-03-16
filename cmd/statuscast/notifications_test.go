package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const templateJSON = `{"id":5,"subject":"Incident Update","contents":"Update body"}`

const templateListJSON = `[
	{"id":5,"subject":"Incident Update","contents":"Update body"},
	{"id":6,"subject":"Resolution","contents":"Issue resolved"}
]`

func TestTemplatesList_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/contenttemplate", jsonBody(200, templateListJSON))

	out, err := runCmd(t, mux, "notifications", "templates", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Incident Update") {
		t.Errorf("output missing template subject; got:\n%s", out)
	}
	if !strings.Contains(out, "Resolution") {
		t.Errorf("output missing second template; got:\n%s", out)
	}
}

func TestTemplatesList_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/contenttemplate", jsonBody(200, templateListJSON))

	out, err := runCmd(t, mux, "--json", "notifications", "templates", "list")
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

func TestTemplatesList_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/contenttemplate", statusOnly(401))

	_, err := runCmd(t, mux, "notifications", "templates", "list")
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestTemplatesCreate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/contenttemplate", jsonBody(200, templateJSON))

	out, err := runCmd(t, mux,
		"notifications", "templates", "create",
		"--name", "Incident Update",
		"--channel", "email",
		"--body", "Update body",
		"--subject", "Incident Update",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("output missing template id; got:\n%s", out)
	}
}

func TestTemplatesCreate_MissingName(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "notifications", "templates", "create", "--channel", "email", "--body", "Body")
	if err == nil {
		t.Fatal("expected error for missing --name, got nil")
	}
}

func TestTemplatesCreate_MissingChannel(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "notifications", "templates", "create", "--name", "T", "--body", "Body")
	if err == nil {
		t.Fatal("expected error for missing --channel, got nil")
	}
}

func TestTemplatesCreate_MissingBody(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "notifications", "templates", "create", "--name", "T", "--channel", "email")
	if err == nil {
		t.Fatal("expected error for missing --body, got nil")
	}
}

func TestTemplatesUpdate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/contenttemplate", jsonBody(200, templateJSON))

	out, err := runCmd(t, mux,
		"notifications", "templates", "update", "5",
		"--body", "Updated body",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("output missing template id; got:\n%s", out)
	}
}

func TestTemplatesUpdate_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "notifications", "templates", "update", "--body", "Body")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}
