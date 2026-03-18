# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

An MCP (Model Context Protocol) server in Go that exposes 48 tools for interacting with the OpenProject API v3. Built with the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) (`github.com/modelcontextprotocol/go-sdk v0.3.0`).

**Dual-mode operation:**
1. **MCP Server** (default): stdio/SSE/HTTP transports for AI assistants
2. **CLI**: Direct command-line interaction via Cobra subcommands

## Build & Run

```bash
make build          # Build binary to ./build/openproject-mcp
make test           # Run all tests (go test -v ./...)
make build-all      # Cross-compile for linux/darwin/windows
make deps           # go mod download && go mod tidy
make clean          # Remove build artifacts
```

Run a single test: `go test -v -run TestName ./path/to/package`

**Environment:** Requires `OPENPROJECT_URL` and `OPENPROJECT_API_KEY` env vars. Supports three transport modes: `stdio` (default, for MCP clients), `sse`, and `http`. Transport and port configurable via flags (`-transport=sse -port=3000`) or env vars (`TRANSPORT`, `PORT`).

## Architecture

```
main.go                        → Entry point: delegates to cmd.Execute()
cmd/                           → Cobra CLI (root.go + domain files)
  root.go                      → Global flags, client init, default → MCP server
  mcp.go                       → MCP subcommand: transport/port flags
  *.go                         → Domain CLI commands (project.go, work_package.go, etc.)
internal/config/               → Env-based configuration (Config struct, validation)
internal/openproject/          → API client layer
  client.go                    → HTTP client with Basic Auth (apikey:<token>) against /api/v3/*
  types.go                     → HAL+JSON response types, RichText helper
  *.go                         → Domain methods (projects, work_packages, users, etc.)
internal/tools/                → MCP tool definitions
  tools.go                     → Registry that wires all tools to the MCP server
  *.go                         → Tool handlers grouped by domain with Args structs
pkg/server/                    → MCP server setup: creates mcp.Server, registers tools, runs transport
```

**Key flow (MCP mode):** `main` → `cmd.Execute()` → `runMCPServer()` → `server.New()` → `server.Run()` → transport-specific handler → `createMCPServer()` → `tools.RegisterAllTools()`

**Key flow (CLI mode):** `main` → `cmd.Execute()` → Cobra subcommand → uses global `client` → calls `openproject.Client` methods

## Credential Modes (HTTP/SSE only)

1. **Server pre-configured (single-tenant):** Set `OPENPROJECT_URL` and `OPENPROJECT_API_KEY` at startup
2. **Per-request credentials (multi-tenant):** Clients pass `X-OpenProject-URL` and `X-OpenProject-API-Key` headers; `pkg/server/withClientMiddleware` resolves client per-request with caching

For stdio transport, credentials are required at startup.

## Patterns to Follow

- Tool handlers are methods on `tools.Registry` and return `(*mcp.CallToolResult, error)`. Errors from the OpenProject API are returned as `IsError: true` in the result, not as Go errors.
- Each domain (projects, work_packages, etc.) has parallel files: `internal/openproject/<domain>.go` (API methods) and `internal/tools/<domain>.go` (MCP tools). CLI commands live in `cmd/<domain>.go`.
- Tool argument structs are defined at the top of each tools file (e.g., `ListProjectsArgs`, `CreateProjectArgs`).
- API response types use HAL+JSON: `_links` mapped to typed `*Links` structs, `_embedded` mapped to structs with `Elements` slices.
- `RichText` type handles OpenProject's `{"format":"markdown","raw":"..."}` format — use `NewRichText(s)` for description fields in write requests.
- The `openproject.Client` exposes `Get`, `Post`, `Patch`, `Delete` methods that prepend `/api/v3` to paths.
- Version is injected at build time via `-ldflags "-X main.Version=$(VERSION)"`.
- Logging goes to stderr (JSON for stdio transport, text for others) to avoid interfering with MCP protocol on stdout.
- URL query parameters with JSON filters must be URL-encoded (use `url.QueryEscape`).
