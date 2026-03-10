package statuscast_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	statuscast "statuscast-go"
)

func TestReportsUptime_Success(t *testing.T) {
	const body = `{"uptime":99.97,"start":"2024-01-01T00:00:00Z","end":"2024-01-31T23:59:59Z"}`
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components/uptime", jsonHandler(200, body))
	c := newMockClient(t, mux)

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	reports, resp, err := c.Reports.Uptime(context.Background(), since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(reports) != 1 {
		t.Fatalf("len(reports) = %d; want 1", len(reports))
	}
	if reports[0].Uptime != 99.97 {
		t.Errorf("Uptime = %f; want 99.97", reports[0].Uptime)
	}
	if reports[0].WindowStart.IsZero() {
		t.Error("WindowStart should not be zero")
	}
	if reports[0].WindowEnd.IsZero() {
		t.Error("WindowEnd should not be zero")
	}
}

func TestReportsUptime_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/components/uptime", statusHandler(401))
	c := newMockClient(t, mux)

	since := time.Now().Add(-30 * 24 * time.Hour)
	until := time.Now()

	_, _, err := c.Reports.Uptime(context.Background(), since, until)
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}
