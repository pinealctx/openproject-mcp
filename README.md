# OpenProject MCP Server

A Go implementation of the Model Context Protocol (MCP) server for [OpenProject](https://www.openproject.org/). Provides 41 tools for managing projects, work packages, users, memberships, time entries, versions, and relations through the OpenProject API v3.

Built with the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk).

## Quick Start

```bash
# Build
go build -o openproject-mcp ./cmd/openproject-mcp

# Configure
export OPENPROJECT_URL=https://your-instance.openproject.com
export OPENPROJECT_API_KEY=your-api-key

# Run (stdio mode, for Claude Desktop / MCP clients)
./openproject-mcp
```

## Transport Modes

```bash
./openproject-mcp                      # stdio (default)
./openproject-mcp -transport=sse       # SSE on :8080
./openproject-mcp -transport=http      # Streamable HTTP on :8080
./openproject-mcp -transport=sse -port=3000  # SSE on custom port
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `OPENPROJECT_URL` | Yes | OpenProject instance URL |
| `OPENPROJECT_API_KEY` | Yes | API key for authentication |
| `OPENPROJECT_PROXY` | No | Proxy URL (also reads `HTTP_PROXY`/`HTTPS_PROXY`) |
| `LOG_LEVEL` | No | Log level: debug, info, warn, error (default: info) |
| `TRANSPORT` | No | Transport type: stdio, sse, http (default: stdio) |
| `PORT` | No | Port for SSE/HTTP transport (default: 8080) |

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

## Tools (41)

| Category | Tools |
|---|---|
| Connection | `test_connection`, `check_permissions` |
| Projects | `list_projects`, `get_project`, `create_project`, `update_project`, `delete_project` |
| Work Packages | `list_work_packages`, `get_work_package`, `create_work_package`, `update_work_package`, `delete_work_package`, `list_types`, `list_statuses`, `list_priorities` |
| Hierarchy | `set_work_package_parent`, `remove_work_package_parent`, `list_work_package_children` |
| Relations | `create_work_package_relation`, `list_work_package_relations`, `get_work_package_relation`, `update_work_package_relation`, `delete_work_package_relation` |
| Users | `list_users`, `get_user` |
| Memberships | `list_memberships`, `get_membership`, `create_membership`, `update_membership`, `delete_membership`, `list_project_members`, `list_roles`, `get_role` |
| Time Entries | `list_time_entries`, `create_time_entry`, `update_time_entry`, `delete_time_entry`, `list_time_entry_activities` |
| Versions | `list_versions`, `create_version` |

## Build

```bash
make build          # Build for current platform
make build-all      # Cross-compile for Linux/macOS/Windows
make test           # Run tests
make clean          # Clean build artifacts
```

## License

MIT
