# openproject-mcp

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for [OpenProject](https://www.openproject.org/), written in Go. Exposes **48 tools** covering projects, work packages, users, memberships, time entries, versions, relations, boards, search, and notifications — enabling AI assistants such as Claude and GitHub Copilot to manage OpenProject directly.

Built with the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk).

## Installation

**Via `go install` (recommended):**

```bash
go install github.com/pinealctx/openproject-mcp@latest
```

**Download a pre-built binary:**

Grab the latest release for your platform from [GitHub Releases](https://github.com/pinealctx/openproject-mcp/releases).

**Build from source:**

```bash
git clone https://github.com/pinealctx/openproject-mcp.git
cd openproject-mcp
make build   # output: ./build/openproject-mcp
```

## Quick Start

```bash
# Build
go build -o openproject-mcp ./cmd/openproject-mcp

# Run in stdio mode (Claude Desktop / MCP clients)
OPENPROJECT_URL=https://your-instance.openproject.com \
OPENPROJECT_API_KEY=your-api-key \
./openproject-mcp
```

## Transport Modes

| Mode | Command | Use case |
|------|---------|----------|
| `stdio` | `./openproject-mcp` (default) | Claude Desktop, single-user |
| `sse` | `./openproject-mcp -transport sse -port 3000` | SSE-based MCP clients |
| `http` | `./openproject-mcp -transport http -port 8080` | Streamable HTTP, multi-tenant |

## Credential Modes

**Mode 1 — Server pre-configured (single-tenant / self-hosted)**

Set credentials via environment variables at startup. All requests share the same OpenProject account.

```bash
export OPENPROJECT_URL=https://your-instance.openproject.com
export OPENPROJECT_API_KEY=your-api-key
./openproject-mcp -transport http -port 8080
```

> For `stdio` transport, credentials are **required** at startup.

**Mode 2 — Per-request credentials (multi-tenant / service)**

Start without credentials; each HTTP/SSE client supplies its own via request headers.

```bash
./openproject-mcp -transport http -port 8080
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
| `LOG_LEVEL` | No | `debug`, `info`, `warn`, `error` (default: `info`) |
| `TRANSPORT` | No | `stdio`, `sse`, `http` (default: `stdio`) |
| `PORT` | No | Port for SSE/HTTP transport (default: `8080`) |

## Claude Desktop Configuration

```json
{
  "mcpServers": {
    "openproject": {
      "command": "/path/to/openproject-mcp",
      "env": {
        "OPENPROJECT_URL": "https://your-instance.openproject.com",
        "OPENPROJECT_API_KEY": "your-api-key"
      }
    }
  }
}
```

## VS Code / HTTP Client Configuration

```json
{
  "servers": {
    "openproject": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

## Tools (48)

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
| `create_work_package` | Create a work package |
| `update_work_package` | Update a work package (auto-fetches lockVersion) |
| `delete_work_package` | Delete a work package |
| `move_work_package` | Move a work package to another project |
| `watch_work_package` | Watch a work package for notifications |
| `list_types` | List available work package types |
| `list_statuses` | List available work package statuses |
| `list_priorities` | List available work package priorities |

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

### Time Tracking
| Tool | Description |
|------|-------------|
| `list_time_entries` | List time entries with optional filters |
| `create_time_entry` | Log time on a work package |
| `update_time_entry` | Update a time entry |
| `delete_time_entry` | Delete a time entry |
| `list_time_entry_activities` | List available activity types |

### Versions
| Tool | Description |
|------|-------------|
| `list_versions` | List versions in a project |
| `create_version` | Create a new version |

### Boards (Kanban)
| Tool | Description |
|------|-------------|
| `get_boards` | List boards, optionally filtered by project |
| `get_board` | Get a board by ID |
| `create_board` | Create a new board |
| `update_board` | Update a board |
| `delete_board` | Delete a board |
| `add_board_widget` | Add a widget (column) to a board |
| `remove_board_widget` | Remove a widget from a board |

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
