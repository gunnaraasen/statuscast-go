package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// version is set at build time via -ldflags.
var version = "dev"

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print the version",
		Action: func(_ context.Context, _ *cli.Command) error {
			fmt.Println(version)
			return nil
		},
	}
}
