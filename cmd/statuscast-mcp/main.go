// Command statuscast-mcp is an MCP server that exposes Statuscast operations
// as Claude tools. It reads STATUSCAST_API_KEY from the environment and
// communicates over stdio using the Model Context Protocol.
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	statuscast "statuscast-go"
)

func main() {
	apiKey := os.Getenv("STATUSCAST_API_KEY")
	if apiKey == "" {
		log.Fatal("STATUSCAST_API_KEY is required")
	}
	opts := []statuscast.Option{statuscast.WithAPIKey(apiKey)}
	if baseURL := os.Getenv("STATUSCAST_BASE_URL"); baseURL != "" {
		opts = append(opts, statuscast.WithBaseURL(baseURL))
	}
	client, err := statuscast.New(opts...)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "statuscast",
		Version: "1.0.0",
	}, nil)
	registerTools(s, client)

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
