# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

An MCP (Model Context Protocol) server in Go that exposes 80+ tools for interacting with the OpenProject API v3. Built with the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) and the generated [OpenProject API client](https://github.com/pinealctx/openproject) (`github.com/pinealctx/openproject`).

Also provides a CLI for direct OpenProject operations.

## Architecture

```
cmd/                     CLI commands and MCP server entrypoint
  root.go               Root cobra command, client init
  mcp.go                MCP server subcommand (stdio/sse/http)
  output.go             CLI output formatting (tabwriter + JSON)
  *.go                  One file per CLI domain (project, work_package, etc.)

internal/
  config/               Configuration from env vars
    config.go           Config struct, Load(), Validate()
  openproject/          Adapter wrapping the external API client
    client.go           Client struct, NewClient(), NewClientDirect(), APIClient()
    types.go            Type aliases re-exporting external module types
  tools/                MCP tool handlers (one file per group)
    tools.go            Registry, tool mode selection, schema helpers
    helpers.go          Shared helpers (parseArgs, formatUser, etc.)
    connection.go       test_connection, check_permissions, get_current_user, get_api_info
    projects.go         list/get/create/update/delete_project
    work_packages.go    list/get/create/update/delete_work_package + types/statuses/priorities
    users.go            list_users, get_user
    memberships.go      list/get/create/update/delete_membership + list_project_members, list_roles, get_role
    time_entries.go     list/create/update/delete_time_entry + list_time_entry_activities
    versions.go         list/create/update/delete_version
    relations.go        parent/child ops + relation CRUD
    search.go           search
    boards.go           board CRUD + widget management
    notifications.go    list/read notifications
    comments.go         list activities, create comment
    watchers.go         list/add/remove watchers
    groups.go           group CRUD
    documents.go        list/get/update documents
    queries.go          list/get queries
    wiki.go             list/get/update wiki pages
    placeholders.go     placeholder user CRUD
    configurations.go   view configuration

pkg/server/             MCP server with multi-transport support
  server.go             Server struct, stdio/sse/http transports, middleware
```

## Key Dependencies

- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk/mcp` — server, tools, transports
- **OpenProject client**: `github.com/pinealctx/openproject` — oapi-codegen generated from OpenAPI spec
- **JSON Schema**: `github.com/google/jsonschema-go/jsonschema` — tool input schemas

## External Client Patterns

The external client (`github.com/pinealctx/openproject`) returns raw `(*http.Response, error)` from all API calls. Use `openproject.ReadResponse(resp, &target)` to unmarshal.

Key API method naming differences:
- `ViewProject` (not `GetProject`)
- `ListAllTypes` / `ListTypesAvailableInAProject` (not `ListTypes`)
- `ListAllPriorities` (not `ListPriorities`)
- `CreateProjectWorkPackage` takes `WorkPackageModel` body
- `UpdateWorkPackage` uses `WorkPackagePatchModel` (requires `LockVersion`)

Auth is injected via `external.WithRequestEditorFn(basicAuthEditor(apiKey))`.

## Tool Mode System

Three modes control which MCP tools are registered:

| Mode | Tool Count | Config |
|------|-----------|--------|
| `default` | ~22 core tools | `TOOL_MODE=default` |
| `full` | ~80+ all tools | `TOOL_MODE=full` |
| `custom` | user-selected | `TOOL_MODE=custom ENABLED_TOOLS=list_projects,get_project` |

Default groups: connection, project, work_package, user, version, search
Full-only groups: membership, time_entry, relation, board, notification, comment, watcher, group, document, query, wiki, placeholder, configuration

## Transports

- **stdio** — requires `OPENPROJECT_URL` + `OPENPROJECT_API_KEY` env vars
- **sse** — port-based, supports per-request auth via `X-OpenProject-URL` / `X-OpenProject-API-Key` headers
- **http** — streamable HTTP, same header-based auth as SSE

## Build & Run

```bash
make build                    # Build binary
make test                     # Run tests
./build/openproject-mcp mcp   # Start MCP server (stdio)
./build/openproject-mcp mcp --transport http --port 8080 --tool-mode full
./build/openproject-mcp project list  # CLI usage
```

## Conventions

- All external types use pointer fields (`*string`, `*int`, `*bool`) — use `derefStr`/`derefInt`/`derefBool` helpers
- `Formattable` type for rich text (description, comment) with `Format` and `Raw` fields
- `FormattableFormat("markdown")` is a named type, not `*string`
- Tool schemas use `jsonschema.Schema` — build with `newSchema()`, `schemaStr()`, etc.
- CLI output uses type switches in `cmd/output.go` — add new types there
