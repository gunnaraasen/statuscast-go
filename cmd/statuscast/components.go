package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	statuscast "statuscast-go"
)

func componentsCommand() *cli.Command {
	return &cli.Command{
		Name:  "components",
		Usage: "Manage status page components",
		Commands: []*cli.Command{
			componentsList(),
			componentsGet(),
			componentsCreate(),
			componentsUpdate(),
			componentsDelete(),
			componentsSetStatus(),
		},
	}
}

func componentsList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List components",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "parent-id", Usage: "Filter by parent component ID"},
			&cli.IntFlag{Name: "page", Value: 1},
			&cli.IntFlag{Name: "per-page", Value: 25},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			page := statuscast.Pagination{Page: int(cmd.Int("page")), PerPage: int(cmd.Int("per-page"))}
			result, _, err := c.Components.List(ctx, cmd.String("parent-id"), page)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(result)
			}
			rows := make([][]string, 0, len(result.Items))
			for _, comp := range result.Items {
				rows = append(rows, []string{comp.ID, comp.Name, string(comp.Status), string(comp.Type), comp.ParentID})
			}
			printTable([]string{"ID", "NAME", "STATUS", "TYPE", "PARENT_ID"}, rows)
			return nil
		},
	}
}

func componentsGet() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a component by ID",
		ArgsUsage: "<id>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			comp, _, err := c.Components.Get(ctx, cmd.Args().First())
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(comp)
			}
			printTable(
				[]string{"FIELD", "VALUE"},
				[][]string{
					{"ID", comp.ID},
					{"Name", comp.Name},
					{"Description", comp.Description},
					{"Status", string(comp.Status)},
					{"Type", string(comp.Type)},
					{"Parent ID", comp.ParentID},
					{"Created", formatTime(comp.CreatedAt)},
					{"Updated", formatTime(comp.UpdatedAt)},
				},
			)
			return nil
		},
	}
}

func componentsCreate() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a component",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Required: true, Usage: "Component name"},
			&cli.StringFlag{Name: "description", Usage: "Component description"},
			&cli.StringFlag{Name: "type", Usage: "Component type (native|beacon|third_party)"},
			&cli.StringFlag{Name: "parent-id", Usage: "Parent component ID"},
			&cli.StringFlag{Name: "status", Usage: "Initial status"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			req := statuscast.CreateComponentRequest{
				Name:        cmd.String("name"),
				Description: cmd.String("description"),
				Type:        statuscast.ComponentType(cmd.String("type")),
				ParentID:    cmd.String("parent-id"),
			}
			if s := cmd.String("status"); s != "" {
				req.InitialStatus = statuscast.ComponentStatus(s)
			}
			comp, _, err := c.Components.Create(ctx, req)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(comp)
			}
			fmt.Printf("Created component %s (%s)\n", comp.ID, comp.Name)
			return nil
		},
	}
}

func componentsUpdate() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a component",
		ArgsUsage: "<id>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Usage: "New name"},
			&cli.StringFlag{Name: "description", Usage: "New description"},
			&cli.StringFlag{Name: "status", Usage: "New status"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			req := statuscast.UpdateComponentRequest{}
			if s := cmd.String("name"); s != "" {
				req.Name = &s
			}
			if s := cmd.String("description"); s != "" {
				req.Description = &s
			}
			if s := cmd.String("status"); s != "" {
				st := statuscast.ComponentStatus(s)
				req.Status = &st
			}
			comp, _, err := c.Components.Update(ctx, cmd.Args().First(), req)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(comp)
			}
			fmt.Printf("Updated component %s\n", comp.ID)
			return nil
		},
	}
}

func componentsDelete() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a component",
		ArgsUsage: "<id>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required")
			}
			c := getClient(cmd)
			_, err := c.Components.Delete(ctx, cmd.Args().First())
			if err != nil {
				return err
			}
			fmt.Printf("Deleted component %s\n", cmd.Args().First())
			return nil
		},
	}
}

func componentsSetStatus() *cli.Command {
	return &cli.Command{
		Name:      "set-status",
		Usage:     "Set the status of a component",
		ArgsUsage: "<id> <status>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() < 2 {
				return fmt.Errorf("id and status required")
			}
			c := getClient(cmd)
			id := cmd.Args().Get(0)
			status := statuscast.ComponentStatus(cmd.Args().Get(1))
			comp, _, err := c.Components.SetStatus(ctx, id, status)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(comp)
			}
			fmt.Printf("Component %s status set to %s\n", comp.ID, strings.ToUpper(string(comp.Status)))
			return nil
		},
	}
}
