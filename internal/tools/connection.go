package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerConnectionTools registers connection-related tools.
func (r *Registry) registerConnectionTools(server *mcp.Server) {
	// test_connection tool
	server.AddTool(&mcp.Tool{
		Name:        "test_connection",
		Description: "Test the connection to OpenProject and verify authentication",
	}, r.testConnection)

	// check_permissions tool
	server.AddTool(&mcp.Tool{
		Name:        "check_permissions",
		Description: "Check the current user's permissions and access level",
	}, r.checkPermissions)
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
