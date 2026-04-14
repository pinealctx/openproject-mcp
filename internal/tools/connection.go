package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

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

func (r *Registry) testConnection(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user, err := r.client.GetCurrentUser(ctx)
	if err != nil {
		return errorResult("Failed to connect to OpenProject: %v", err), nil
	}
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	return textResult(fmt.Sprintf("Successfully connected to OpenProject as %s (%s)", user.Name, email)), nil
}

func (r *Registry) checkPermissions(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user, err := r.client.GetCurrentUser(ctx)
	if err != nil {
		return errorResult("Failed to get current user: %v", err), nil
	}

	result := fmt.Sprintf("User: %s\n", user.Name)
	if user.Email != nil {
		result += fmt.Sprintf("Email: %s\n", *user.Email)
	}
	if user.Admin != nil {
		result += fmt.Sprintf("Admin: %v\n", *user.Admin)
	}
	if user.Status != nil {
		result += fmt.Sprintf("Status: %s\n", *user.Status)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) getCurrentUser(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user, err := r.client.GetCurrentUser(ctx)
	if err != nil {
		return errorResult("Failed to get current user: %v", err), nil
	}
	return textResult(formatUser(user)), nil
}

func (r *Registry) getAPIInfo(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := r.client.APIClient().ViewRoot(ctx)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get API info: %v", err), nil
	}
	var root external.RootModel
	if err := openproject.ReadResponse(resp, &root); err != nil {
		return errorResult("Failed to parse API info: %v", err), nil
	}

	result := "OpenProject API Information:\n\n"
	result += fmt.Sprintf("- **Instance Name:** %s\n", root.InstanceName)
	if root.CoreVersion != nil {
		result += fmt.Sprintf("- **Core Version:** %s\n", *root.CoreVersion)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}
