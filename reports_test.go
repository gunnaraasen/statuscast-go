package statuscast_test

import (
	"context"
	"net/http"
	"sync/atomic"
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

// summaryIncidentsJSON contains two incidents:
//   - Incident 1: startDate before dateCreated (MTTD=30m), endDate set (MTTR=2h), affects component 10
//   - Incident 2: startDate == dateCreated (no MTTD contribution), endDate set (MTTR=1h), affects component 10
const summaryIncidentsJSON = `{
	"items": [
		{
			"id": 1,
			"dateCreated": "2024-01-15T10:30:00Z",
			"startDate": "2024-01-15T10:00:00Z",
			"endDate": "2024-01-15T12:30:00Z",
			"affectedComponents": [{"componentId": 10}]
		},
		{
			"id": 2,
			"dateCreated": "2024-01-20T08:00:00Z",
			"startDate": "2024-01-20T08:00:00Z",
			"endDate": "2024-01-20T09:00:00Z",
			"affectedComponents": [{"componentId": 10}]
		}
	],
	"totalItems": 2,
	"page": 1,
	"pages": 1
}`

func TestReportsIncidentSummary_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonHandler(200, summaryIncidentsJSON))
	c := newMockClient(t, mux)

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	report, resp, err := c.Reports.IncidentSummary(context.Background(), since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if report.TotalIncidents != 2 {
		t.Errorf("TotalIncidents = %d; want 2", report.TotalIncidents)
	}
	// Only incident 1 has startDate before dateCreated, so MTTD = 30m.
	wantMTTD := 30 * time.Minute
	if report.MeanTimeToDetect != wantMTTD {
		t.Errorf("MeanTimeToDetect = %v; want %v", report.MeanTimeToDetect, wantMTTD)
	}
	// Both incidents resolved: incident 1 = 2h, incident 2 = 1h → mean = 1.5h.
	wantMTTR := 90 * time.Minute
	if report.MeanTimeToResolve != wantMTTR {
		t.Errorf("MeanTimeToResolve = %v; want %v", report.MeanTimeToResolve, wantMTTR)
	}
	if report.ByComponent["10"] != 2 {
		t.Errorf("ByComponent[\"10\"] = %d; want 2", report.ByComponent["10"])
	}
	if !report.Since.Equal(since) {
		t.Errorf("Since = %v; want %v", report.Since, since)
	}
}

func TestReportsIncidentSummary_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", jsonHandler(200, `{"items":[],"totalItems":0,"page":1,"pages":1}`))
	c := newMockClient(t, mux)

	report, _, err := c.Reports.IncidentSummary(context.Background(), time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalIncidents != 0 {
		t.Errorf("TotalIncidents = %d; want 0", report.TotalIncidents)
	}
	if report.MeanTimeToDetect != 0 {
		t.Errorf("MeanTimeToDetect should be zero for empty result")
	}
}

func TestReportsIncidentSummary_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", statusHandler(401))
	c := newMockClient(t, mux)

	_, _, err := c.Reports.IncidentSummary(context.Background(), time.Now().Add(-24*time.Hour), time.Now())
	if err != statuscast.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestReportsIncidentSummary_Paginated(t *testing.T) {
	// Page 1 returns incident 1 (component 10); page 2 returns incident 2 (component 20).
	const page1JSON = `{
		"items": [{"id":1,"dateCreated":"2024-01-15T10:30:00Z","startDate":"2024-01-15T10:00:00Z","endDate":"2024-01-15T12:30:00Z","affectedComponents":[{"componentId":10}]}],
		"totalItems":2,"page":1,"pages":2
	}`
	const page2JSON = `{
		"items": [{"id":2,"dateCreated":"2024-01-20T08:00:00Z","startDate":"2024-01-20T08:00:00Z","endDate":"2024-01-20T09:00:00Z","affectedComponents":[{"componentId":20}]}],
		"totalItems":2,"page":2,"pages":2
	}`
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v4/incidents", func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			jsonHandler(200, page1JSON)(w, r)
		} else {
			jsonHandler(200, page2JSON)(w, r)
		}
	})
	c := newMockClient(t, mux)

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	report, _, err := c.Reports.IncidentSummary(context.Background(), since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("API calls = %d; want 2", calls.Load())
	}
	if report.TotalIncidents != 2 {
		t.Errorf("TotalIncidents = %d; want 2", report.TotalIncidents)
	}
	// Both components present in ByComponent.
	if report.ByComponent["10"] != 1 {
		t.Errorf("ByComponent[\"10\"] = %d; want 1", report.ByComponent["10"])
	}
	if report.ByComponent["20"] != 1 {
		t.Errorf("ByComponent[\"20\"] = %d; want 1", report.ByComponent["20"])
	}
}
