package statuscast_test

import (
	"context"
	"net/http"
	"testing"

	statuscast "statuscast-go"
)

const incidentJSON = `{
	"id": 100,
	"title": "Database Down",
	"dateCreated": "2024-01-15T10:00:00Z",
	"posts": [
		{"id": 1, "text": "We are investigating.", "postType": "Investigating", "isPublished": true, "date": "2024-01-15T10:00:00Z"}
	],
	"affectedComponents": [{"componentId": 42}]
}`

const incidentResolvedJSON = `{
	"id": 100,
	"title": "Database Down",
	"dateCreated": "2024-01-15T10:00:00Z",
	"endDate": "2024-01-15T11:30:00Z",
	"posts": [
		{"id": 2, "text": "Issue resolved.", "postType": "Closed", "isPublished": true, "date": "2024-01-15T11:30:00Z"}
	]
}`

const incidentListJSON = `{
	"items": [
		{"id": 100, "title": "Database Down", "dateCreated": "2024-01-15T10:00:00Z"},
		{"id": 101, "title": "Maintenance", "dateCreated": "2024-01-16T08:00:00Z"}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

func TestIncidentsCreate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incident", jsonHandler(200, incidentJSON))
	c := newMockClient(t, mux)

	inc, resp, err := c.Incidents.Create(context.Background(), statuscast.CreateIncidentRequest{
		Title:   "Database Down",
		Message: "We are investigating.",
		Status:  statuscast.StatusInvestigating,
		Notify:  false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if inc.ID != "100" {
		t.Errorf("ID = %q; want %q", inc.ID, "100")
	}
	if inc.Title != "Database Down" {
		t.Errorf("Title = %q; want %q", inc.Title, "Database Down")
	}
	if inc.Status != statuscast.StatusInvestigating {
		t.Errorf("Status = %q; want %q", inc.Status, statuscast.StatusInvestigating)
	}
	if len(inc.Updates) != 1 {
		t.Fatalf("len(Updates) = %d; want 1", len(inc.Updates))
	}
	if inc.Updates[0].Message != "We are investigating." {
		t.Errorf("Updates[0].Message = %q; want %q", inc.Updates[0].Message, "We are investigating.")
	}
	if len(inc.Components) != 1 || inc.Components[0] != "42" {
		t.Errorf("Components = %v; want [42]", inc.Components)
	}
}

func TestIncidentsCreate_WithComponents(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incident", jsonHandler(200, incidentJSON))
	c := newMockClient(t, mux)

	_, _, err := c.Incidents.Create(context.Background(), statuscast.CreateIncidentRequest{
		Title:      "Multi-Component",
		Components: []string{"42", "43"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIncidentsCreate_InvalidComponentID(t *testing.T) {
	c, _ := statuscast.New(statuscast.WithAPIKey("key"))
	_, _, err := c.Incidents.Create(context.Background(), statuscast.CreateIncidentRequest{
		Title:      "Test",
		Components: []string{"not-a-number"},
	})
	if err == nil {
		t.Fatal("expected error for invalid component ID, got nil")
	}
}

func TestIncidentsCreate_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incident", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Incidents.Create(context.Background(), statuscast.CreateIncidentRequest{Title: "T"})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestIncidentsGet_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/incident/{id}", jsonHandler(200, incidentJSON))
	c := newMockClient(t, mux)

	inc, resp, err := c.Incidents.Get(context.Background(), "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if inc.ID != "100" {
		t.Errorf("ID = %q; want %q", inc.ID, "100")
	}
	if inc.Title != "Database Down" {
		t.Errorf("Title = %q; want %q", inc.Title, "Database Down")
	}
}

func TestIncidentsGet_InvalidID(t *testing.T) {
	c, _ := statuscast.New(statuscast.WithAPIKey("key"))
	_, _, err := c.Incidents.Get(context.Background(), "bad-id")
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestIncidentsGet_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/incident/{id}", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Incidents.Get(context.Background(), "100")
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestIncidentsList_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonHandler(200, incidentListJSON))
	c := newMockClient(t, mux)

	result, resp, err := c.Incidents.List(context.Background(), statuscast.IncidentFilter{}, statuscast.Pagination{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d; want 2", result.TotalCount)
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d; want 2", len(result.Items))
	}
	if result.Items[0].ID != "100" {
		t.Errorf("Items[0].ID = %q; want %q", result.Items[0].ID, "100")
	}
	if result.TotalPages != 1 {
		t.Errorf("TotalPages = %d; want 1", result.TotalPages)
	}
}

func TestIncidentsList_ActiveOnly(t *testing.T) {
	// Verifies the request is made (filter is applied server-side).
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonHandler(200, `{"items":[{"id":100,"title":"Active","dateCreated":"2024-01-15T10:00:00Z"}],"totalItems":1,"page":1,"pages":1}`))
	c := newMockClient(t, mux)

	result, _, err := c.Incidents.List(context.Background(), statuscast.IncidentFilter{ActiveOnly: true}, statuscast.Pagination{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(Items) = %d; want 1", len(result.Items))
	}
}

func TestIncidentsList_Pagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonHandler(200, `{"items":[],"totalItems":50,"page":2,"pages":5}`))
	c := newMockClient(t, mux)

	result, _, err := c.Incidents.List(context.Background(), statuscast.IncidentFilter{}, statuscast.Pagination{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 2 {
		t.Errorf("Page = %d; want 2", result.Page)
	}
	if result.TotalPages != 5 {
		t.Errorf("TotalPages = %d; want 5", result.TotalPages)
	}
}

func TestIncidentsPostUpdate_Success(t *testing.T) {
	const body = `{
		"id": 100,
		"title": "Database Down",
		"dateCreated": "2024-01-15T10:00:00Z",
		"posts": [
			{"id": 1, "text": "First update.", "postType": "Investigating", "isPublished": true, "date": "2024-01-15T10:00:00Z"},
			{"id": 2, "text": "Root cause identified.", "postType": "Identified", "isPublished": true, "date": "2024-01-15T10:30:00Z"}
		]
	}`
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", jsonHandler(200, body))
	c := newMockClient(t, mux)

	update, resp, err := c.Incidents.PostUpdate(context.Background(), "100", statuscast.UpdateIncidentRequest{
		Message: "Root cause identified.",
		Status:  statuscast.StatusIdentified,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	// PostUpdate returns the last update.
	if update.Message != "Root cause identified." {
		t.Errorf("Message = %q; want %q", update.Message, "Root cause identified.")
	}
	if update.Status != statuscast.StatusIdentified {
		t.Errorf("Status = %q; want %q", update.Status, statuscast.StatusIdentified)
	}
}

func TestIncidentsPostUpdate_InvalidID(t *testing.T) {
	c, _ := statuscast.New(statuscast.WithAPIKey("key"))
	_, _, err := c.Incidents.PostUpdate(context.Background(), "bad-id", statuscast.UpdateIncidentRequest{Message: "m"})
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestIncidentsResolve_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/incident", jsonHandler(200, incidentResolvedJSON))
	c := newMockClient(t, mux)

	inc, resp, err := c.Incidents.Resolve(context.Background(), "100", statuscast.ResolveRequest{
		Message: "Issue resolved.",
		Notify:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if inc.Status != statuscast.StatusResolved {
		t.Errorf("Status = %q; want %q", inc.Status, statuscast.StatusResolved)
	}
	if inc.ResolvedAt == nil {
		t.Error("ResolvedAt should be set for a resolved incident")
	}
}
