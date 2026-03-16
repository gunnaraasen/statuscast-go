package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const uptimeJSON = `{"uptime":99.97,"start":"2024-01-01T00:00:00Z","end":"2024-01-31T23:59:59Z"}`

const incidentSummaryListJSON = `{
	"items": [
		{
			"id": 100,
			"title": "Outage",
			"dateCreated": "2024-01-15T10:05:00Z",
			"startDate": "2024-01-15T10:00:00Z",
			"endDate": "2024-01-15T11:30:00Z",
			"affectedComponents": [{"componentId":42}]
		},
		{
			"id": 101,
			"title": "Brief Outage",
			"dateCreated": "2024-01-20T09:00:00Z",
			"startDate": "2024-01-20T08:55:00Z",
			"endDate": "2024-01-20T09:30:00Z"
		}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

func TestReportsUptime_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components/uptime", jsonBody(200, uptimeJSON))

	out, err := runCmd(t, mux, "reports", "uptime")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "99.9700%") {
		t.Errorf("output missing uptime percentage; got:\n%s", out)
	}
}

func TestReportsUptime_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components/uptime", jsonBody(200, uptimeJSON))

	out, err := runCmd(t, mux, "--json", "reports", "uptime")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result []any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 uptime record; got %d", len(result))
	}
}

func TestReportsUptime_WithDateRange(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components/uptime", jsonBody(200, uptimeJSON))

	_, err := runCmd(t, mux, "reports", "uptime", "--since", "2024-01-01", "--until", "2024-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReportsUptime_InvalidSince(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "reports", "uptime", "--since", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid --since, got nil")
	}
}

func TestReportsIncidentSummary_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, incidentSummaryListJSON))

	out, err := runCmd(t, mux,
		"reports", "incident-summary",
		"--since", "2024-01-01",
		"--until", "2024-01-31",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("output missing total incident count; got:\n%s", out)
	}
	if !strings.Contains(out, "Mean Time to Detect") {
		t.Errorf("output missing MTTD label; got:\n%s", out)
	}
	if !strings.Contains(out, "Mean Time to Resolve") {
		t.Errorf("output missing MTTR label; got:\n%s", out)
	}
}

func TestReportsIncidentSummary_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, incidentSummaryListJSON))

	out, err := runCmd(t, mux, "--json",
		"reports", "incident-summary",
		"--since", "2024-01-01",
		"--until", "2024-01-31",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if result["total_incidents"].(float64) != 2 {
		t.Errorf("total_incidents = %v; want 2", result["total_incidents"])
	}
}

func TestReportsIncidentSummary_MissingSince(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "reports", "incident-summary", "--until", "2024-01-31")
	if err == nil {
		t.Fatal("expected error for missing --since, got nil")
	}
}

func TestReportsIncidentSummary_MissingUntil(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "reports", "incident-summary", "--since", "2024-01-01")
	if err == nil {
		t.Fatal("expected error for missing --until, got nil")
	}
}
