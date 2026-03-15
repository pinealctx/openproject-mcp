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
	OrderBy  string `json:"orderBy,omitempty"`
}

type GetUserArgs struct {
	ID int `json:"id"`
}

// registerUserTools registers user-related tools.
func (r *Registry) registerUserTools(server *mcp.Server) {
	server.AddTool(&mcp.Tool{Name: "list_users", Description: "List all users in OpenProject"}, r.listUsers)
	server.AddTool(&mcp.Tool{Name: "get_user", Description: "Get details of a specific user by ID"}, r.getUser)
}

func (r *Registry) listUsers(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListUsersArgs
	parseArgs(req.Params.Arguments, &args)

	opts := &openproject.ListUsersOptions{Offset: args.Offset, PageSize: args.PageSize, OrderBy: args.OrderBy}
	users, err := r.client.ListUsers(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list users: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d users:\n\n", users.Total)
	for _, u := range users.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", u.Name, u.ID)
		result += fmt.Sprintf("  Email: %s\n", u.Email)
		if u.Admin {
			result += "  (Admin)\n"
		}
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
	result += fmt.Sprintf("- **Email:** %s\n", user.Email)
	result += fmt.Sprintf("- **Admin:** %v\n", user.Admin)
	result += fmt.Sprintf("- **Status:** %s\n", user.Status)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}
