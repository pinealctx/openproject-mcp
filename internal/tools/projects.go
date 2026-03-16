package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

// ListProjectsArgs represents arguments for the list_projects tool.
type ListProjectsArgs struct {
	Offset   int    `json:"offset,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	SortBy   string `json:"sortBy,omitempty"`
	Filters  string `json:"filters,omitempty"`
}

// GetProjectArgs represents arguments for the get_project tool.
type GetProjectArgs struct {
	ID int `json:"id"`
}

// CreateProjectArgs represents arguments for the create_project tool.
type CreateProjectArgs struct {
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
	Public      bool   `json:"public,omitempty"`
}

// UpdateProjectArgs represents arguments for the update_project tool.
type UpdateProjectArgs struct {
	ID          int    `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Public      *bool  `json:"public,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

// DeleteProjectArgs represents arguments for the delete_project tool.
type DeleteProjectArgs struct {
	ID int `json:"id"`
}

// registerProjectTools registers project-related tools.
func (r *Registry) registerProjectTools(server *mcp.Server) {
	addTool(server, "list_projects", "List all projects in OpenProject",
		newSchema(schemaProps{
			"offset":   schemaInt("Pagination offset (default: 0)"),
			"pageSize": schemaInt("Number of items per page (default: 20)"),
			"sortBy":   schemaStr(`Sort criteria, e.g. "name:asc"`),
			"filters":  schemaStr("Raw JSON filter string, e.g. [{\"active\":{\"operator\":\"=\",\"values\":[\"t\"]}}]"),
		}),
		r.listProjects)

	addTool(server, "get_project", "Get details of a specific project by ID",
		newSchema(schemaProps{
			"id": schemaInt("Project ID"),
		}, "id"),
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
		newSchema(schemaProps{
			"id": schemaInt("Project ID"),
		}, "id"),
		r.deleteProject)
}

// parseArgs parses arguments from the request.
func parseArgs(args any, target any) error {
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// listProjects lists all projects.
func (r *Registry) listProjects(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListProjectsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.ListProjectsOptions{
		Offset:     args.Offset,
		PageSize:   args.PageSize,
		SortBy:     args.SortBy,
		RawFilters: args.Filters,
	}

	projects, err := r.client.ListProjects(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list projects: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d projects:\n\n", projects.Total)
	for _, p := range projects.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", p.Name, p.ID)
		result += fmt.Sprintf("  Identifier: %s\n", p.Identifier)
		if p.Description.Raw != "" {
			result += fmt.Sprintf("  Description: %s\n", p.Description.Raw)
		}
		result += "\n"
	}

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

// getProject gets a project by ID.
func (r *Registry) getProject(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetProjectArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	project, err := r.client.GetProject(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get project: %v", err)}}}, nil
	}

	result := fmt.Sprintf("# %s\n\n", project.Name)
	result += fmt.Sprintf("- **ID:** %d\n", project.ID)
	result += fmt.Sprintf("- **Identifier:** %s\n", project.Identifier)
	result += fmt.Sprintf("- **Active:** %v\n", project.Active)
	result += fmt.Sprintf("- **Public:** %v\n", project.Public)
	if project.Description.Raw != "" {
		result += fmt.Sprintf("\n## Description\n%s\n", project.Description.Raw)
	}

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

// createProject creates a new project.
func (r *Registry) createProject(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateProjectArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.CreateProjectOptions{
		Name:        args.Name,
		Identifier:  args.Identifier,
		Description: args.Description,
		Public:      args.Public,
	}

	project, err := r.client.CreateProject(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create project: %v", err)}}}, nil
	}

	result := "Project created successfully!\n\n"
	result += fmt.Sprintf("- **ID:** %d\n", project.ID)
	result += fmt.Sprintf("- **Name:** %s\n", project.Name)
	result += fmt.Sprintf("- **Identifier:** %s\n", project.Identifier)

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

// updateProject updates an existing project.
func (r *Registry) updateProject(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateProjectArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.UpdateProjectOptions{
		Name:        args.Name,
		Description: args.Description,
		Public:      args.Public,
		Active:      args.Active,
	}

	project, err := r.client.UpdateProject(ctx, args.ID, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update project: %v", err)}}}, nil
	}

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Project %d updated successfully!\n\nName: %s", project.ID, project.Name)}}}, nil
}

// deleteProject deletes a project.
func (r *Registry) deleteProject(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteProjectArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	err := r.client.DeleteProject(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete project: %v", err)}}}, nil
	}

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Project %d deleted successfully!", args.ID)}}}, nil
}
