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
- `components.go`, `incidents.go`, `subscribers.go`, `access.go`, `notifications.go`, `groups.go`, `reports.go`, `teams.go` — one file per sub-client
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
Nine operations return `errors.New("not supported by StatusCast API v4")`:
`Teams.List`, `Incidents.FileRCA`, `Subscribers.BulkImport`, `Notifications.GetLog`, `Notifications.ListLogs`, `Reports.IncidentSummary`, `Reports.ListRCAs`, `Access.SetPageVisibility`

### Base URL
`https://app.statuscast.com` (domain root). API paths contain `/api/v4/…` prefix so the server URL must be the domain root, not `https://api.statuscast.com/v1`.

## Testing

Tests use `httptest.Server` as a mock backend. See `testhelpers_test.go` for `newMockClient()`, `jsonHandler()`, and `statusHandler()`. Each domain has its own `*_test.go`. `unsupported_test.go` asserts all nine unsupported operations return the right error.
