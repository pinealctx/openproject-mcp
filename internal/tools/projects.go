package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type ListProjectsArgs struct {
	SortBy  string `json:"sortBy,omitempty"`
	Filters string `json:"filters,omitempty"`
}

type GetProjectArgs struct {
	ID int `json:"id"`
}

type CreateProjectArgs struct {
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
	Public      bool   `json:"public,omitempty"`
}

type UpdateProjectArgs struct {
	ID          int    `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Public      *bool  `json:"public,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

type DeleteProjectArgs struct {
	ID int `json:"id"`
}

func (r *Registry) registerProjectTools(server *mcp.Server) {
	addTool(server, "list_projects", "List all projects in OpenProject",
		newSchema(schemaProps{
			"sortBy":  schemaStr(`Sort criteria, e.g. "name:asc"`),
			"filters": schemaStr(`Raw JSON filter string, e.g. [{"active":{"operator":"=","values":["t"]}}]`),
		}),
		r.listProjects)

	addTool(server, "get_project", "Get details of a specific project by ID",
		newSchema(schemaProps{"id": schemaInt("Project ID")}, "id"),
		r.getProject)

	addTool(server, "create_project", "Create a new project in OpenProject",
		newSchema(schemaProps{
			"name":        schemaStr("Project name"),
			"identifier":  schemaStr("Unique project identifier (slug)"),
			"description": schemaStr("Project description"),
			"public":      schemaBool("Whether the project is publicly visible"),
		}, "name", "identifier"),
		r.createProject)

	addTool(server, "update_project", "Update an existing project",
		newSchema(schemaProps{
			"id":          schemaInt("Project ID"),
			"name":        schemaStr("New project name"),
			"description": schemaStr("New project description"),
			"public":      schemaBool("Whether the project is publicly visible"),
			"active":      schemaBool("Whether the project is active"),
		}, "id"),
		r.updateProject)

	addTool(server, "delete_project", "Delete a project from OpenProject",
		newSchema(schemaProps{"id": schemaInt("Project ID")}, "id"),
		r.deleteProject)
}

func (r *Registry) listProjects(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListProjectsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	projects, err := r.client.ListProjects(ctx, openproject.ProjectListInput{
		SortBy:  normalizeSortBy(args.SortBy),
		Filters: args.Filters,
	})
	if err != nil {
		return errorResult("Failed to list projects: %v", err), nil
	}

	result := fmt.Sprintf("Found %d projects:\n\n", projects.Total)
	for _, p := range projects.UnderscoreEmbedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", derefStr(p.Name), derefInt(p.Id))
		result += fmt.Sprintf("  Identifier: %s\n", derefStr(p.Identifier))
		if p.Description != nil && p.Description.Raw != nil && *p.Description.Raw != "" {
			result += fmt.Sprintf("  Description: %s\n", *p.Description.Raw)
		}
		result += "\n"
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) getProject(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetProjectArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	project, err := r.client.GetProject(ctx, args.ID)
	if err != nil {
		return errorResult("Failed to get project: %v", err), nil
	}
	return textResult(formatProject(project)), nil
}

func (r *Registry) createProject(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateProjectArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	project, err := r.client.CreateProject(ctx, openproject.ProjectCreateInput{
		Name:        args.Name,
		Identifier:  args.Identifier,
		Description: args.Description,
		Public:      args.Public,
	})
	if err != nil {
		return errorResult("Failed to create project: %v", err), nil
	}

	result := "Project created successfully!\n\n"
	result += fmt.Sprintf("- **ID:** %d\n", derefInt(project.Id))
	result += fmt.Sprintf("- **Name:** %s\n", derefStr(project.Name))
	result += fmt.Sprintf("- **Identifier:** %s\n", derefStr(project.Identifier))
	return textResult(result), nil
}

func (r *Registry) updateProject(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateProjectArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	project, err := r.client.UpdateProject(ctx, openproject.ProjectUpdateInput{
		ID:          args.ID,
		Name:        args.Name,
		Description: args.Description,
		Public:      args.Public,
		Active:      args.Active,
	})
	if err != nil {
		return errorResult("Failed to update project: %v", err), nil
	}
	return textResult(fmt.Sprintf("Project %d updated successfully!\n\nName: %s", derefInt(project.Id), derefStr(project.Name))), nil
}

func (r *Registry) deleteProject(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteProjectArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	if err := r.client.DeleteProject(ctx, args.ID); err != nil {
		return errorResult("Failed to delete project: %v", err), nil
	}
	return textResult(fmt.Sprintf("Project %d deleted successfully!", args.ID)), nil
}
