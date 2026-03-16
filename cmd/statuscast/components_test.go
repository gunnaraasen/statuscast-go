package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const componentListJSON = `[
	{"id":42,"name":"API Server","status":"Available"},
	{"id":43,"name":"Database","status":"DegradedPerformance","parentId":42}
]`

const componentJSON = `{"id":42,"name":"API Server","description":"Primary API","status":"Available"}`

func TestComponentsList_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", jsonBody(200, componentListJSON))

	out, err := runCmd(t, mux, "components", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "API Server") {
		t.Errorf("output missing component name; got:\n%s", out)
	}
	if !strings.Contains(out, "operational") {
		t.Errorf("output missing status; got:\n%s", out)
	}
	if !strings.Contains(out, "Database") {
		t.Errorf("output missing second component; got:\n%s", out)
	}
}

func TestComponentsList_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", jsonBody(200, componentListJSON))

	out, err := runCmd(t, mux, "--json", "components", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	items, ok := result["items"].([]any)
	if !ok || len(items) != 2 {
		t.Errorf("expected items array of length 2; got: %v", result["items"])
	}
}

func TestComponentsList_ParentFilter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", jsonBody(200, componentListJSON))

	out, err := runCmd(t, mux, "components", "list", "--parent-id", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "API Server") {
		t.Errorf("root component should be filtered out; got:\n%s", out)
	}
	if !strings.Contains(out, "Database") {
		t.Errorf("child component should appear; got:\n%s", out)
	}
}

func TestComponentsList_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", statusOnly(401))

	_, err := runCmd(t, mux, "components", "list")
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestComponentsGet_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/component/{id}", jsonBody(200, componentJSON))

	out, err := runCmd(t, mux, "components", "get", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "API Server") {
		t.Errorf("output missing name; got:\n%s", out)
	}
	if !strings.Contains(out, "Primary API") {
		t.Errorf("output missing description; got:\n%s", out)
	}
}

func TestComponentsGet_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "components", "get")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestComponentsGet_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/component/{id}", jsonBody(200, componentJSON))

	out, err := runCmd(t, mux, "--json", "components", "get", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var comp map[string]any
	if err := json.Unmarshal([]byte(out), &comp); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if comp["id"] != "42" {
		t.Errorf("id = %v; want %q", comp["id"], "42")
	}
}

func TestComponentsCreate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/component", jsonBody(200, `{"id":99,"name":"New Component","status":"Available"}`))

	out, err := runCmd(t, mux, "components", "create", "--name", "New Component")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "99") {
		t.Errorf("output missing component id; got:\n%s", out)
	}
}

func TestComponentsCreate_MissingName(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "components", "create")
	if err == nil {
		t.Fatal("expected error for missing --name, got nil")
	}
}

func TestComponentsUpdate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/component", jsonBody(200, `{"id":42,"name":"Updated","status":"Available"}`))

	out, err := runCmd(t, mux, "components", "update", "42", "--name", "Updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("output missing id; got:\n%s", out)
	}
}

func TestComponentsUpdate_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "components", "update", "--name", "X")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestComponentsDelete_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/component", statusOnly(200))

	out, err := runCmd(t, mux, "components", "delete", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("output missing id; got:\n%s", out)
	}
}

func TestComponentsDelete_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "components", "delete")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestComponentsSetStatus_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/component", jsonBody(200, `{"id":42,"name":"API Server","status":"Maintenance"}`))

	out, err := runCmd(t, mux, "components", "set-status", "42", "under_maintenance")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("output missing id; got:\n%s", out)
	}
}

func TestComponentsSetStatus_MissingArgs(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "components", "set-status", "42")
	if err == nil {
		t.Fatal("expected error for missing status arg, got nil")
	}
}
