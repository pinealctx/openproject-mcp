package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type ListVersionsArgs struct{ ProjectID int }
type CreateVersionArgs struct {
	ProjectID   int    `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	EndDate     string `json:"endDate,omitempty"`
}

// registerVersionTools registers version-related tools.
func (r *Registry) registerVersionTools(server *mcp.Server) {
	server.AddTool(&mcp.Tool{Name: "list_versions", Description: "List all versions for a project"}, r.listVersions)
	server.AddTool(&mcp.Tool{Name: "create_version", Description: "Create a new version/milestone in a project"}, r.createVersion)
}

func (r *Registry) listVersions(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListVersionsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	list, err := r.client.ListVersions(ctx, args.ProjectID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list versions: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d versions:\n\n", list.Total)
	for _, v := range list.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d) - %s\n", v.Name, v.ID, v.Status)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) createVersion(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateVersionArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.CreateVersionOptions{
		Name: args.Name, Description: args.Description, StartDate: args.StartDate, EndDate: args.EndDate, ProjectID: args.ProjectID,
	}
	version, err := r.client.CreateVersion(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create version: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Version #%d created: %s", version.ID, version.Name)}}}, nil
}
