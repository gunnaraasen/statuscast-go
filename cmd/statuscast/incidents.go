package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	statuscast "statuscast-go"
)

func incidentsCommand() *cli.Command {
	return &cli.Command{
		Name:  "incidents",
		Usage: "Manage incidents",
		Commands: []*cli.Command{
			incidentsList(),
			incidentsGet(),
			incidentsCreate(),
			incidentsUpdate(),
			incidentsResolve(),
		},
	}
}

func incidentsList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List incidents",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "active", Usage: "Only show active incidents"},
			&cli.StringSliceFlag{Name: "component", Usage: "Filter by component ID (repeatable)"},
			&cli.StringFlag{Name: "since", Usage: "Filter incidents after date (RFC3339 or 2006-01-02)"},
			&cli.StringFlag{Name: "until", Usage: "Filter incidents before date (RFC3339 or 2006-01-02)"},
			&cli.IntFlag{Name: "page", Value: 1},
			&cli.IntFlag{Name: "per-page", Value: 25},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			filter := statuscast.IncidentFilter{
				ActiveOnly:   cmd.Bool("active"),
				ComponentIDs: cmd.StringSlice("component"),
			}
			if s := cmd.String("since"); s != "" {
				t, err := parseTime(s)
				if err != nil {
					return err
				}
				filter.Since = &t
			}
			if s := cmd.String("until"); s != "" {
				t, err := parseTime(s)
				if err != nil {
					return err
				}
				filter.Until = &t
			}
			page := getPagination(cmd)
			result, _, err := c.Incidents.List(ctx, filter, page)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(result)
			}
			rows := make([][]string, 0, len(result.Items))
			for _, inc := range result.Items {
				rows = append(rows, []string{
					inc.ID,
					inc.Title,
					string(inc.Status),
					string(inc.PostType),
					formatTime(inc.CreatedAt),
					formatOptTime(inc.ResolvedAt),
				})
			}
			printTable([]string{"ID", "TITLE", "STATUS", "TYPE", "CREATED", "RESOLVED"}, rows)
			return nil
		},
	}
}

func incidentsGet() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get an incident by ID",
		ArgsUsage: "<id>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			inc, _, err := c.Incidents.Get(ctx, cmd.Args().First())
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(inc)
			}
			printTable(
				[]string{"FIELD", "VALUE"},
				[][]string{
					{"ID", inc.ID},
					{"Title", inc.Title},
					{"Status", string(inc.Status)},
					{"Type", string(inc.PostType)},
					{"Components", strings.Join(inc.Components, ", ")},
					{"Created", formatTime(inc.CreatedAt)},
					{"Resolved", formatOptTime(inc.ResolvedAt)},
					{"Updates", fmt.Sprintf("%d", len(inc.Updates))},
				},
			)
			return nil
		},
	}
}

func incidentsCreate() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create an incident",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "title", Required: true, Usage: "Incident title"},
			&cli.StringFlag{Name: "message", Required: true, Usage: "Initial update message"},
			&cli.StringFlag{Name: "status", Value: "investigating", Usage: "Initial status"},
			&cli.StringFlag{Name: "type", Usage: "Post type (outage|maintenance|info)"},
			&cli.StringSliceFlag{Name: "component", Usage: "Affected component IDs (repeatable)"},
			&cli.StringSliceFlag{Name: "channel", Usage: "Notification channels (repeatable)"},
			&cli.StringFlag{Name: "template-id", Usage: "Notification template ID"},
			&cli.BoolFlag{Name: "notify", Usage: "Send subscriber notifications"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			req := statuscast.CreateIncidentRequest{
				Title:      cmd.String("title"),
				Message:    cmd.String("message"),
				Status:     statuscast.IncidentStatus(cmd.String("status")),
				PostType:   statuscast.IncidentPostType(cmd.String("type")),
				Components: cmd.StringSlice("component"),
				TemplateID: cmd.String("template-id"),
				Notify:     cmd.Bool("notify"),
			}
			req.Channels = toChannels(cmd.StringSlice("channel"))
			inc, _, err := c.Incidents.Create(ctx, req)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(inc)
			}
			fmt.Printf("Created incident %s: %s\n", inc.ID, inc.Title)
			return nil
		},
	}
}

func incidentsUpdate() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Post an update to an incident",
		ArgsUsage: "<id>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "message", Required: true, Usage: "Update message"},
			&cli.StringFlag{Name: "status", Usage: "New incident status"},
			&cli.BoolFlag{Name: "notify", Usage: "Send subscriber notifications"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			req := statuscast.UpdateIncidentRequest{
				Message: cmd.String("message"),
				Status:  statuscast.IncidentStatus(cmd.String("status")),
				Notify:  cmd.Bool("notify"),
			}
			update, _, err := c.Incidents.PostUpdate(ctx, cmd.Args().First(), req)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(update)
			}
			fmt.Printf("Posted update %s to incident %s\n", update.ID, cmd.Args().First())
			return nil
		},
	}
}

func incidentsResolve() *cli.Command {
	return &cli.Command{
		Name:      "resolve",
		Usage:     "Resolve an incident",
		ArgsUsage: "<id>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "message", Required: true, Usage: "Resolution message"},
			&cli.BoolFlag{Name: "notify", Usage: "Send subscriber notifications"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			req := statuscast.ResolveRequest{
				Message: cmd.String("message"),
				Notify:  cmd.Bool("notify"),
			}
			inc, _, err := c.Incidents.Resolve(ctx, cmd.Args().First(), req)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(inc)
			}
			fmt.Printf("Resolved incident %s\n", inc.ID)
			return nil
		},
	}
}
