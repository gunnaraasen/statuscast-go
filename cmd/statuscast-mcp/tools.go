package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	statuscast "statuscast-go"
)

func registerTools(s *mcp.Server, client *statuscast.Client) {
	// ─── list_components ─────────────────────────────────────────────────────

	type listComponentsArgs struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_components",
		Description: "List all Statuscast components and their current operational status.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listComponentsArgs) (*mcp.CallToolResult, any, error) {
		result, _, err := client.Components.List(ctx, "", statuscast.Pagination{})
		if err != nil {
			return nil, nil, err
		}
		if len(result.Items) == 0 {
			return textResult("No components found."), nil, nil
		}
		var sb strings.Builder
		for _, c := range result.Items {
			fmt.Fprintf(&sb, "ID: %s | Name: %s | Status: %s\n", c.ID, c.Name, c.Status)
		}
		return textResult(sb.String()), nil, nil
	})

	// ─── set_component_status ─────────────────────────────────────────────────

	type setComponentStatusArgs struct {
		ComponentID string `json:"component_id" jsonschema:"The ID of the component to update"`
		Status      string `json:"status"       jsonschema:"New status: operational, degraded_performance, partial_outage, major_outage, or under_maintenance"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "set_component_status",
		Description: "Set the operational status of a Statuscast component.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args setComponentStatusArgs) (*mcp.CallToolResult, any, error) {
		comp, _, err := client.Components.SetStatus(ctx, args.ComponentID, statuscast.ComponentStatus(args.Status))
		if err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Component %q (ID: %s) status set to %s.", comp.Name, comp.ID, comp.Status)), nil, nil
	})

	// ─── list_incidents ───────────────────────────────────────────────────────

	type listIncidentsArgs struct {
		ActiveOnly *bool `json:"active_only" jsonschema:"If true, return only open/active incidents"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_incidents",
		Description: "List Statuscast incidents. Optionally filter to active incidents only.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args listIncidentsArgs) (*mcp.CallToolResult, any, error) {
		filter := statuscast.IncidentFilter{}
		if args.ActiveOnly != nil {
			filter.ActiveOnly = *args.ActiveOnly
		}
		result, _, err := client.Incidents.List(ctx, filter, statuscast.Pagination{})
		if err != nil {
			return nil, nil, err
		}
		if len(result.Items) == 0 {
			return textResult("No incidents found."), nil, nil
		}
		var sb strings.Builder
		for _, inc := range result.Items {
			state := "open"
			if inc.ResolvedAt != nil {
				state = "resolved " + inc.ResolvedAt.Format(time.RFC3339)
			}
			fmt.Fprintf(&sb, "ID: %s | Title: %s | Status: %s | %s | Created: %s\n",
				inc.ID, inc.Title, inc.Status, state, inc.CreatedAt.Format(time.RFC3339))
		}
		return textResult(sb.String()), nil, nil
	})

	// ─── create_incident ──────────────────────────────────────────────────────

	type createIncidentArgs struct {
		Title      string   `json:"title"                jsonschema:"Incident title"`
		Message    string   `json:"message"              jsonschema:"Initial update body (supports Markdown)"`
		Components []string `json:"components,omitempty" jsonschema:"Component IDs to mark as affected"`
		Status     *string  `json:"status,omitempty"     jsonschema:"Incident status: investigating (default), identified, monitoring, resolved"`
		PostType   *string  `json:"post_type,omitempty"  jsonschema:"Post type: outage (default), maintenance, info"`
		Notify     *bool    `json:"notify,omitempty"     jsonschema:"Send subscriber notifications immediately"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_incident",
		Description: "Open a new Statuscast incident with an initial update message.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args createIncidentArgs) (*mcp.CallToolResult, any, error) {
		req := statuscast.CreateIncidentRequest{
			Title:      args.Title,
			Message:    args.Message,
			Components: args.Components,
		}
		if args.Status != nil {
			req.Status = statuscast.IncidentStatus(*args.Status)
		}
		if args.PostType != nil {
			req.PostType = statuscast.IncidentPostType(*args.PostType)
		}
		if args.Notify != nil {
			req.Notify = *args.Notify
		}
		inc, _, err := client.Incidents.Create(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Incident created: ID=%s, Title=%q, Status=%s", inc.ID, inc.Title, inc.Status)), nil, nil
	})

	// ─── update_incident ──────────────────────────────────────────────────────

	type updateIncidentArgs struct {
		IncidentID string  `json:"incident_id"         jsonschema:"The ID of the incident to update"`
		Message    string  `json:"message"             jsonschema:"Update message body (supports Markdown)"`
		Status     *string `json:"status,omitempty"    jsonschema:"New status: investigating, identified, monitoring, resolved"`
		Notify     *bool   `json:"notify,omitempty"    jsonschema:"Send subscriber notifications for this update"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_incident",
		Description: "Post a new timeline update to an existing Statuscast incident.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args updateIncidentArgs) (*mcp.CallToolResult, any, error) {
		req := statuscast.UpdateIncidentRequest{
			Message: args.Message,
		}
		if args.Status != nil {
			req.Status = statuscast.IncidentStatus(*args.Status)
		}
		if args.Notify != nil {
			req.Notify = *args.Notify
		}
		update, _, err := client.Incidents.PostUpdate(ctx, args.IncidentID, req)
		if err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Update posted to incident %s: status=%s, message=%q",
			args.IncidentID, update.Status, update.Message)), nil, nil
	})

	// ─── resolve_incident ─────────────────────────────────────────────────────

	type resolveIncidentArgs struct {
		IncidentID string `json:"incident_id"      jsonschema:"The ID of the incident to resolve"`
		Message    string `json:"message"          jsonschema:"Final resolution message shown to subscribers"`
		Notify     *bool  `json:"notify,omitempty" jsonschema:"Send subscriber notifications"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "resolve_incident",
		Description: "Resolve a Statuscast incident with a final message.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args resolveIncidentArgs) (*mcp.CallToolResult, any, error) {
		req := statuscast.ResolveRequest{
			Message: args.Message,
		}
		if args.Notify != nil {
			req.Notify = *args.Notify
		}
		inc, _, err := client.Incidents.Resolve(ctx, args.IncidentID, req)
		if err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Incident %s resolved. Title: %q, Status: %s", inc.ID, inc.Title, inc.Status)), nil, nil
	})

	// ─── get_uptime_report ────────────────────────────────────────────────────

	type getUptimeReportArgs struct {
		Since string `json:"since" jsonschema:"Start of the reporting window in RFC3339 format (e.g. 2024-01-01T00:00:00Z)"`
		Until string `json:"until" jsonschema:"End of the reporting window in RFC3339 format (e.g. 2024-01-31T23:59:59Z)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_uptime_report",
		Description: "Get uptime percentages per component for a given time window.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args getUptimeReportArgs) (*mcp.CallToolResult, any, error) {
		since, until, err := parseWindow(args.Since, args.Until)
		if err != nil {
			return nil, nil, err
		}
		reports, _, err := client.Reports.Uptime(ctx, since, until)
		if err != nil {
			return nil, nil, err
		}
		if len(reports) == 0 {
			return textResult("No uptime data found."), nil, nil
		}
		var sb strings.Builder
		for _, r := range reports {
			name := r.ComponentName
			if name == "" {
				name = r.ComponentID
			}
			fmt.Fprintf(&sb, "%s: %.4f%% uptime (%s – %s)\n",
				name, r.Uptime,
				r.WindowStart.Format(time.DateOnly), r.WindowEnd.Format(time.DateOnly))
		}
		return textResult(sb.String()), nil, nil
	})

	// ─── get_incident_summary ─────────────────────────────────────────────────

	type getIncidentSummaryArgs struct {
		Since string `json:"since" jsonschema:"Start of the reporting window in RFC3339 format (e.g. 2024-01-01T00:00:00Z)"`
		Until string `json:"until" jsonschema:"End of the reporting window in RFC3339 format (e.g. 2024-01-31T23:59:59Z)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_incident_summary",
		Description: "Get MTTD/MTTR analytics for incidents in a given time window.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args getIncidentSummaryArgs) (*mcp.CallToolResult, any, error) {
		since, until, err := parseWindow(args.Since, args.Until)
		if err != nil {
			return nil, nil, err
		}
		report, _, err := client.Reports.IncidentSummary(ctx, since, until)
		if err != nil {
			return nil, nil, err
		}
		var sb strings.Builder
		fmt.Fprintf(&sb, "Total incidents: %d\n", report.TotalIncidents)
		fmt.Fprintf(&sb, "MTTD: %s\n", formatDuration(report.MeanTimeToDetect))
		fmt.Fprintf(&sb, "MTTR: %s\n", formatDuration(report.MeanTimeToResolve))
		if len(report.ByComponent) > 0 {
			fmt.Fprintf(&sb, "By component:\n")
			for compID, count := range report.ByComponent {
				fmt.Fprintf(&sb, "  %s: %d incidents\n", compID, count)
			}
		}
		return textResult(sb.String()), nil, nil
	})
}

func parseWindow(since, until string) (time.Time, time.Time, error) {
	s, err := time.Parse(time.RFC3339, since)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid since: %w", err)
	}
	u, err := time.Parse(time.RFC3339, until)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid until: %w", err)
	}
	return s, u, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "N/A"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
