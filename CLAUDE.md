# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

An MCP (Model Context Protocol) server in Go that exposes 41 tools for interacting with the OpenProject API v3. Built with the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) (`github.com/modelcontextprotocol/go-sdk v0.3.0`).

## Build & Run

```bash
make build          # Build binary to ./build/openproject-mcp
make test           # Run all tests (go test -v ./...)
make build-all      # Cross-compile for linux/darwin/windows
make deps           # go mod download && go mod tidy
make clean          # Remove build artifacts
```

Run a single test: `go test -v -run TestName ./path/to/package`

The binary requires `OPENPROJECT_URL` and `OPENPROJECT_API_KEY` env vars. Supports three transport modes: `stdio` (default, for MCP clients), `sse`, and `http`. Transport and port are configurable via flags (`-transport=sse -port=3000`) or env vars (`TRANSPORT`, `PORT`).

## Architecture

```
cmd/openproject-mcp/main.go   → Entry point: flag parsing, config loading, signal handling
internal/config/               → Env-based configuration (Config struct, validation)
internal/openproject/          → API client layer
  client.go                    → HTTP client with Bearer auth against /api/v3/*
  types.go                     → HAL+JSON response types (all follow OpenProject's _links/_embedded pattern)
  *.go                         → Domain methods (projects, work_packages, users, etc.)
internal/tools/                → MCP tool definitions
  tools.go                     → Registry that wires all tools to the MCP server
  *.go                         → Tool handlers grouped by domain, each with Args structs + handler methods
pkg/server/                    → MCP server setup: creates mcp.Server, registers tools, runs transport
```

Key flow: `main` → `config.Load()` → `server.New()` → `openproject.NewClient()` + `tools.NewRegistry()` → `registry.RegisterAllTools()` → `server.Run()` (stdio/SSE/HTTP).

## Patterns to Follow

- Tool handlers are methods on `tools.Registry` and return `(*mcp.CallToolResult, error)`. Errors from the OpenProject API are returned as `IsError: true` in the result, not as Go errors.
- Each domain (projects, work_packages, etc.) has a parallel file in both `internal/openproject/` (API methods) and `internal/tools/` (MCP tool definitions + handlers).
- Tool argument structs are defined at the top of each tools file (e.g., `ListProjectsArgs`, `CreateProjectArgs`).
- API response types use HAL+JSON conventions: `_links` mapped to typed `*Links` structs, `_embedded` mapped to `map[string]interface{}`. List types embed a struct with `Elements` slice.
- The `openproject.Client` exposes `Get`, `Post`, `Patch`, `Delete` methods that prepend `/api/v3` to paths.
- Version is injected at build time via `-ldflags "-X main.Version=$(VERSION)"`.
- Logging goes to stderr (JSON for stdio transport, text for others) to avoid interfering with MCP protocol on stdout.
