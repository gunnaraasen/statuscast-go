# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Vet
go vet ./...

# Run all tests
go test ./...

# Run a single test
go test -run TestComponentsClient_Create .

# Regenerate internal client from OpenAPI spec
go run github.com/ogen-go/ogen/cmd/ogen --config ogen.yml openapi.json
```

## Architecture

**statuscast-go** is a facade SDK over an ogen-generated OpenAPI client.

### Two-layer design
- **Root package** (`package statuscast`): Hand-written facade in root dir — the public API consumed by users
- **`internal/statuscast/`** (`package statuscast`): Auto-generated ogen client (do not edit)
- Import alias in root files: `api "statuscast-go/internal/statuscast"` to avoid the package name collision

### Key files
- `statuscast.go` — `Client` struct, all public types/constants, `New()` constructor, functional options
- `helpers.go` — `idToInt32`, `int32ToID`, status/role enum mapping between API and facade
- `components.go`, `incidents.go`, `subscribers.go`, `access.go`, `notifications.go`, `groups.go`, `reports.go` — one file per sub-client
- `openapi.json` / `ogen.yml` — source of truth for regenerating `internal/`

### ID handling
- API uses `int32` for most IDs; facade exposes `string`
- Admin users have **two** ID forms: `int32` (for `UpdateRole`) and UUID (for `RemoveUser`)
- `AdminUser.ID` stores UUID when available, int32 string as fallback

### Status enums
API and facade use different strings for component status. Translation lives in `helpers.go`:
- API: `Available`, `DegradedPerformance`, `Unavailable`, `Maintenance`
- Facade: `operational`, `degraded_performance`, `major_outage`, `under_maintenance`
- Both `StatusPartialOutage` and `StatusMajorOutage` map to API `Unavailable`

### Unsupported operations
Six operations have no backing API endpoint and are not implemented (they have been removed from the codebase entirely, along with their associated types):
`Teams.List`, `Incidents.FileRCA`, `Notifications.GetLog`, `Notifications.ListLogs`, `Reports.ListRCAs`, `Access.SetPageVisibility`

`Subscribers.BulkImport` and `Reports.IncidentSummary` are implemented client-side (see below).

### Base URL
`https://app.statuscast.com` (domain root). API paths contain `/api/v4/…` prefix so the server URL must be the domain root, not `https://api.statuscast.com/v1`.

### Client-side implementations
Two operations are implemented without a direct API endpoint:

- **`Subscribers.BulkImport`** — parses CSV (`encoding/csv`), finds the `email` column by header name (case-insensitive), calls `APIV4SubscriberPost` per row. Empty/whitespace emails are `Skipped`; per-row API failures accumulate into `BulkImportResult.Errors` rather than returning a function error. Only structural CSV failures (missing header, malformed CSV) return an error.

- **`Reports.IncidentSummary`** — calls `APIV4IncidentsPost` with `StartDateAfter`/`EndDateBefore` filters, paginating through all results (100/page). Computes MTTD (`DateCreated − StartDate` when `StartDate < DateCreated`) and MTTR (`EndDate − DateCreated` for resolved incidents) client-side.

## Testing

Tests use `httptest.Server` as a mock backend. See `testhelpers_test.go` for helpers:
- `mustNew(t, opts...)` — creates a Client from opts; panics on error (not `t.Fatal`) so SA5011 models it as a definite non-nil return
- `newMockClient(t, handler)` — starts an httptest.Server and returns a configured Client; server closed via `t.Cleanup`
- `jsonHandler(status, body)` — writes JSON with the given status code
- `statusHandler(status)` — writes only a status code

Each domain has its own `*_test.go`. Routes are registered with `http.NewServeMux()` using Go 1.22 method+path syntax (e.g. `GET /api/v4/component/{id}`).

## Linting

Run the linter:
```bash
golangci-lint run ./...
```

`.golangci.yml` enables: `errcheck`, `govet`, `staticcheck` (all checks; ST1000 and ST1020 disabled), `ineffassign`, `misspell`, `godot` (declarations scope). Formatter: `goimports` with local prefix `statuscast-go`.

Exclusions:
- `internal/` — never linted (ogen-generated)
- `*_test.go` — errcheck and godot disabled
