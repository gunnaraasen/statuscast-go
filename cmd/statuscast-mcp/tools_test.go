package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	statuscast "statuscast-go"
)

// newTestSession starts a mock HTTP server backed by handler, creates a
// statuscast client pointing at it, registers all MCP tools, and returns a
// connected MCP ClientSession ready for CallTool calls.
func newTestSession(t *testing.T, handler http.Handler) *mcp.ClientSession {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := statuscast.New(
		statuscast.WithAPIKey("test-key"),
		statuscast.WithBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("statuscast.New: %v", err)
	}

	impl := &mcp.Implementation{Name: "test"}
	s := mcp.NewServer(impl, nil)
	registerTools(s, client)

	ct, st := mcp.NewInMemoryTransports()

	ctx := context.Background()
	if _, err := s.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server.Connect: %v", err)
	}

	c := mcp.NewClient(impl, nil)
	cs, err := c.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	return cs
}

// callTool invokes a named MCP tool with the given arguments map.
func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	return cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
}

// resultText extracts the text from the first TextContent in the result.
func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		return ""
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

// jsonBody returns an http.HandlerFunc that writes a JSON body with status.
func jsonBody(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// statusOnly returns an http.HandlerFunc that writes only the given status.
func statusOnly(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}

// ─── JSON fixtures ────────────────────────────────────────────────────────────

const componentListJSON = `[
	{"id":42,"name":"API Server","status":"Available"},
	{"id":43,"name":"Database","status":"DegradedPerformance"}
]`

const componentJSON = `{"id":42,"name":"API Server","status":"Maintenance"}`

const incidentListJSON = `{
	"items": [
		{"id":100,"title":"Database Down","dateCreated":"2024-01-15T10:00:00Z","posts":[{"postType":"Investigating","date":"2024-01-15T10:00:00Z"}]},
		{"id":101,"title":"Scheduled Maintenance","dateCreated":"2024-01-16T08:00:00Z","posts":[{"postType":"Closed","date":"2024-01-16T08:00:00Z"}]}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

const incidentItemJSON = `{
	"id": 100,
	"title": "Database Down",
	"dateCreated": "2024-01-15T10:00:00Z",
	"posts": [{"id":1,"text":"Investigating.","postType":"Investigating","isPublished":true,"date":"2024-01-15T10:00:00Z"}],
	"affectedComponents": [{"componentId":42}]
}`

const incidentUpdatedJSON = `{
	"id": 100,
	"title": "Database Down",
	"dateCreated": "2024-01-15T10:00:00Z",
	"posts": [
		{"id":1,"text":"Investigating.","postType":"Investigating","isPublished":true,"date":"2024-01-15T10:00:00Z"},
		{"id":2,"text":"Root cause found.","postType":"Identified","isPublished":true,"date":"2024-01-15T10:30:00Z"}
	]
}`

const incidentResolvedJSON = `{
	"id": 100,
	"title": "Database Down",
	"dateCreated": "2024-01-15T10:00:00Z",
	"endDate": "2024-01-15T11:30:00Z",
	"posts": [{"id":2,"text":"Resolved.","postType":"Closed","isPublished":true,"date":"2024-01-15T11:30:00Z"}]
}`

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
		}
	],
	"totalItems": 1,
	"page": 1,
	"pages": 1
}`

// ─── list_components ──────────────────────────────────────────────────────────

func TestListComponents_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", jsonBody(200, componentListJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "list_components", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "API Server") {
		t.Errorf("output missing first component; got:\n%s", text)
	}
	if !strings.Contains(text, "Database") {
		t.Errorf("output missing second component; got:\n%s", text)
	}
}

func TestListComponents_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", jsonBody(200, "[]"))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "list_components", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "No components found") {
		t.Errorf("expected empty message; got:\n%s", text)
	}
}

func TestListComponents_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", statusOnly(401))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "list_components", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for 401; text:\n%s", resultText(t, result))
	}
}

// ─── set_component_status ─────────────────────────────────────────────────────

