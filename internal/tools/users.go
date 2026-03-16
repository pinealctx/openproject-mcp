package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type ListUsersArgs struct {
	Offset   int    `json:"offset,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	SortBy   string `json:"sortBy,omitempty"`
	OrderBy  string `json:"orderBy,omitempty"`
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
	parseArgs(req.Params.Arguments, &args)

	opts := &openproject.ListUsersOptions{Offset: args.Offset, PageSize: args.PageSize, SortBy: firstNonEmpty(args.SortBy, args.OrderBy)}
	users, err := r.client.ListUsers(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list users: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d users:\n\n", users.Total)
	for _, u := range users.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d) — %s", u.Name, u.ID, u.Email)
		if u.Admin {
			result += " [Admin]"
		}
		result += "\n"
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) getUser(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetUserArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	user, err := r.client.GetUser(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get user: %v", err)}}}, nil
	}

	result := fmt.Sprintf("# %s\n\n", user.Name)
	result += fmt.Sprintf("- **ID:** %d\n", user.ID)
	result += fmt.Sprintf("- **Login:** %s\n", user.Login)
	result += fmt.Sprintf("- **Email:** %s\n", user.Email)
	result += fmt.Sprintf("- **Admin:** %v\n", user.Admin)
	result += fmt.Sprintf("- **Status:** %s\n", user.Status)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}
