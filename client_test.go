package statuscast_test

import (
	"testing"

	statuscast "statuscast-go"
)

func TestNew_MissingAPIKey(t *testing.T) {
	_, err := statuscast.New()
	if err == nil {
		t.Fatal("expected error when API key is missing")
	}
	if err != statuscast.ErrMissingAPIKey {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestNew_ValidAPIKey(t *testing.T) {
	client, err := statuscast.New(statuscast.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Components == nil {
		t.Error("Components sub-client is nil")
	}
	if client.Incidents == nil {
		t.Error("Incidents sub-client is nil")
	}
	if client.Subscribers == nil {
		t.Error("Subscribers sub-client is nil")
	}
	if client.Groups == nil {
		t.Error("Groups sub-client is nil")
	}
	if client.Notifications == nil {
		t.Error("Notifications sub-client is nil")
	}
	if client.Reports == nil {
		t.Error("Reports sub-client is nil")
	}
	if client.Access == nil {
		t.Error("Access sub-client is nil")
	}
}

func TestNew_CustomBaseURL(t *testing.T) {
	client, err := statuscast.New(
		statuscast.WithAPIKey("test-key"),
		statuscast.WithBaseURL("https://custom.example.com"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
