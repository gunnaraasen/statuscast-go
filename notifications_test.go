package statuscast_test

import (
	"context"
	"net/http"
	"testing"

	statuscast "statuscast-go"
)

const templateJSON = `{"id":300,"subject":"Incident Alert","contents":"Hello {{.IncidentTitle}}"}`
const templateListJSON = `[{"id":300,"subject":"Incident Alert","contents":"Hello {{.IncidentTitle}}"},{"id":301,"subject":"Resolved","contents":"All clear"}]`

func TestNotificationsCreateTemplate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/contenttemplate", jsonHandler(200, templateJSON))
	c := newMockClient(t, mux)

	tmpl, resp, err := c.Notifications.CreateTemplate(context.Background(), statuscast.NotificationTemplate{
		Subject: "Incident Alert",
		Body:    "Hello {{.IncidentTitle}}",
		Channel: statuscast.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if tmpl.ID != "300" {
		t.Errorf("ID = %q; want %q", tmpl.ID, "300")
	}
	if tmpl.Subject != "Incident Alert" {
		t.Errorf("Subject = %q; want %q", tmpl.Subject, "Incident Alert")
	}
	if tmpl.Body != "Hello {{.IncidentTitle}}" {
		t.Errorf("Body = %q; want %q", tmpl.Body, "Hello {{.IncidentTitle}}")
	}
}

func TestNotificationsCreateTemplate_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/contenttemplate", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Notifications.CreateTemplate(context.Background(), statuscast.NotificationTemplate{
		Subject: "Test",
	})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestNotificationsUpdateTemplate_Success(t *testing.T) {
	const updatedJSON = `{"id":300,"subject":"Updated Alert","contents":"New body"}`
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v4/contenttemplate", jsonHandler(200, updatedJSON))
	c := newMockClient(t, mux)

	tmpl, resp, err := c.Notifications.UpdateTemplate(context.Background(), "300", statuscast.NotificationTemplate{
		Subject: "Updated Alert",
		Body:    "New body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if tmpl.Subject != "Updated Alert" {
		t.Errorf("Subject = %q; want %q", tmpl.Subject, "Updated Alert")
	}
}

func TestNotificationsUpdateTemplate_InvalidID(t *testing.T) {
	c, _ := statuscast.New(statuscast.WithAPIKey("key"))
	_, _, err := c.Notifications.UpdateTemplate(context.Background(), "bad-id", statuscast.NotificationTemplate{})
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestNotificationsListTemplates_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/contenttemplate", jsonHandler(200, templateListJSON))
	c := newMockClient(t, mux)

	result, resp, err := c.Notifications.ListTemplates(context.Background(), statuscast.Pagination{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d; want 2", len(result.Items))
	}
	if result.Items[0].ID != "300" {
		t.Errorf("Items[0].ID = %q; want %q", result.Items[0].ID, "300")
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d; want 2", result.TotalCount)
	}
}

func TestNotificationsListTemplates_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/contenttemplate", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Notifications.ListTemplates(context.Background(), statuscast.Pagination{})
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}
