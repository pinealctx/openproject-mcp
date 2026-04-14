package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
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
	addTool(server, "list_versions", "List all versions for a project",
		newSchema(schemaProps{
			"projectId": schemaInt("Project ID"),
		}, "projectId"),
		r.listVersions)

	addTool(server, "create_version", "Create a new version/milestone in a project",
		newSchema(schemaProps{
			"projectId":   schemaInt("Project ID"),
			"name":        schemaStr("Version name"),
			"description": schemaStr("Version description"),
			"startDate":   schemaStr("Start date (YYYY-MM-DD)"),
			"endDate":     schemaStr("End date / due date (YYYY-MM-DD)"),
		}, "projectId", "name"),
		r.createVersion)
}

func (r *Registry) listVersions(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListVersionsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// ListVersionsAvailableInAProject returns versions for a project
	resp, err := r.client.APIClient().ListVersionsAvailableInAProject(ctx, args.ProjectID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list versions: %v", err), nil
	}

	var list external.VersionCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list versions: %v", err), nil
	}

	result := fmt.Sprintf("Found %d versions:\n\n", list.Total)
	for _, v := range list.UnderscoreEmbedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d) — %s\n", v.Name, v.Id, string(v.Status))
	}
	return textResult(result), nil
}

func (r *Registry) createVersion(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateVersionArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.VersionWriteModel{
		Name: strPtr(args.Name),
		UnderscoreLinks: &struct {
			DefiningProject *external.Link `json:"definingProject,omitempty"`
		}{
			DefiningProject: &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/projects/%d", args.ProjectID))},
		},
	}
	if args.Description != "" {
		fmt := external.FormattableFormat("markdown")
		body.Description = &external.Formattable{Format: &fmt, Raw: strPtr(args.Description)}
	}
	if args.StartDate != "" {
		body.StartDate = parseDatePtr(args.StartDate)
	}
	if args.EndDate != "" {
		body.EndDate = parseDatePtr(args.EndDate)
	}

	resp, err := r.client.APIClient().CreateVersion(ctx, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to create version: %v", err), nil
	}
	var version external.VersionReadModel
	if err := openproject.ReadResponse(resp, &version); err != nil {
		return errorResult("Failed to create version: %v", err), nil
	}
	return textResult(fmt.Sprintf("Version #%d created: %s", version.Id, version.Name)), nil
}
