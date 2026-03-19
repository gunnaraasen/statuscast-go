# statuscast-go

A Go SDK for the [StatusCast](https://statuscast.com) API. Manage status pages, incidents, components, subscribers, and notifications from your Go applications.

## Features

- **Incidents** — create, update, resolve, and list incidents with full timeline support
- **Components** — CRUD operations and one-call status updates
- **Subscribers** — add, update, remove, and bulk-import from CSV
- **Groups** — list subscriber groups
- **Notifications** — manage notification templates
- **Reports** — uptime percentages and MTTD/MTTR analytics
- **Access** — invite team members, update roles, remove users
- Idiomatic Go API with typed enums and sentinel errors
- Functional options for custom HTTP clients and base URLs

## Installation

### SDK

Requires Go 1.22 or later.

```bash
go get statuscast-go
```

### CLI

Download a pre-built binary from the [releases page](https://github.com/gunnaraasen/statuscast-go/releases), or install from source:

```bash
go install statuscast-go/cmd/statuscast@latest
```

### MCP server

Download a pre-built binary from the [releases page](https://github.com/gunnaraasen/statuscast-go/releases), or install from source:

```bash
go install statuscast-go/cmd/statuscast-mcp@latest
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "statuscast-go"
)

func main() {
    client, err := statuscast.New(
        statuscast.WithAPIKey("your-api-key"),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Open an incident
    incident, _, err := client.Incidents.Create(ctx, statuscast.CreateIncidentRequest{
        Title:      "Database connectivity degraded",
        Message:    "We are investigating elevated error rates on the primary database.",
        Components: []string{"12345"},
        Status:     statuscast.StatusInvestigating,
        Notify:     true,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Incident created:", incident.ID)
}
```

## Usage

### Creating a client

```go
client, err := statuscast.New(
    statuscast.WithAPIKey("your-api-key"),

    // Optional: override the default base URL
    statuscast.WithBaseURL("https://app.statuscast.com"),

    // Optional: provide a custom HTTP client (e.g. with retry logic)
    statuscast.WithHTTPClient(&http.Client{Timeout: 15 * time.Second}),
)
```

### Components

```go
// Create a component
comp, _, err := client.Components.Create(ctx, statuscast.CreateComponentRequest{
    Name:          "API Gateway",
    Description:   "Primary API entry point",
    InitialStatus: statuscast.StatusOperational,
})

// Update component status
comp, _, err = client.Components.SetStatus(ctx, comp.ID, statuscast.StatusDegradedPerf)

// List all components
result, _, err := client.Components.List(ctx, "", statuscast.Pagination{Page: 1, PerPage: 50})
for _, c := range result.Items {
    fmt.Printf("%s: %s\n", c.Name, c.Status)
}

// Delete a component
_, err = client.Components.Delete(ctx, comp.ID)
```

**Component status constants:**

| Constant | Value |
|---|---|
| `StatusOperational` | `operational` |
| `StatusDegradedPerf` | `degraded_performance` |
| `StatusPartialOutage` | `partial_outage` |
| `StatusMajorOutage` | `major_outage` |
| `StatusUnderMaintenance` | `under_maintenance` |

**Component type constants:**

| Constant | Value |
|---|---|
| `ComponentTypeNative` | `native` |
| `ComponentTypeBeacon` | `beacon` |
| `ComponentTypeThirdPt` | `third_party` |

### Incidents

```go
// Open an incident
incident, _, err := client.Incidents.Create(ctx, statuscast.CreateIncidentRequest{
    Title:    "Payment processing delays",
    Message:  "Customers may experience slow checkout times.",
    Status:   statuscast.StatusInvestigating,
    PostType: statuscast.PostTypeOutage,
    Notify:   true,
})

// Post an update
_, _, err = client.Incidents.PostUpdate(ctx, incident.ID, statuscast.UpdateIncidentRequest{
    Message: "Root cause identified. Fix is being deployed.",
    Status:  statuscast.StatusIdentified,
    Notify:  true,
})

// Resolve the incident
_, _, err = client.Incidents.Resolve(ctx, incident.ID, statuscast.ResolveRequest{
    Message: "Service fully restored. We will publish a post-mortem within 48 hours.",
    Notify:  true,
})

// List active incidents
result, _, err := client.Incidents.List(ctx,
    statuscast.IncidentFilter{ActiveOnly: true},
    statuscast.Pagination{Page: 1},
)
```

**Incident status constants:**

| Constant | Value |
|---|---|
| `StatusInvestigating` | `investigating` |
| `StatusIdentified` | `identified` |
| `StatusMonitoring` | `monitoring` |
| `StatusResolved` | `resolved` |

**Incident post type constants:**

| Constant | Value |
|---|---|
| `PostTypeOutage` | `outage` |
| `PostTypeMaintenance` | `maintenance` |
| `PostTypeInfo` | `info` |

### Subscribers

```go
// Add a subscriber
sub, _, err := client.Subscribers.Add(ctx, statuscast.AddSubscriberRequest{
    Email:    "alice@example.com",
    Channels: []statuscast.NotificationChannel{statuscast.ChannelEmail},
})

// Bulk import from CSV
csvData := []byte("email\nalice@example.com\nbob@example.com\n")
result, _, err := client.Subscribers.BulkImport(ctx, csvData)
fmt.Printf("Imported: %d, Failed: %d\n", result.Imported, result.Failed)

// Remove a subscriber
_, err = client.Subscribers.Remove(ctx, sub.ID)
```

**Notification channel constants:**

| Constant | Value |
|---|---|
| `ChannelEmail` | `email` |
| `ChannelSMS` | `sms` |
| `ChannelSlack` | `slack` |
| `ChannelTeams` | `teams` |
| `ChannelWebhook` | `webhook` |

### Reports

```go
since := time.Now().AddDate(0, -1, 0) // last 30 days
until := time.Now()

// Uptime report
uptimes, _, err := client.Reports.Uptime(ctx, since, until)
for _, u := range uptimes {
    fmt.Printf("%.2f%% uptime\n", u.Uptime)
}

// Incident analytics (MTTD / MTTR)
summary, _, err := client.Reports.IncidentSummary(ctx, since, until)
fmt.Printf("MTTD: %v, MTTR: %v\n", summary.MeanTimeToDetect, summary.MeanTimeToResolve)
```

### Access control

```go
// Invite a new team member
user, _, err := client.Access.InviteUser(ctx, statuscast.InviteUserRequest{
    Email: "bob@example.com",
    Name:  "Bob Smith",
    Role:  statuscast.RoleManager,
})

// Update role
_, _, err = client.Access.UpdateRole(ctx, user.ID, statuscast.RoleAdministrator)

// Remove user (requires UUID — use the ID returned by InviteUser or ListUsers)
_, err = client.Access.RemoveUser(ctx, user.ID)
```

**Role constants:**

| Constant | Value |
|---|---|
| `RoleEmployee` | `employee` |
| `RoleManager` | `manager` |
| `RoleAdministrator` | `administrator` |
| `RoleCompanyAdministrator` | `company_administrator` |

### Error handling

The SDK returns typed errors for common failure modes:

```go
import "errors"

_, _, err := client.Components.Get(ctx, "99999")
if errors.Is(err, statuscast.ErrNotFound) {
    // component does not exist
}
if errors.Is(err, statuscast.ErrUnauthorized) {
    // invalid or expired API key
}
if errors.Is(err, statuscast.ErrRateLimited) {
    // back off and retry
}

// Inspect the raw API error
var apiErr *statuscast.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message)
}
```

**Sentinel errors:**

| Error | HTTP status |
|---|---|
| `ErrMissingAPIKey` | — |
| `ErrUnauthorized` | 401 |
| `ErrNotFound` | 404 |
| `ErrRateLimited` | 429 |
| `ErrIncidentClosed` | — |

## CLI

The `statuscast` CLI manages your StatusCast status page from the terminal.

### Authentication

Pass your API key via flag or environment variable:

```bash
export STATUSCAST_API_KEY=your-api-key
statuscast components list
```

```bash
statuscast --api-key your-api-key components list
```

### Global flags

| Flag | Description |
|---|---|
| `--api-key` | StatusCast API key (`STATUSCAST_API_KEY`) |
| `--base-url` | Override the default API base URL |
| `--json` | Output results as JSON |

### Commands

```
statuscast components list   [--parent-id ID] [--page N] [--per-page N]
statuscast components get    <id>
statuscast components create --name NAME [--description TEXT] [--type TYPE] [--status STATUS] [--parent-id ID]
statuscast components update <id> [--name NAME] [--description TEXT] [--status STATUS]
statuscast components delete <id>
statuscast components set-status <id> <status>

statuscast incidents list    [--active] [--component ID]... [--since DATE] [--until DATE] [--page N] [--per-page N]
statuscast incidents get     <id>
statuscast incidents create  --title TITLE --message MESSAGE [--status STATUS] [--type TYPE] [--component ID]... [--channel CHANNEL]... [--template-id ID] [--notify]
statuscast incidents update  <id> --message MESSAGE [--status STATUS] [--notify]
statuscast incidents resolve <id> --message MESSAGE [--notify]

statuscast subscribers list        [--group-id ID] [--page N] [--per-page N]
statuscast subscribers get         <id>
statuscast subscribers add         --email EMAIL [--phone PHONE] [--group ID]... [--component ID]... [--channel CHANNEL]...
statuscast subscribers update      <id> [--group ID]... [--component ID]... [--channel CHANNEL]...
statuscast subscribers remove      <id>
statuscast subscribers bulk-import --file subscribers.csv

statuscast groups list [--page N] [--per-page N]

statuscast notifications templates list   [--page N] [--per-page N]
statuscast notifications templates create --name NAME --channel CHANNEL --body BODY [--subject SUBJECT]
statuscast notifications templates update <id> [--name NAME] [--channel CHANNEL] [--body BODY] [--subject SUBJECT]

statuscast reports uptime           [--since DATE] [--until DATE]
statuscast reports incident-summary --since DATE --until DATE

statuscast access users list        [--page N] [--per-page N]
statuscast access users invite      --email EMAIL --name NAME --role ROLE
statuscast access users update-role <id> --role ROLE
statuscast access users remove      <id>
```

Dates accept RFC 3339 (`2006-01-02T15:04:05Z`) or plain date (`2006-01-02`) format.

## MCP server

`statuscast-mcp` is a [Model Context Protocol](https://modelcontextprotocol.io) stdio server that exposes Statuscast operations as tools for Claude and other MCP-compatible clients.

### Running the server

```bash
export STATUSCAST_API_KEY=your-api-key
statuscast-mcp
```

### Configuring Claude Desktop

Add the server to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "statuscast": {
      "command": "/path/to/statuscast-mcp",
      "env": {
        "STATUSCAST_API_KEY": "your-api-key"
      }
    }
  }
}
```

### Available tools

| Tool | Description |
|---|---|
| `list_components` | List all components and their current status |
| `set_component_status` | Update a component's status |
| `list_incidents` | List incidents with optional filters |
| `create_incident` | Open a new incident |
| `update_incident` | Post an update to an existing incident |
| `resolve_incident` | Resolve an incident |
| `get_uptime_report` | Get uptime percentages for a time window |
| `get_incident_summary` | Get MTTD/MTTR analytics for a time window |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
