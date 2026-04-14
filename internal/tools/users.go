package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

type ListUsersArgs struct {
	Offset   int    `json:"offset,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	SortBy   string `json:"sortBy,omitempty"`
}

type GetUserArgs struct {
	ID int `json:"id"`
}

// registerUserTools registers user-related tools.
func (r *Registry) registerUserTools(server *mcp.Server) {
	addTool(server, "list_users", "List all users in OpenProject",
		newSchema(schemaProps{
			"offset":   schemaInt("Pagination offset (default: 0)"),
			"pageSize": schemaInt("Number of items per page (default: 20)"),
			"sortBy":   schemaStr(`Sort criteria, e.g. "name:asc"`),
		}),
		r.listUsers)

	addTool(server, "get_user", "Get details of a specific user by ID",
		newSchema(schemaProps{
			"id": schemaInt("User ID"),
		}, "id"),
		r.getUser)
}

func (r *Registry) listUsers(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListUsersArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListUsersParams{}
	if args.Offset > 0 {
		params.Offset = intPtr(args.Offset)
	}
	if args.PageSize > 0 {
		params.PageSize = intPtr(args.PageSize)
	}
	if args.SortBy != "" {
		params.SortBy = strPtr(normalizeSortBy(args.SortBy))
	}

	resp, err := r.client.APIClient().ListUsers(ctx, params)
	if err != nil {
		return errorResult("Failed to list users: %v", err), nil
	}

	var list external.UserCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list users: %v", err), nil
	}

	result := fmt.Sprintf("Found %d users:\n\n", list.Total)
	for _, u := range list.UnderscoreEmbedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d) — %s", u.Name, u.Id, derefStr(u.Email))
		if derefBool(u.Admin) {
			result += " [Admin]"
		}
		result += "\n"
	}
	return textResult(result), nil
}

func (r *Registry) getUser(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetUserArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// ViewUser takes string ID
	resp, err := r.client.APIClient().ViewUser(ctx, fmt.Sprintf("%d", args.ID))
	if err != nil {
		return errorResult("Failed to get user: %v", err), nil
	}
	var user external.UserModel
	if err := openproject.ReadResponse(resp, &user); err != nil {
		return errorResult("Failed to get user: %v", err), nil
	}

	return textResult(formatUser(&user)), nil
}
