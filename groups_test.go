package statuscast_test

import (
	"context"
	"net/http"
	"testing"

	statuscast "statuscast-go"
)

const groupListJSON = `{
	"items": [
		{"id":10,"name":"Engineering","dateCreated":"2024-01-01T00:00:00Z"},
		{"id":11,"name":"Operations","dateCreated":"2024-01-02T00:00:00Z"}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

func TestGroupsList_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/groups", jsonHandler(200, groupListJSON))
	c := newMockClient(t, mux)

	result, resp, err := c.Groups.List(context.Background(), statuscast.Pagination{})
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
	if result.Items[0].ID != "10" {
		t.Errorf("Items[0].ID = %q; want %q", result.Items[0].ID, "10")
	}
	if result.Items[0].Name != "Engineering" {
		t.Errorf("Items[0].Name = %q; want %q", result.Items[0].Name, "Engineering")
	}
}

func TestGroupsList_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/groups", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Groups.List(context.Background(), statuscast.Pagination{})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGroupsList_WithPagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/groups", jsonHandler(200, `{"items":[],"totalItems":100,"page":3,"pages":10}`))
	c := newMockClient(t, mux)

	result, _, err := c.Groups.List(context.Background(), statuscast.Pagination{Page: 3, PerPage: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 3 {
		t.Errorf("Page = %d; want 3", result.Page)
	}
	if result.TotalPages != 10 {
		t.Errorf("TotalPages = %d; want 10", result.TotalPages)
	}
}
