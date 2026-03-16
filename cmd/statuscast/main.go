// Command statuscast is a CLI for managing StatusCast status pages,
// incidents, components, subscribers, and notifications.
package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

func newApp() *cli.Command {
	return &cli.Command{
		Name:  "statuscast",
		Usage: "Manage StatusCast status pages from the command line",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-key",
				Usage:   "StatusCast API key (env: STATUSCAST_API_KEY)",
				Sources: cli.EnvVars("STATUSCAST_API_KEY"),
			},
			&cli.StringFlag{
				Name:  "base-url",
				Usage: "Override the default API base URL",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output results as JSON",
			},
		},
		Before: buildClient,
		Commands: []*cli.Command{
			componentsCommand(),
			incidentsCommand(),
			subscribersCommand(),
			groupsCommand(),
			notificationsCommand(),
			reportsCommand(),
			accessCommand(),
		},
	}
}

func main() {
	if err := newApp().Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