func TestSetComponentStatus_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/component", jsonBody(200, componentJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "set_component_status", map[string]any{
		"component_id": "42",
		"status":       "under_maintenance",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "API Server") {
		t.Errorf("output missing component name; got:\n%s", text)
	}
	if !strings.Contains(text, "42") {
		t.Errorf("output missing component ID; got:\n%s", text)
	}
}

func TestSetComponentStatus_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/component", statusOnly(401))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "set_component_status", map[string]any{
		"component_id": "42",
		"status":       "operational",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for 401; text:\n%s", resultText(t, result))
	}
}

// ─── list_incidents ───────────────────────────────────────────────────────────

func TestListIncidents_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, incidentListJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "list_incidents", map[string]any{
		"active_only": false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "Database Down") {
		t.Errorf("output missing first incident; got:\n%s", text)
	}
	if !strings.Contains(text, "Scheduled Maintenance") {
		t.Errorf("output missing second incident; got:\n%s", text)
	}
}

func TestListIncidents_ActiveOnly(t *testing.T) {
	const activeListJSON = `{
		"items": [{"id":100,"title":"Active Outage","dateCreated":"2024-01-15T10:00:00Z","posts":[{"postType":"Investigating","date":"2024-01-15T10:00:00Z"}]}],
		"totalItems": 1,
		"page": 1,
		"pages": 1
	}`
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, activeListJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "list_incidents", map[string]any{
		"active_only": true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "Active Outage") {
		t.Errorf("output missing incident; got:\n%s", text)
	}
}

func TestListIncidents_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, `{"items":[],"totalItems":0,"page":1,"pages":1}`))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "list_incidents", map[string]any{
		"active_only": false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "No incidents found") {
		t.Errorf("expected empty message; got:\n%s", text)
	}
}

func TestListIncidents_ResolvedShowsTimestamp(t *testing.T) {
	const resolvedListJSON = `{
		"items": [{"id":101,"title":"Past Outage","dateCreated":"2024-01-15T10:00:00Z","endDate":"2024-01-15T11:00:00Z","posts":[{"postType":"Closed","date":"2024-01-15T11:00:00Z"}]}],
		"totalItems": 1,
		"page": 1,
		"pages": 1
	}`
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, resolvedListJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "list_incidents", map[string]any{
		"active_only": false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "resolved") {
		t.Errorf("output missing resolved state; got:\n%s", text)
	}
}

// ─── create_incident ──────────────────────────────────────────────────────────

func TestCreateIncident_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incident", jsonBody(200, incidentItemJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "create_incident", map[string]any{
		"title":   "Database Down",
		"message": "Investigating the issue.",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "100") {
		t.Errorf("output missing incident ID; got:\n%s", text)
	}
	if !strings.Contains(text, "Database Down") {
		t.Errorf("output missing incident title; got:\n%s", text)
	}
}

func TestCreateIncident_WithAllOptions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incident", jsonBody(200, incidentItemJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "create_incident", map[string]any{
		"title":      "API Outage",
		"message":    "Root cause identified.",
		"components": []any{"42", "43"},
		"status":     "identified",
		"post_type":  "outage",
		"notify":     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected tool error: %s", resultText(t, result))
	}
}

func TestCreateIncident_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incident", statusOnly(401))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "create_incident", map[string]any{
		"title":   "Outage",
		"message": "Something broke.",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for 401; text:\n%s", resultText(t, result))
	}
}

// ─── update_incident ──────────────────────────────────────────────────────────

func TestUpdateIncident_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", jsonBody(200, incidentUpdatedJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "update_incident", map[string]any{
		"incident_id": "100",
		"message":     "Root cause found.",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "100") {
		t.Errorf("output missing incident ID; got:\n%s", text)
	}
}

func TestUpdateIncident_WithStatusAndNotify(t *testing.T) {
	const monitoringJSON = `{
		"id": 100,
		"title": "Database Down",
		"dateCreated": "2024-01-15T10:00:00Z",
		"posts": [{"id":3,"text":"Monitoring.","postType":"Monitoring","isPublished":true,"date":"2024-01-15T11:00:00Z"}]
	}`
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", jsonBody(200, monitoringJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "update_incident", map[string]any{
		"incident_id": "100",
		"message":     "Monitoring closely.",
		"status":      "monitoring",
		"notify":      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected tool error: %s", resultText(t, result))
	}
}

