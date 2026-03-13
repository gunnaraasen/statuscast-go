package statuscast_test

import (
	"context"
	"net/http"
	"testing"

	statuscast "statuscast-go"
)

const componentJSON = `{"id":42,"name":"API Server","description":"Primary API","status":"Available"}`
const componentDegradedJSON = `{"id":42,"name":"API Server","status":"DegradedPerformance"}`

func TestComponentsGet_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/component/{id}", jsonHandler(200, componentJSON))
	c := newMockClient(t, mux)

	comp, resp, err := c.Components.Get(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if comp.ID != "42" {
		t.Errorf("ID = %q; want %q", comp.ID, "42")
	}
	if comp.Name != "API Server" {
		t.Errorf("Name = %q; want %q", comp.Name, "API Server")
	}
	if comp.Description != "Primary API" {
		t.Errorf("Description = %q; want %q", comp.Description, "Primary API")
	}
	if comp.Status != statuscast.StatusOperational {
		t.Errorf("Status = %q; want %q", comp.Status, statuscast.StatusOperational)
	}
}

func TestComponentsGet_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/component/{id}", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Components.Get(context.Background(), "42")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestComponentsGet_InvalidID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Components.Get(context.Background(), "not-a-number")
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestComponentsList_Success(t *testing.T) {
	const body = `[{"id":42,"name":"API Server","status":"Available"},{"id":43,"name":"DB","status":"DegradedPerformance"}]`
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", jsonHandler(200, body))
	c := newMockClient(t, mux)

	result, resp, err := c.Components.List(context.Background(), "", statuscast.Pagination{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d; want 2", len(result.Items))
	}
	if result.Items[0].ID != "42" {
		t.Errorf("Items[0].ID = %q; want %q", result.Items[0].ID, "42")
	}
	if result.Items[1].Status != statuscast.StatusDegradedPerf {
		t.Errorf("Items[1].Status = %q; want %q", result.Items[1].Status, statuscast.StatusDegradedPerf)
	}
}

func TestComponentsList_WithParentFilter(t *testing.T) {
	// The API returns two items; one has parentId=42, one does not.
	const body = `[{"id":43,"name":"Sub Component","status":"Available","parentId":42},{"id":44,"name":"Root","status":"Available"}]`
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", jsonHandler(200, body))
	c := newMockClient(t, mux)

	// Filter by parentID "42" — should return only the sub-component.
	result, _, err := c.Components.List(context.Background(), "42", statuscast.Pagination{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(Items) = %d; want 1 (filtered by parentID)", len(result.Items))
	}
	if result.Items[0].ID != "43" {
		t.Errorf("filtered item ID = %q; want %q", result.Items[0].ID, "43")
	}
}

func TestComponentsList_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Components.List(context.Background(), "", statuscast.Pagination{})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestComponentsCreate_Success(t *testing.T) {
	const body = `{"id":99,"name":"New Component","status":"Available"}`
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/component", jsonHandler(200, body))
	c := newMockClient(t, mux)

	comp, resp, err := c.Components.Create(context.Background(), statuscast.CreateComponentRequest{
		Name: "New Component",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if comp.ID != "99" {
		t.Errorf("ID = %q; want %q", comp.ID, "99")
	}
	if comp.Name != "New Component" {
		t.Errorf("Name = %q; want %q", comp.Name, "New Component")
	}
}

func TestComponentsCreate_InvalidParentID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Components.Create(context.Background(), statuscast.CreateComponentRequest{
		Name:     "Sub",
		ParentID: "not-a-number",
	})
	if err == nil {
		t.Fatal("expected error for invalid parent ID, got nil")
	}
}

func TestComponentsCreate_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/component", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Components.Create(context.Background(), statuscast.CreateComponentRequest{Name: "X"})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestComponentsUpdate_Success(t *testing.T) {
	const body = `{"id":42,"name":"Updated Name","status":"DegradedPerformance"}`
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/component", jsonHandler(200, body))
	c := newMockClient(t, mux)

	newName := "Updated Name"
	comp, _, err := c.Components.Update(context.Background(), "42", statuscast.UpdateComponentRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp.Name != "Updated Name" {
		t.Errorf("Name = %q; want %q", comp.Name, "Updated Name")
	}
	if comp.Status != statuscast.StatusDegradedPerf {
		t.Errorf("Status = %q; want %q", comp.Status, statuscast.StatusDegradedPerf)
	}
}

func TestComponentsUpdate_InvalidID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Components.Update(context.Background(), "bad-id", statuscast.UpdateComponentRequest{})
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestComponentsDelete_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/component", statusHandler(200))
	c := newMockClient(t, mux)

	resp, err := c.Components.Delete(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestComponentsDelete_InvalidID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, err := c.Components.Delete(context.Background(), "bad-id")
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestComponentsDelete_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/component", statusHandler(401))
	c := newMockClient(t, mux)

	_, err := c.Components.Delete(context.Background(), "42")
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestComponentsSetStatus(t *testing.T) {
	// SetStatus is a convenience wrapper for Update; verify it works end-to-end.
	const body = `{"id":42,"name":"API Server","status":"Maintenance"}`
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/component", jsonHandler(200, body))
	c := newMockClient(t, mux)

	comp, _, err := c.Components.SetStatus(context.Background(), "42", statuscast.StatusUnderMaintenance)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp.Status != statuscast.StatusUnderMaintenance {
		t.Errorf("Status = %q; want %q", comp.Status, statuscast.StatusUnderMaintenance)
	}
}
