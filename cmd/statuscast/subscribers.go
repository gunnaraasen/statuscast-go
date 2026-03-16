package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	statuscast "statuscast-go"
)

func subscribersCommand() *cli.Command {
	return &cli.Command{
		Name:  "subscribers",
		Usage: "Manage subscribers",
		Commands: []*cli.Command{
			subscribersList(),
			subscribersGet(),
			subscribersAdd(),
			subscribersUpdate(),
			subscribersRemove(),
			subscribersBulkImport(),
		},
	}
}

func subscribersList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List subscribers",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "group-id", Usage: "Filter by group ID"},
			&cli.IntFlag{Name: "page", Value: 1},
			&cli.IntFlag{Name: "per-page", Value: 25},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			page := getPagination(cmd)
			result, _, err := c.Subscribers.List(ctx, cmd.String("group-id"), page)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(result)
			}
			rows := make([][]string, 0, len(result.Items))
			for _, s := range result.Items {
				rows = append(rows, []string{s.ID, s.Email, s.Phone, strings.Join(channelStrings(s.Channels), ","), formatTime(s.CreatedAt)})
			}
			printTable([]string{"ID", "EMAIL", "PHONE", "CHANNELS", "CREATED"}, rows)
			return nil
		},
	}
}

func subscribersGet() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a subscriber by ID",
		ArgsUsage: "<id>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			s, _, err := c.Subscribers.Get(ctx, cmd.Args().First())
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(s)
			}
			printTable(
				[]string{"FIELD", "VALUE"},
				[][]string{
					{"ID", s.ID},
					{"Email", s.Email},
					{"Phone", s.Phone},
					{"Channels", strings.Join(channelStrings(s.Channels), ", ")},
					{"Groups", strings.Join(s.Groups, ", ")},
					{"Components", strings.Join(s.Components, ", ")},
					{"Created", formatTime(s.CreatedAt)},
				},
			)
			return nil
		},
	}
}

func subscribersAdd() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a subscriber",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "email", Required: true, Usage: "Subscriber email address"},
			&cli.StringFlag{Name: "phone", Usage: "Phone number (E.164)"},
			&cli.StringSliceFlag{Name: "group", Usage: "Group ID (repeatable)"},
			&cli.StringSliceFlag{Name: "component", Usage: "Component ID (repeatable)"},
			&cli.StringSliceFlag{Name: "channel", Usage: "Notification channel (repeatable)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			req := statuscast.AddSubscriberRequest{
				Email:      cmd.String("email"),
				Phone:      cmd.String("phone"),
				Groups:     cmd.StringSlice("group"),
				Components: cmd.StringSlice("component"),
			}
			req.Channels = toChannels(cmd.StringSlice("channel"))
			s, _, err := c.Subscribers.Add(ctx, req)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(s)
			}
			fmt.Printf("Added subscriber %s (%s)\n", s.ID, s.Email)
			return nil
		},
	}
}

func subscribersUpdate() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a subscriber",
		ArgsUsage: "<id>",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{Name: "group", Usage: "Group ID (repeatable)"},
			&cli.StringSliceFlag{Name: "component", Usage: "Component ID (repeatable)"},
			&cli.StringSliceFlag{Name: "channel", Usage: "Notification channel (repeatable)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			req := statuscast.UpdateSubscriberRequest{
				Groups:     cmd.StringSlice("group"),
				Components: cmd.StringSlice("component"),
			}
			req.Channels = toChannels(cmd.StringSlice("channel"))
			s, _, err := c.Subscribers.Update(ctx, cmd.Args().First(), req)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(s)
			}
			fmt.Printf("Updated subscriber %s\n", s.ID)
			return nil
		},
	}
}

func subscribersRemove() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a subscriber",
		ArgsUsage: "<id>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			_, err := c.Subscribers.Remove(ctx, cmd.Args().First())
			if err != nil {
				return err
			}
			fmt.Printf("Removed subscriber %s\n", cmd.Args().First())
			return nil
		},
	}
}

func subscribersBulkImport() *cli.Command {
	return &cli.Command{
		Name:  "bulk-import",
		Usage: "Import subscribers from a CSV file",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "file", Required: true, Usage: "Path to CSV file with 'email' column"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			data, err := os.ReadFile(cmd.String("file"))
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
			c := getClient(cmd)
			result, _, err := c.Subscribers.BulkImport(ctx, data)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(result)
			}
			printTable(
				[]string{"IMPORTED", "SKIPPED", "FAILED"},
				[][]string{{
					fmt.Sprintf("%d", result.Imported),
					fmt.Sprintf("%d", result.Skipped),
					fmt.Sprintf("%d", result.Failed),
				}},
			)
			if len(result.Errors) > 0 {
				fmt.Println("\nErrors:")
				for _, e := range result.Errors {
					fmt.Printf("  row %d (%s): %s\n", e.Row, e.Email, e.Message)
				}
			}
			return nil
		},
	}
}
