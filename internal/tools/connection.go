package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerConnectionTools registers connection-related tools.
func (r *Registry) registerConnectionTools(server *mcp.Server) {
	addTool(server, "test_connection",
		"Test the connection to OpenProject and verify authentication",
		noSchema, r.testConnection)

	addTool(server, "check_permissions",
		"Check the current user's permissions and access level",
		noSchema, r.checkPermissions)

	addTool(server, "get_current_user",
		"Get full profile of the currently authenticated user",
		noSchema, r.getCurrentUser)

	addTool(server, "get_api_info",
		"Get OpenProject API root information (version, links, etc.)",
		noSchema, r.getAPIInfo)
}

// testConnection tests the connection to OpenProject.
func (r *Registry) testConnection(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user, err := r.client.TestConnection(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to connect to OpenProject: %v", err)}},
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Successfully connected to OpenProject as %s (%s)", user.Name, user.Email)}},
	}, nil
}

// checkPermissions checks the current user's permissions.
func (r *Registry) checkPermissions(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user, err := r.client.GetCurrentUser(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get current user: %v", err)}},
		}, nil
	}

	result := fmt.Sprintf("User: %s\n", user.Name)
	result += fmt.Sprintf("Email: %s\n", user.Email)
	result += fmt.Sprintf("Admin: %v\n", user.Admin)
	result += fmt.Sprintf("Status: %s\n", user.Status)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

// getCurrentUser returns full profile of the authenticated user.
func (r *Registry) getCurrentUser(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user, err := r.client.GetCurrentUser(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get current user: %v", err)}},
		}, nil
	}

	result := fmt.Sprintf("# %s\n\n", user.Name)
	result += fmt.Sprintf("- **ID:** %d\n", user.ID)
	result += fmt.Sprintf("- **Login:** %s\n", user.Login)
	result += fmt.Sprintf("- **Email:** %s\n", user.Email)
	result += fmt.Sprintf("- **First Name:** %s\n", user.FirstName)
	result += fmt.Sprintf("- **Last Name:** %s\n", user.LastName)
	result += fmt.Sprintf("- **Admin:** %v\n", user.Admin)
	result += fmt.Sprintf("- **Status:** %s\n", user.Status)
	result += fmt.Sprintf("- **Language:** %s\n", user.Language)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

// getAPIInfo returns OpenProject API root information.
func (r *Registry) getAPIInfo(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	info, err := r.client.GetAPIRoot(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get API info: %v", err)}},
		}, nil
	}

	result := "OpenProject API Information:\n\n"
	for k, v := range info {
		if k == "_links" || k == "_embedded" || k == "_type" {
			continue
		}
		result += fmt.Sprintf("- **%s:** %v\n", k, v)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}
