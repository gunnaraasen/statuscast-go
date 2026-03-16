package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const incidentItemJSON = `{
	"id": 100,
	"title": "Database Down",
	"dateCreated": "2024-01-15T10:00:00Z",
	"posts": [{"id":1,"text":"Investigating.","postType":"Investigating","isPublished":true,"date":"2024-01-15T10:00:00Z"}],
	"affectedComponents": [{"componentId":42}]
}`

const incidentListJSON = `{
	"items": [
		{"id":100,"title":"Database Down","dateCreated":"2024-01-15T10:00:00Z","posts":[{"postType":"Investigating","date":"2024-01-15T10:00:00Z"}]},
		{"id":101,"title":"Scheduled Maintenance","dateCreated":"2024-01-16T08:00:00Z","posts":[{"postType":"Closed","date":"2024-01-16T08:00:00Z"}]}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

const incidentResolvedJSON = `{
	"id": 100,
	"title": "Database Down",
	"dateCreated": "2024-01-15T10:00:00Z",
	"endDate": "2024-01-15T11:30:00Z",
	"posts": [{"id":2,"text":"Resolved.","postType":"Closed","isPublished":true,"date":"2024-01-15T11:30:00Z"}]
}`

func TestIncidentsList_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, incidentListJSON))

	out, err := runCmd(t, mux, "incidents", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Database Down") {
		t.Errorf("output missing incident title; got:\n%s", out)
	}
	if !strings.Contains(out, "Scheduled Maintenance") {
		t.Errorf("output missing second incident; got:\n%s", out)
	}
}

func TestIncidentsList_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, incidentListJSON))

	out, err := runCmd(t, mux, "--json", "incidents", "list")
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

func TestIncidentsList_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", statusOnly(401))

	_, err := runCmd(t, mux, "incidents", "list")
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestIncidentsList_SinceUntilParsed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, `{"items":[],"totalItems":0,"page":1,"pages":1}`))

	_, err := runCmd(t, mux, "incidents", "list", "--since", "2024-01-01", "--until", "2024-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIncidentsList_InvalidSince(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "incidents", "list", "--since", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid --since, got nil")
	}
}

func TestIncidentsGet_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/incident/{id}", jsonBody(200, incidentItemJSON))

	out, err := runCmd(t, mux, "incidents", "get", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Database Down") {
		t.Errorf("output missing title; got:\n%s", out)
	}
	if !strings.Contains(out, "investigating") {
		t.Errorf("output missing status; got:\n%s", out)
	}
}

func TestIncidentsGet_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "incidents", "get")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestIncidentsCreate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incident", jsonBody(200, incidentItemJSON))

	out, err := runCmd(t, mux,
		"incidents", "create",
		"--title", "Database Down",
		"--message", "Investigating.",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "100") {
		t.Errorf("output missing incident id; got:\n%s", out)
	}
}

func TestIncidentsCreate_WithComponents(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incident", jsonBody(200, incidentItemJSON))

	_, err := runCmd(t, mux,
		"incidents", "create",
		"--title", "Multi-Component Outage",
		"--message", "Investigating.",
		"--component", "42",
		"--component", "43",
		"--notify",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIncidentsCreate_MissingTitle(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "incidents", "create", "--message", "Investigating.")
	if err == nil {
		t.Fatal("expected error for missing --title, got nil")
	}
}

func TestIncidentsCreate_MissingMessage(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "incidents", "create", "--title", "Outage")
	if err == nil {
		t.Fatal("expected error for missing --message, got nil")
	}
}

func TestIncidentsUpdate_Success(t *testing.T) {
	const putResponse = `{
		"id": 100,
		"title": "Database Down",
		"dateCreated": "2024-01-15T10:00:00Z",
		"posts": [
			{"id":1,"text":"Investigating.","postType":"Investigating","isPublished":true,"date":"2024-01-15T10:00:00Z"},
			{"id":2,"text":"Root cause found.","postType":"Identified","isPublished":true,"date":"2024-01-15T10:30:00Z"}
		]
	}`
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", jsonBody(200, putResponse))

	out, err := runCmd(t, mux,
		"incidents", "update", "100",
		"--message", "Root cause found.",
		"--status", "identified",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "100") {
		t.Errorf("output missing incident id; got:\n%s", out)
	}
}

func TestIncidentsUpdate_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "incidents", "update", "--message", "Update.")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestIncidentsResolve_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", jsonBody(200, incidentResolvedJSON))

	out, err := runCmd(t, mux,
		"incidents", "resolve", "100",
		"--message", "Issue is resolved.",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "100") {
		t.Errorf("output missing incident id; got:\n%s", out)
	}
}

func TestIncidentsResolve_MissingMessage(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "incidents", "resolve", "100")
	if err == nil {
		t.Fatal("expected error for missing --message, got nil")
	}
}

func TestIncidentsResolve_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "incidents", "resolve", "--message", "Done.")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}
