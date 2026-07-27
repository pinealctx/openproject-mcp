# openproject-mcp

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server and command-line client for [OpenProject](https://www.openproject.org/), written in Go. It covers projects, work packages, comments, attachments, users, memberships, versions, relations, search, watchers, and notifications.

Built with the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk).

## Installation

**Download a pre-built binary (recommended):**

Grab the latest release for your platform from [GitHub Releases](https://github.com/pinealctx/openproject-mcp/releases).

**Via `go install` (developer fallback):**

```bash
go install github.com/pinealctx/openproject-mcp@latest
```

**Build from source:**

```bash
git clone https://github.com/pinealctx/openproject-mcp.git
cd openproject-mcp
make build   # output: ./build/openproject-mcp
```

## Quick Start

```bash
# Build
go build -o openproject-mcp .

# Run in stdio mode (Claude Desktop / MCP clients)
OPENPROJECT_URL=https://your-instance.openproject.com \
OPENPROJECT_API_KEY=your-api-key \
openproject-mcp
```

## Usage Modes

This tool operates in two modes:

### 1. MCP Server Mode (default)

Starts an MCP server for AI assistants. No subcommand needed — just run the binary.

```bash
# stdio mode (default) - for Claude Desktop, Cursor, etc.
openproject-mcp

# SSE mode - for web-based MCP clients
openproject-mcp mcp -t sse -p 3000

# HTTP mode - for HTTP-based MCP clients
openproject-mcp mcp -t http -p 8080
```

### 2. CLI Mode

Direct command-line interaction with OpenProject API. Useful for scripting and automation.

```bash
# List all projects
openproject-mcp project list

# Get project details
openproject-mcp project get 42

# Create a new project
openproject-mcp project create -n "My Project" -i "my-project"

# List work packages in a project
openproject-mcp wp list -p 42

# Create a work package with a current worker and delivery owner
openproject-mcp wp create -p 42 -s "Implement feature X" --assignee 5 --accountable 8

# Add a comment to a work package
openproject-mcp wp comment 123 -m "Investigated and updated the deployment notes."

# List and download work package attachments
openproject-mcp attachment list 123
openproject-mcp attachment download-all 123

# Search across OpenProject
openproject-mcp search "bug"

# Output as JSON for scripting
openproject-mcp project list -o json
```

## CLI Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `project` | `proj`, `p` | Manage projects |
| `work-package` | `wp` | Manage work packages (tasks, bugs, features) |
| `attachment` | `attachments` | Manage work package attachments |
| `user` | `u` | Manage users |
| `membership` | `member`, `m` | Manage project memberships |
| `notification` | `notify` | Manage notifications |
| `search` | - | Search across projects, work packages, users |
| `status` | - | List work package statuses |
| `priority` | `priorities` | List work package priorities |
| `type` | `types` | List work package types |
| `role` | `roles` | List user roles |
| `version` | - | Manage project versions/milestones |
| `mcp` | - | Start MCP server |

### CLI Examples

```bash
# === Projects ===
openproject-mcp project list
openproject-mcp project get 42
openproject-mcp project create -n "Website Redesign" -i "website-redesign"
openproject-mcp project update 42 -n "New Name"
openproject-mcp project delete 42

# === Work Packages ===
openproject-mcp wp list -p 42
openproject-mcp wp get 123
openproject-mcp wp available-assignees 123
openproject-mcp wp create -p 42 -s "Fix login bug" -d "Description here"
openproject-mcp wp update 123 --status "In Progress" --progress 50
# Assignee is the current worker; Accountable is the delivery owner
openproject-mcp wp update 123 --assignee 5
openproject-mcp wp update 123 --accountable 8
openproject-mcp wp update 123 --clear-assignee
openproject-mcp wp update 123 --clear-accountable
openproject-mcp wp activities 123
openproject-mcp wp comment 123 -m "Investigated and updated the deployment notes."
openproject-mcp wp delete 123

# === Attachments ===
openproject-mcp attachment list 123 -o json
openproject-mcp attachment get 456 -o json
openproject-mcp attachment upload 123 ./report.pdf --description "Release report"
openproject-mcp attachment download 456 --destination ./report.pdf
openproject-mcp attachment download-all 123 --directory ./ticket-123-files
openproject-mcp attachment delete 456 --yes

# === Work Package Relations ===
openproject-mcp wp set-parent 123 -p 100
openproject-mcp wp relation create --from 123 --to 456 --type blocks

# === Users ===
openproject-mcp user list
openproject-mcp user get 5
openproject-mcp user me

# === Memberships ===
openproject-mcp membership list -p 42
openproject-mcp membership create -p 42 -u 5 -r "3,4"
openproject-mcp membership delete 123

# === Search ===
openproject-mcp search "bug"
openproject-mcp search "website" -t project
openproject-mcp search "john" -t user

# === Notifications ===
openproject-mcp notification list
openproject-mcp notification list -u      # unread only
openproject-mcp notification read-all
```

Run `openproject-mcp [command] --help` for detailed usage of each command.

## Transport Modes

| Mode | Command | Use Case |
|------|---------|----------|
| `stdio` | `openproject-mcp` or `openproject-mcp mcp -t stdio` | Claude Desktop, Cursor, Continue (default) |
| `sse` | `openproject-mcp mcp -t sse -p 3000` | Server-Sent Events for web-based MCP clients |
| `http` | `openproject-mcp mcp -t http -p 8080` | Streamable HTTP for web-based clients |

### stdio Mode (Default)

Communication via stdin/stdout. The server reads JSON-RPC messages from stdin and writes responses to stdout. All logging is suppressed to avoid protocol interference.

**Compatible with:** Claude Desktop, Cursor, Continue, Zed, and other MCP clients that spawn the server as a subprocess.

```bash
# Set environment variables
export OPENPROJECT_URL="https://your-instance.openproject.com"
export OPENPROJECT_API_KEY="your-api-token"

# Start MCP server (all three commands are equivalent)
openproject-mcp
openproject-mcp mcp
openproject-mcp mcp -t stdio
```

### SSE / HTTP Mode

For web-based MCP clients. Requires a port to listen on.

```bash
# SSE mode
openproject-mcp mcp -t sse -p 3000

# HTTP mode
openproject-mcp mcp -t http -p 8080
```

## Credential Modes

**Mode 1 — Server pre-configured (single-tenant / self-hosted)**

Set credentials via environment variables at startup. All requests share the same OpenProject account.

```bash
export OPENPROJECT_URL=https://your-instance.openproject.com
export OPENPROJECT_API_KEY=your-api-key
openproject-mcp mcp -t http -p 8080
```

> For `stdio` transport, credentials are **required** at startup.

**Mode 2 — Per-request credentials (multi-tenant / service)**

Start without credentials; each HTTP/SSE client supplies its own via request headers.

```bash
openproject-mcp mcp -t http -p 8080
# no env vars needed
```

Clients include headers with every request:
```
X-OpenProject-URL: https://your-instance.openproject.com
X-OpenProject-API-Key: your-api-key
```

If neither server-level nor per-request credentials are available, the server responds with `HTTP 401`.

## Environment Variables

| Variable | Required (stdio) | Description |
|---|---|---|
| `OPENPROJECT_URL` | Yes | OpenProject instance URL |
| `OPENPROJECT_API_KEY` | Yes | API key for authentication |
| `OPENPROJECT_PROXY` | No | HTTP/HTTPS/SOCKS5 proxy URL |
| `OPENPROJECT_TRANSFER_TIMEOUT` | No | Upload/download timeout as a Go duration (default: `5m`) |
| `LOG_LEVEL` | No | `debug`, `info`, `warn`, `error` (default: `info`) |
| `TRANSPORT` | No | `stdio`, `sse`, `http` (default: `stdio`) |
| `PORT` | No | Port for SSE/HTTP transport (default: `8080`) |

## MCP Client Configuration

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "openproject": {
      "command": "/path/to/openproject-mcp",
      "args": [],
      "env": {
        "OPENPROJECT_URL": "https://your-instance.openproject.com",
        "OPENPROJECT_API_KEY": "your-api-key"
      }
    }
  }
}
```

### Cursor

Add to `.cursor/mcp.json`

```json
{
  "mcpServers": {
    "openproject": {
      "command": "/path/to/openproject-mcp",
      "args": [],
      "env": {
        "OPENPROJECT_URL": "https://your-instance.openproject.com",
        "OPENPROJECT_API_KEY": "your-api-key"
      }
    }
  }
}
```

### VS Code / HTTP Client

```json
{
  "servers": {
    "openproject": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
  }
}
```

### Zed

Add to your Zed settings:

```json
{
  "context_servers": {
    "openproject": {
      "command": "/path/to/openproject-mcp",
      "args": [],
      "env": {
        "OPENPROJECT_URL": "https://your-instance.openproject.com",
        "OPENPROJECT_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Tools (60+)

### Connection & Auth
| Tool | Description |
|------|-------------|
| `test_connection` | Test connectivity and verify authentication |
| `check_permissions` | Check current user permissions |
| `get_current_user` | Get the authenticated user's full profile |
| `get_api_info` | Get OpenProject API root information |

### Projects
| Tool | Description |
|------|-------------|
| `list_projects` | List projects with optional filters |
| `get_project` | Get a project by ID |
| `create_project` | Create a new project |
| `update_project` | Update a project |
| `delete_project` | Delete a project |

### Work Packages
| Tool | Description |
|------|-------------|
| `list_work_packages` | List work packages with filters |
| `list_project_work_packages` | List work packages in a specific project |
| `get_work_package` | Get a work package by ID |
| `create_work_package` | Create a work package, optionally setting Assignee and Accountable |
| `update_work_package` | Update a work package, including setting or clearing Assignee and Accountable (auto-fetches lockVersion) |
| `delete_work_package` | Delete a work package |
| `list_work_package_activities` | List activities and comments for a work package |
| `create_work_package_comment` | Add a comment to a work package |
| `list_work_package_attachments` | List file attachment metadata for a work package |
| `get_attachment` | Get attachment metadata without downloading file content |
| `list_types` | List available work package types |
| `list_statuses` | List available work package statuses |
| `list_priorities` | List available work package priorities |
| `list_available_assignees` | List users available for the Assignee field |

### Work Package Hierarchy & Relations
| Tool | Description |
|------|-------------|
| `set_work_package_parent` | Set the parent of a work package |
| `remove_work_package_parent` | Remove the parent relationship |
| `list_work_package_children` | List child work packages |
| `create_work_package_relation` | Create a relation between two work packages |
| `list_work_package_relations` | List all relations for a work package |
| `get_work_package_relation` | Get a specific relation |
| `update_work_package_relation` | Update a relation |
| `delete_work_package_relation` | Delete a relation |

### Users
| Tool | Description |
|------|-------------|
| `list_users` | List users |
| `get_user` | Get a user by ID |

### Memberships
| Tool | Description |
|------|-------------|
| `list_memberships` | List memberships |
| `get_membership` | Get a membership by ID |
| `create_membership` | Add a user to a project |
| `update_membership` | Update a membership's roles |
| `delete_membership` | Remove a user from a project |
| `list_project_members` | List members of a specific project |

### Versions
| Tool | Description |
|------|-------------|
| `list_versions` | List versions in a project |
| `create_version` | Create a new version |

### Notifications
| Tool | Description |
|------|-------------|
| `list_notifications` | List notifications for the current user |
| `mark_notification_read` | Mark a notification as read |
| `mark_all_notifications_read` | Mark all notifications as read |

### Search
| Tool | Description |
|------|-------------|
| `search` | Search across projects, work packages, and users |

## Build

```bash
make build          # Build for current platform → ./build/openproject-mcp
make build-all      # Cross-compile for Linux / macOS / Windows
make test           # Run tests
make clean          # Remove build artifacts
```

## License

MIT