func TestUpdateIncident_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", statusOnly(401))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "update_incident", map[string]any{
		"incident_id": "100",
		"message":     "Update.",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for 401; text:\n%s", resultText(t, result))
	}
}

// ─── resolve_incident ─────────────────────────────────────────────────────────

func TestResolveIncident_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", jsonBody(200, incidentResolvedJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "resolve_incident", map[string]any{
		"incident_id": "100",
		"message":     "Issue resolved.",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "100") {
		t.Errorf("output missing incident ID; got:\n%s", text)
	}
}

func TestResolveIncident_WithNotify(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", jsonBody(200, incidentResolvedJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "resolve_incident", map[string]any{
		"incident_id": "100",
		"message":     "Resolved after maintenance window.",
		"notify":      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected tool error: %s", resultText(t, result))
	}
}

func TestResolveIncident_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", statusOnly(401))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "resolve_incident", map[string]any{
		"incident_id": "100",
		"message":     "Resolved.",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for 401; text:\n%s", resultText(t, result))
	}
}

// ─── get_uptime_report ────────────────────────────────────────────────────────

func TestGetUptimeReport_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components/uptime", jsonBody(200, uptimeJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "get_uptime_report", map[string]any{
		"since": "2024-01-01T00:00:00Z",
		"until": "2024-01-31T23:59:59Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "99.9700") {
		t.Errorf("output missing uptime percentage; got:\n%s", text)
	}
}

func TestGetUptimeReport_InvalidSince(t *testing.T) {
	mux := http.NewServeMux()
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "get_uptime_report", map[string]any{
		"since": "not-a-date",
		"until": "2024-01-31T23:59:59Z",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid since, got false")
	}
}

func TestGetUptimeReport_InvalidUntil(t *testing.T) {
	mux := http.NewServeMux()
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "get_uptime_report", map[string]any{
		"since": "2024-01-01T00:00:00Z",
		"until": "not-a-date",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid until, got false")
	}
}

func TestGetUptimeReport_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components/uptime", statusOnly(401))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "get_uptime_report", map[string]any{
		"since": "2024-01-01T00:00:00Z",
		"until": "2024-01-31T23:59:59Z",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for 401; text:\n%s", resultText(t, result))
	}
}

// ─── get_incident_summary ─────────────────────────────────────────────────────

func TestGetIncidentSummary_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonBody(200, incidentSummaryListJSON))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "get_incident_summary", map[string]any{
		"since": "2024-01-01T00:00:00Z",
		"until": "2024-01-31T23:59:59Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "Total incidents: 1") {
		t.Errorf("output missing total count; got:\n%s", text)
	}
	if !strings.Contains(text, "MTTD:") {
		t.Errorf("output missing MTTD label; got:\n%s", text)
	}
	if !strings.Contains(text, "MTTR:") {
		t.Errorf("output missing MTTR label; got:\n%s", text)
	}
	if !strings.Contains(text, "By component:") {
		t.Errorf("output missing by-component breakdown; got:\n%s", text)
	}
}

func TestGetIncidentSummary_InvalidSince(t *testing.T) {
	mux := http.NewServeMux()
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "get_incident_summary", map[string]any{
		"since": "not-a-date",
		"until": "2024-01-31T23:59:59Z",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid since, got false")
	}
}

func TestGetIncidentSummary_InvalidUntil(t *testing.T) {
	mux := http.NewServeMux()
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "get_incident_summary", map[string]any{
		"since": "2024-01-01T00:00:00Z",
		"until": "not-a-date",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid until, got false")
	}
}

func TestGetIncidentSummary_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", statusOnly(401))
	cs := newTestSession(t, mux)

	result, err := callTool(t, cs, "get_incident_summary", map[string]any{
		"since": "2024-01-01T00:00:00Z",
		"until": "2024-01-31T23:59:59Z",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for 401; text:\n%s", resultText(t, result))
	}
}

// ─── formatDuration ───────────────────────────────────────────────────────────

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "N/A"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q; want %q", tt.d, got, tt.want)
		}
	}
}
