package statuscast_test

import (
	"context"
	"testing"
	"time"

	statuscast "statuscast-go"
)

// unsupportedErr verifies that err is non-nil and contains "not supported".
func assertNotSupported(t *testing.T, name string, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error, got nil", name)
		return
	}
	msg := err.Error()
	if msg != "not supported by StatusCast API v4" {
		t.Errorf("%s: error = %q; want %q", name, msg, "not supported by StatusCast API v4")
	}
}

func newTestClient(t *testing.T) *statuscast.Client {
	t.Helper()
	c, err := statuscast.New(statuscast.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

func TestTeamsList_NotSupported(t *testing.T) {
	c := newTestClient(t)
	_, _, err := c.Teams.List(context.Background(), statuscast.Pagination{})
	assertNotSupported(t, "Teams.List", err)
}

func TestIncidentsFileRCA_NotSupported(t *testing.T) {
	c := newTestClient(t)
	_, _, err := c.Incidents.FileRCA(context.Background(), "100", statuscast.RCARequest{
		Summary:   "Summary",
		RootCause: "Root cause",
	})
	assertNotSupported(t, "Incidents.FileRCA", err)
}

func TestSubscribersBulkImport_NotSupported(t *testing.T) {
	c := newTestClient(t)
	_, _, err := c.Subscribers.BulkImport(context.Background(), []byte("email\ntest@example.com"))
	assertNotSupported(t, "Subscribers.BulkImport", err)
}

func TestNotificationsGetLog_NotSupported(t *testing.T) {
	c := newTestClient(t)
	_, _, err := c.Notifications.GetLog(context.Background(), "log-id-123")
	assertNotSupported(t, "Notifications.GetLog", err)
}

func TestNotificationsListLogs_NotSupported(t *testing.T) {
	c := newTestClient(t)
	_, _, err := c.Notifications.ListLogs(context.Background(), "100", statuscast.Pagination{})
	assertNotSupported(t, "Notifications.ListLogs", err)
}

func TestReportsIncidentSummary_NotSupported(t *testing.T) {
	c := newTestClient(t)
	since := time.Now().Add(-30 * 24 * time.Hour)
	until := time.Now()
	_, _, err := c.Reports.IncidentSummary(context.Background(), since, until)
	assertNotSupported(t, "Reports.IncidentSummary", err)
}

func TestReportsListRCAs_NotSupported(t *testing.T) {
	c := newTestClient(t)
	_, _, err := c.Reports.ListRCAs(context.Background(), statuscast.Pagination{})
	assertNotSupported(t, "Reports.ListRCAs", err)
}

func TestAccessSetPageVisibility_NotSupported(t *testing.T) {
	c := newTestClient(t)
	_, err := c.Access.SetPageVisibility(context.Background(), statuscast.VisibilityPublic)
	assertNotSupported(t, "Access.SetPageVisibility", err)
}
