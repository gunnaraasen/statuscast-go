# Contributing to statuscast-go

Thank you for your interest in contributing. This document covers everything you need to get started.

## Table of Contents

- [Getting started](#getting-started)
- [Development workflow](#development-workflow)
- [Reporting issues](#reporting-issues)
- [Submitting a pull request](#submitting-a-pull-request)
- [Code style](#code-style)
- [Testing](#testing)
- [Regenerating the internal client](#regenerating-the-internal-client)

---

## Getting started

**Prerequisites:**

- Go 1.22 or later
- [`golangci-lint`](https://golangci-lint.run/usage/install/) for linting

**Clone and build:**

```bash
git clone https://github.com/gunnaraasen/statuscast-go.git
cd statuscast-go
go build ./...
go test ./...
```

---

## Development workflow

1. Fork the repository and create a feature branch from `main`.
2. Make your changes. Keep commits focused and atomic.
3. Ensure all checks pass before opening a PR (see [Testing](#testing) and [Code style](#code-style)).
4. Open a pull request against `main`.

---

## Reporting issues

Before opening an issue, search [existing issues](https://github.com/gunnaraasen/statuscast-go/issues) to avoid duplicates.

When filing a bug, include:

- Go version (`go version`)
- A minimal, self-contained code sample that reproduces the problem
- The actual behavior and the expected behavior
- Any relevant error messages or stack traces

For feature requests, describe the use case and why the existing API does not cover it.

---

## Submitting a pull request

- **One PR per logical change.** Split unrelated fixes or features into separate PRs.
- **Write tests.** New behavior must be covered by tests in the corresponding `*_test.go` file. See [Testing](#testing).
- **Update documentation.** If you add or change public API surface, update the godoc comments and `README.md` examples as needed.
- **Do not edit `internal/`.** The `internal/statuscast/` package is auto-generated from `openapi.json` via ogen. Changes there will be overwritten. If the API spec needs updating, see [Regenerating the internal client](#regenerating-the-internal-client).

**PR checklist:**

- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `golangci-lint run ./...` passes with no new warnings
- [ ] Godoc comments are present on all exported symbols you added

---

## Code style

- Formatting is enforced by `goimports` with local prefix `statuscast-go`. Run it before committing:

  ```bash
  goimports -local statuscast-go -w .
  ```

- Follow the two-layer convention: public facade types live in the root package; the `internal/` generated client is never edited by hand.
- Use `switch r := res.(type)` for dispatching ogen response types — do not use `if` chains.
- All exported declarations must have a godoc comment ending in a period.

Linting configuration is in `.golangci.yml`. Run the full linter with:

```bash
golangci-lint run ./...
```

---

## Testing

Tests use `net/http/httptest` as a mock backend — no live StatusCast account is needed.

```bash
# Run all tests
go test ./...

# Run a single test
go test -run TestIncidentsClient_Create .

# Run with race detector
go test -race ./...
```

**Test helpers** (defined in `testhelpers_test.go`):

| Helper | Purpose |
|---|---|
| `mustNew(t, opts...)` | Creates a `Client` from options; panics on error |
| `newMockClient(t, handler)` | Starts an `httptest.Server` and returns a configured `Client` |
| `jsonHandler(status, body)` | Handler that responds with a JSON body and given status code |
| `statusHandler(status)` | Handler that responds with only a status code |

When adding a new method, add a test in the corresponding `*_test.go` (e.g. `components_test.go`) using `http.NewServeMux()` with Go 1.22 method+path syntax:

```go
mux.HandleFunc("POST /api/v4/component", jsonHandler(200, `{...}`))
```

---

## Regenerating the internal client

The `internal/statuscast/` package is generated from `openapi.json` using [ogen](https://github.com/ogen-go/ogen). If the StatusCast API spec changes, regenerate it with:

```bash
go run github.com/ogen-go/ogen/cmd/ogen --config ogen.yml openapi.json
```

After regenerating, update any facade methods in the root package that are affected by the spec change, and update tests accordingly.
