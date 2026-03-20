package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"

	statuscast "statuscast-go"
)

const clientKey = "client"

func buildClient(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().First() == "version" {
		return ctx, nil
	}

	key := cmd.String("api-key")
	if key == "" {
		return ctx, errors.New("API key required: set --api-key or STATUSCAST_API_KEY")
	}

	opts := []statuscast.Option{statuscast.WithAPIKey(key)}
	if u := cmd.String("base-url"); u != "" {
		opts = append(opts, statuscast.WithBaseURL(u))
	}

	c, err := statuscast.New(opts...)
	if err != nil {
		return ctx, fmt.Errorf("init client: %w", err)
	}
	cmd.Metadata[clientKey] = c
	return ctx, nil
}

func getClient(cmd *cli.Command) *statuscast.Client {
	return cmd.Root().Metadata[clientKey].(*statuscast.Client)
}
