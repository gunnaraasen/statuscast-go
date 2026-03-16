package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func groupsCommand() *cli.Command {
	return &cli.Command{
		Name:  "groups",
		Usage: "Manage subscriber groups",
		Commands: []*cli.Command{
			groupsList(),
		},
	}
}

func groupsList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List groups",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "page", Value: 1},
			&cli.IntFlag{Name: "per-page", Value: 25},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			page := getPagination(cmd)
			result, _, err := c.Groups.List(ctx, page)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(result)
			}
			rows := make([][]string, 0, len(result.Items))
			for _, g := range result.Items {
				rows = append(rows, []string{g.ID, g.Name, formatTime(g.CreatedAt)})
			}
			printTable([]string{"ID", "NAME", "CREATED"}, rows)
			return nil
		},
	}
}
