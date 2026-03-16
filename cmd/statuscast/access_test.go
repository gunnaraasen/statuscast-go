package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

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

const testUUID = "123e4567-e89b-12d3-a456-426614174000"

func TestAccessUsersList_Table(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/users", jsonBody(200, userListJSON))

	out, err := runCmd(t, mux, "access", "users", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "admin@example.com") {
		t.Errorf("output missing email; got:\n%s", out)
	}
	if !strings.Contains(out, "manager@example.com") {
		t.Errorf("output missing second user; got:\n%s", out)
	}
}

func TestAccessUsersList_JSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/users", jsonBody(200, userListJSON))

	out, err := runCmd(t, mux, "--json", "access", "users", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if result["total_count"].(float64) != 2 {
		t.Errorf("total_count = %v; want 2", result["total_count"])
	}
}

func TestAccessUsersList_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/users", statusOnly(401))

	_, err := runCmd(t, mux, "access", "users", "list")
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestAccessUsersInvite_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/user", jsonBody(200, userJSON))

	out, err := runCmd(t, mux,
		"access", "users", "invite",
		"--email", "admin@example.com",
		"--name", "Admin User",
		"--role", "administrator",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, testUUID) {
		t.Errorf("output missing user UUID; got:\n%s", out)
	}
}

func TestAccessUsersInvite_MissingEmail(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "access", "users", "invite", "--name", "User", "--role", "employee")
	if err == nil {
		t.Fatal("expected error for missing --email, got nil")
	}
}

func TestAccessUsersInvite_MissingRole(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "access", "users", "invite", "--email", "x@x.com", "--name", "X")
	if err == nil {
		t.Fatal("expected error for missing --role, got nil")
	}
}

func TestAccessUsersUpdateRole_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/user", jsonBody(200, `{"id":500,"email":"admin@example.com","fullName":"Admin User"}`))

	out, err := runCmd(t, mux,
		"access", "users", "update-role", "500",
		"--role", "manager",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "500") {
		t.Errorf("output missing user id; got:\n%s", out)
	}
}

func TestAccessUsersUpdateRole_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "access", "users", "update-role", "--role", "manager")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestAccessUsersUpdateRole_MissingRole(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "access", "users", "update-role", "500")
	if err == nil {
		t.Fatal("expected error for missing --role, got nil")
	}
}

func TestAccessUsersRemove_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/user", statusOnly(200))

	out, err := runCmd(t, mux, "access", "users", "remove", testUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, testUUID) {
		t.Errorf("output missing user UUID; got:\n%s", out)
	}
}

func TestAccessUsersRemove_InvalidUUID(t *testing.T) {
	// remove requires a UUID, not an integer.
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "access", "users", "remove", "500")
	if err == nil {
		t.Fatal("expected error for non-UUID id, got nil")
	}
}

func TestAccessUsersRemove_MissingID(t *testing.T) {
	mux := http.NewServeMux()
	_, err := runCmd(t, mux, "access", "users", "remove")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}
