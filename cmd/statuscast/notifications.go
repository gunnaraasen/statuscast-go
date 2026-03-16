package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	statuscast "statuscast-go"
)

func notificationsCommand() *cli.Command {
	return &cli.Command{
		Name:  "notifications",
		Usage: "Manage notification templates",
		Commands: []*cli.Command{
			notificationsTemplates(),
		},
	}
}

func notificationsTemplates() *cli.Command {
	return &cli.Command{
		Name:  "templates",
		Usage: "Manage notification templates",
		Commands: []*cli.Command{
			templatesList(),
			templatesCreate(),
			templatesUpdate(),
		},
	}
}

func templatesList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List notification templates",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "page", Value: 1},
			&cli.IntFlag{Name: "per-page", Value: 25},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			page := getPagination(cmd)
			result, _, err := c.Notifications.ListTemplates(ctx, page)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(result)
			}
			rows := make([][]string, 0, len(result.Items))
			for _, t := range result.Items {
				rows = append(rows, []string{t.ID, t.Name, string(t.Channel), t.Subject})
			}
			printTable([]string{"ID", "NAME", "CHANNEL", "SUBJECT"}, rows)
			return nil
		},
	}
}

func templatesCreate() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a notification template",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Required: true, Usage: "Template name"},
			&cli.StringFlag{Name: "channel", Required: true, Usage: "Channel (email|sms|slack|teams|webhook)"},
			&cli.StringFlag{Name: "body", Required: true, Usage: "Template body (HTML or Markdown)"},
			&cli.StringFlag{Name: "subject", Usage: "Email subject (for email channel)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			tmpl := statuscast.NotificationTemplate{
				Name:    cmd.String("name"),
				Channel: statuscast.NotificationChannel(cmd.String("channel")),
				Body:    cmd.String("body"),
				Subject: cmd.String("subject"),
			}
			t, _, err := c.Notifications.CreateTemplate(ctx, tmpl)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(t)
			}
			fmt.Printf("Created template %s\n", t.ID)
			return nil
		},
	}
}

func templatesUpdate() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a notification template",
		ArgsUsage: "<id>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Usage: "Template name"},
			&cli.StringFlag{Name: "channel", Usage: "Channel (email|sms|slack|teams|webhook)"},
			&cli.StringFlag{Name: "body", Usage: "Template body (HTML or Markdown)"},
			&cli.StringFlag{Name: "subject", Usage: "Email subject"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			tmpl := statuscast.NotificationTemplate{
				Name:    cmd.String("name"),
				Channel: statuscast.NotificationChannel(cmd.String("channel")),
				Body:    cmd.String("body"),
				Subject: cmd.String("subject"),
			}
			t, _, err := c.Notifications.UpdateTemplate(ctx, cmd.Args().First(), tmpl)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(t)
			}
			fmt.Printf("Updated template %s\n", t.ID)
			return nil
		},
	}
}
