package main

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"
)

func reportsCommand() *cli.Command {
	return &cli.Command{
		Name:  "reports",
		Usage: "View uptime and incident analytics",
		Commands: []*cli.Command{
			reportsUptime(),
			reportsIncidentSummary(),
		},
	}
}

func reportsUptime() *cli.Command {
	return &cli.Command{
		Name:  "uptime",
		Usage: "Show uptime report",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "since", Usage: "Start of window (RFC3339 or 2006-01-02)"},
			&cli.StringFlag{Name: "until", Usage: "End of window (RFC3339 or 2006-01-02)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			since, until, err := parseSinceUntil(cmd)
			if err != nil {
				return err
			}
			c := getClient(cmd)
			reports, _, err := c.Reports.Uptime(ctx, since, until)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(reports)
			}
			rows := make([][]string, 0, len(reports))
			for _, r := range reports {
				rows = append(rows, []string{
					r.ComponentID,
					r.ComponentName,
					fmt.Sprintf("%.4f%%", r.Uptime),
					formatTime(r.WindowStart),
					formatTime(r.WindowEnd),
				})
			}
			printTable([]string{"COMPONENT_ID", "NAME", "UPTIME", "WINDOW_START", "WINDOW_END"}, rows)
			return nil
		},
	}
}

func reportsIncidentSummary() *cli.Command {
	return &cli.Command{
		Name:  "incident-summary",
		Usage: "Show incident summary with MTTD/MTTR",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "since", Required: true, Usage: "Start of window (RFC3339 or 2006-01-02)"},
			&cli.StringFlag{Name: "until", Required: true, Usage: "End of window (RFC3339 or 2006-01-02)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			since, until, err := parseSinceUntil(cmd)
			if err != nil {
				return err
			}
			c := getClient(cmd)
			report, _, err := c.Reports.IncidentSummary(ctx, since, until)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(report)
			}
			printTable(
				[]string{"FIELD", "VALUE"},
				[][]string{
					{"Total Incidents", fmt.Sprintf("%d", report.TotalIncidents)},
					{"Mean Time to Detect", formatDuration(report.MeanTimeToDetect)},
					{"Mean Time to Resolve", formatDuration(report.MeanTimeToResolve)},
					{"Since", formatTime(report.Since)},
					{"Until", formatTime(report.Until)},
				},
			)
			if len(report.ByComponent) > 0 {
				fmt.Println("\nBy Component:")
				for compID, count := range report.ByComponent {
					fmt.Printf("  %s: %d\n", compID, count)
				}
			}
			return nil
		},
	}
}

func parseSinceUntil(cmd *cli.Command) (time.Time, time.Time, error) {
	var since, until time.Time
	if s := cmd.String("since"); s != "" {
		t, err := parseTime(s)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		since = t
	}
	if s := cmd.String("until"); s != "" {
		t, err := parseTime(s)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		until = t
	}
	return since, until, nil
}
