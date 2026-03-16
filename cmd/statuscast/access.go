package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	statuscast "statuscast-go"
)

func accessCommand() *cli.Command {
	return &cli.Command{
		Name:  "access",
		Usage: "Manage admin users and access control",
		Commands: []*cli.Command{
			accessUsers(),
		},
	}
}

func accessUsers() *cli.Command {
	return &cli.Command{
		Name:  "users",
		Usage: "Manage admin users",
		Commands: []*cli.Command{
			accessUsersList(),
			accessUsersInvite(),
			accessUsersUpdateRole(),
			accessUsersRemove(),
		},
	}
}

func accessUsersList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List admin users",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "page", Value: 1},
			&cli.IntFlag{Name: "per-page", Value: 25},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			page := statuscast.Pagination{Page: int(cmd.Int("page")), PerPage: int(cmd.Int("per-page"))}
			result, _, err := c.Access.ListUsers(ctx, page)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(result)
			}
			rows := make([][]string, 0, len(result.Items))
			for _, u := range result.Items {
				rows = append(rows, []string{u.ID, u.Email, u.Name, string(u.Role), formatTime(u.CreatedAt)})
			}
			printTable([]string{"ID", "EMAIL", "NAME", "ROLE", "CREATED"}, rows)
			return nil
		},
	}
}

func accessUsersInvite() *cli.Command {
	return &cli.Command{
		Name:  "invite",
		Usage: "Invite a new admin user",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "email", Required: true, Usage: "User email address"},
			&cli.StringFlag{Name: "name", Required: true, Usage: "Full name"},
			&cli.StringFlag{Name: "role", Required: true, Usage: "Role (employee|manager|administrator|company_administrator)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := getClient(cmd)
			req := statuscast.InviteUserRequest{
				Email: cmd.String("email"),
				Name:  cmd.String("name"),
				Role:  statuscast.Role(cmd.String("role")),
			}
			u, _, err := c.Access.InviteUser(ctx, req)
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(u)
			}
			fmt.Printf("Invited user %s (%s)\n", u.ID, u.Email)
			return nil
		},
	}
}

func accessUsersUpdateRole() *cli.Command {
	return &cli.Command{
		Name:      "update-role",
		Usage:     "Update an admin user's role (id must be numeric)",
		ArgsUsage: "<id>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "role", Required: true, Usage: "New role (employee|manager|administrator|company_administrator)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required (numeric user ID)")
			}
			c := getClient(cmd)
			u, _, err := c.Access.UpdateRole(ctx, cmd.Args().First(), statuscast.Role(cmd.String("role")))
			if err != nil {
				return err
			}
			if useJSON(cmd) {
				return printJSON(u)
			}
			fmt.Printf("Updated role for user %s to %s\n", u.ID, u.Role)
			return nil
		},
	}
}

func accessUsersRemove() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove an admin user (id must be a UUID)",
		ArgsUsage: "<id>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("id required (UUID user ID)")
			}
			c := getClient(cmd)
			_, err := c.Access.RemoveUser(ctx, cmd.Args().First())
			if err != nil {
				return err
			}
			fmt.Printf("Removed user %s\n", cmd.Args().First())
			return nil
		},
	}
}
