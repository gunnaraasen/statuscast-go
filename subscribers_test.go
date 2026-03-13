package statuscast_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	statuscast "statuscast-go"
)

const subscriberJSON = `{"id":200,"emailAddress":"test@example.com","subscribeToIncidentPosts":true}`
const subscriberWithGroupsJSON = `{
	"id": 200,
	"emailAddress": "test@example.com",
	"components": [{"id":42},{"id":43}],
	"groups": [{"id":10}]
}`
const subscriberListJSON = `{
	"items": [
		{"id":200,"emailAddress":"alice@example.com"},
		{"id":201,"emailAddress":"bob@example.com"}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

func TestSubscribersAdd_DefaultEmailChannel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", jsonHandler(200, subscriberJSON))
	c := newMockClient(t, mux)

	// No channels specified → defaults to email only.
	sub, resp, err := c.Subscribers.Add(context.Background(), statuscast.AddSubscriberRequest{
		Email: "test@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if sub.ID != "200" {
		t.Errorf("ID = %q; want %q", sub.ID, "200")
	}
	if sub.Email != "test@example.com" {
		t.Errorf("Email = %q; want %q", sub.Email, "test@example.com")
	}
	// subscribeToIncidentPosts=true in JSON → ChannelEmail in Channels.
	if len(sub.Channels) == 0 {
		t.Error("Channels should not be empty when subscribeToIncidentPosts=true")
	}
	found := false
	for _, ch := range sub.Channels {
		if ch == statuscast.ChannelEmail {
			found = true
		}
	}
	if !found {
		t.Errorf("ChannelEmail not found in Channels: %v", sub.Channels)
	}
}

func TestSubscribersAdd_WithExplicitChannels(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", jsonHandler(200, subscriberJSON))
	c := newMockClient(t, mux)

	_, _, err := c.Subscribers.Add(context.Background(), statuscast.AddSubscriberRequest{
		Email:    "test@example.com",
		Phone:    "+15551234567",
		Channels: []statuscast.NotificationChannel{statuscast.ChannelEmail, statuscast.ChannelSMS},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubscribersAdd_WithGroupsAndComponents(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", jsonHandler(200, subscriberJSON))
	c := newMockClient(t, mux)

	_, _, err := c.Subscribers.Add(context.Background(), statuscast.AddSubscriberRequest{
		Email:      "test@example.com",
		Groups:     []string{"10", "11"},
		Components: []string{"42"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubscribersAdd_InvalidGroupID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Subscribers.Add(context.Background(), statuscast.AddSubscriberRequest{
		Email:  "test@example.com",
		Groups: []string{"not-a-number"},
	})
	if err == nil {
		t.Fatal("expected error for invalid group ID, got nil")
	}
}

func TestSubscribersAdd_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Subscribers.Add(context.Background(), statuscast.AddSubscriberRequest{Email: "x@x.com"})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestSubscribersGet_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/subscriber/{id}", jsonHandler(200, subscriberWithGroupsJSON))
	c := newMockClient(t, mux)

	sub, resp, err := c.Subscribers.Get(context.Background(), "200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if sub.ID != "200" {
		t.Errorf("ID = %q; want %q", sub.ID, "200")
	}
	if len(sub.Components) != 2 {
		t.Errorf("len(Components) = %d; want 2", len(sub.Components))
	}
	if len(sub.Groups) != 1 {
		t.Errorf("len(Groups) = %d; want 1", len(sub.Groups))
	}
}

func TestSubscribersGet_InvalidID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Subscribers.Get(context.Background(), "bad-id")
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestSubscribersGet_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/subscriber/{id}", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Subscribers.Get(context.Background(), "200")
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestSubscribersList_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscribers/search", jsonHandler(200, subscriberListJSON))
	c := newMockClient(t, mux)

	result, resp, err := c.Subscribers.List(context.Background(), "", statuscast.Pagination{})
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
	if result.Items[0].Email != "alice@example.com" {
		t.Errorf("Items[0].Email = %q; want %q", result.Items[0].Email, "alice@example.com")
	}
}

func TestSubscribersList_WithGroupFilter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscribers/search", jsonHandler(200, `{"items":[{"id":200,"emailAddress":"alice@example.com"}],"totalItems":1,"page":1,"pages":1}`))
	c := newMockClient(t, mux)

	result, _, err := c.Subscribers.List(context.Background(), "10", statuscast.Pagination{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(Items) = %d; want 1", len(result.Items))
	}
}

func TestSubscribersList_InvalidGroupID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Subscribers.List(context.Background(), "not-a-number", statuscast.Pagination{})
	if err == nil {
		t.Fatal("expected error for invalid group ID, got nil")
	}
}

func TestSubscribersUpdate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/subscriber", jsonHandler(200, subscriberWithGroupsJSON))
	c := newMockClient(t, mux)

	sub, resp, err := c.Subscribers.Update(context.Background(), "200", statuscast.UpdateSubscriberRequest{
		Channels: []statuscast.NotificationChannel{statuscast.ChannelEmail},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if sub.ID != "200" {
		t.Errorf("ID = %q; want %q", sub.ID, "200")
	}
}

func TestSubscribersUpdate_InvalidID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Subscribers.Update(context.Background(), "bad-id", statuscast.UpdateSubscriberRequest{})
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestSubscribersRemove_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/subscriber/{id}", statusHandler(200))
	c := newMockClient(t, mux)

	resp, err := c.Subscribers.Remove(context.Background(), "200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestSubscribersRemove_InvalidID(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, err := c.Subscribers.Remove(context.Background(), "bad-id")
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestSubscribersRemove_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v4/subscriber/{id}", statusHandler(401))
	c := newMockClient(t, mux)

	_, err := c.Subscribers.Remove(context.Background(), "200")
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestSubscribersBulkImport_AllSucceed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", jsonHandler(200, subscriberJSON))
	c := newMockClient(t, mux)

	csv := "email\nalice@example.com\nbob@example.com\n"
	result, resp, err := c.Subscribers.BulkImport(context.Background(), []byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if result.Imported != 2 {
		t.Errorf("Imported = %d; want 2", result.Imported)
	}
	if result.Failed != 0 {
		t.Errorf("Failed = %d; want 0", result.Failed)
	}
	if result.Skipped != 0 {
		t.Errorf("Skipped = %d; want 0", result.Skipped)
	}
}

func TestSubscribersBulkImport_SkipsEmptyEmail(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", jsonHandler(200, subscriberJSON))
	c := newMockClient(t, mux)

	// Row with whitespace-only email should be skipped, not imported.
	csv := "email\nalice@example.com\n   \nbob@example.com\n"
	result, _, err := c.Subscribers.BulkImport(context.Background(), []byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Imported != 2 {
		t.Errorf("Imported = %d; want 2", result.Imported)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d; want 1", result.Skipped)
	}
}

func TestSubscribersBulkImport_PartialFailure(t *testing.T) {
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/subscriber", func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 2 {
			w.WriteHeader(401)
			return
		}
		jsonHandler(200, subscriberJSON)(w, r)
	})
	c := newMockClient(t, mux)

	csv := "email\nalice@example.com\nbob@example.com\n"
	result, _, err := c.Subscribers.BulkImport(context.Background(), []byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Imported != 1 {
		t.Errorf("Imported = %d; want 1", result.Imported)
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d; want 1", result.Failed)
	}
	if len(result.Errors) != 1 {
		t.Errorf("len(Errors) = %d; want 1", len(result.Errors))
	}
}

func TestSubscribersBulkImport_MissingEmailColumn(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Subscribers.BulkImport(context.Background(), []byte("name,phone\nAlice,+1555\n"))
	if err == nil {
		t.Fatal("expected error for missing email column, got nil")
	}
}

func TestSubscribersBulkImport_EmptyCSV(t *testing.T) {
	c := mustNew(t, statuscast.WithAPIKey("key"))
	_, _, err := c.Subscribers.BulkImport(context.Background(), []byte(""))
	if err == nil {
		t.Fatal("expected error for empty CSV, got nil")
	}
}
