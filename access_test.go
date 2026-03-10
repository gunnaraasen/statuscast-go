package statuscast_test

import (
	"context"
	"net/http"
	"testing"

	statuscast "statuscast-go"
)

// A valid test UUID used as user ID throughout these tests.
const testUserUUID = "123e4567-e89b-12d3-a456-426614174000"

const userJSON = `{"id":500,"userId":"123e4567-e89b-12d3-a456-426614174000","email":"admin@example.com","fullName":"Admin User"}`
const userListJSON = `{
	"items": [
		{"id":500,"userId":"123e4567-e89b-12d3-a456-426614174000","email":"admin@example.com","fullName":"Admin User"},
		{"id":501,"email":"manager@example.com","fullName":"Manager User"}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

func TestAccessInviteUser_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/user", jsonHandler(200, userJSON))
	c := newMockClient(t, mux)

	user, resp, err := c.Access.InviteUser(context.Background(), statuscast.InviteUserRequest{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  statuscast.RoleAdministrator,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	// When userId is set, it is used as the ID.
	if user.ID != testUserUUID {
		t.Errorf("ID = %q; want %q", user.ID, testUserUUID)
	}
	if user.Email != "admin@example.com" {
		t.Errorf("Email = %q; want %q", user.Email, "admin@example.com")
	}
	if user.Name != "Admin User" {
		t.Errorf("Name = %q; want %q", user.Name, "Admin User")
	}
}

func TestAccessInviteUser_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/user", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Access.InviteUser(context.Background(), statuscast.InviteUserRequest{
		Email: "x@x.com",
		Role:  statuscast.RoleEmployee,
	})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAccessUpdateRole_Success(t *testing.T) {
	const updatedUserJSON = `{"id":500,"email":"admin@example.com","fullName":"Admin User"}`
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/user", jsonHandler(200, updatedUserJSON))
	c := newMockClient(t, mux)

	user, resp, err := c.Access.UpdateRole(context.Background(), "500", statuscast.RoleManager)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	// When userId is not set, falls back to int32 id.
	if user.ID != "500" {
		t.Errorf("ID = %q; want %q", user.ID, "500")
	}
}

func TestAccessUpdateRole_InvalidID(t *testing.T) {
	c, _ := statuscast.New(statuscast.WithAPIKey("key"))
	_, _, err := c.Access.UpdateRole(context.Background(), "not-a-number", statuscast.RoleManager)
	if err == nil {
		t.Fatal("expected error for non-numeric user ID, got nil")
	}
}

func TestAccessUpdateRole_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/user", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Access.UpdateRole(context.Background(), "500", statuscast.RoleEmployee)
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAccessRemoveUser_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/user", statusHandler(200))
	c := newMockClient(t, mux)

	resp, err := c.Access.RemoveUser(context.Background(), testUserUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestAccessRemoveUser_InvalidUUID(t *testing.T) {
	// RemoveUser requires a UUID; an integer ID should fail.
	c, _ := statuscast.New(statuscast.WithAPIKey("key"))
	_, err := c.Access.RemoveUser(context.Background(), "500")
	if err == nil {
		t.Fatal("expected error for non-UUID user ID, got nil")
	}
}

func TestAccessRemoveUser_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/user", statusHandler(401))
	c := newMockClient(t, mux)

	_, err := c.Access.RemoveUser(context.Background(), testUserUUID)
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAccessListUsers_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/users", jsonHandler(200, userListJSON))
	c := newMockClient(t, mux)

	result, resp, err := c.Access.ListUsers(context.Background(), statuscast.Pagination{})
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
	// First item has UUID.
	if result.Items[0].ID != testUserUUID {
		t.Errorf("Items[0].ID = %q; want %q", result.Items[0].ID, testUserUUID)
	}
	// Second item has no UUID; falls back to int32 id.
	if result.Items[1].ID != "501" {
		t.Errorf("Items[1].ID = %q; want %q", result.Items[1].ID, "501")
	}
}

func TestAccessListUsers_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/users", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Access.ListUsers(context.Background(), statuscast.Pagination{})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}
