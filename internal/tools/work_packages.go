package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

// Work package argument types
type ListWorkPackagesArgs struct {
	ProjectID int    `json:"projectId,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
	OrderBy   string `json:"orderBy,omitempty"`
}

type GetWorkPackageArgs struct {
	ID int `json:"id"`
}

type CreateWorkPackageArgs struct {
	ProjectID   int    `json:"projectId"`
	Subject     string `json:"subject"`
	Description string `json:"description,omitempty"`
	TypeID      int    `json:"typeId,omitempty"`
	StatusID    int    `json:"statusId,omitempty"`
	PriorityID  int    `json:"priorityId,omitempty"`
	AssigneeID  int    `json:"assigneeId,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	DueDate     string `json:"dueDate,omitempty"`
}

type UpdateWorkPackageArgs struct {
	ID          int    `json:"id"`
	Subject     string `json:"subject,omitempty"`
	Description string `json:"description,omitempty"`
	StatusID    int    `json:"statusId,omitempty"`
	PriorityID  int    `json:"priorityId,omitempty"`
	AssigneeID  int    `json:"assigneeId,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	DueDate     string `json:"dueDate,omitempty"`
}

type DeleteWorkPackageArgs struct {
	ID int `json:"id"`
}

type ListTypesArgs struct {
	ProjectID int `json:"projectId,omitempty"`
}

type ListStatusesArgs struct{}
type ListPrioritiesArgs struct{}

// registerWorkPackageTools registers work package-related tools.
func (r *Registry) registerWorkPackageTools(server *mcp.Server) {
	server.AddTool(&mcp.Tool{Name: "list_work_packages", Description: "List work packages, optionally filtered by project"}, r.listWorkPackages)
	server.AddTool(&mcp.Tool{Name: "get_work_package", Description: "Get details of a specific work package by ID"}, r.getWorkPackage)
	server.AddTool(&mcp.Tool{Name: "create_work_package", Description: "Create a new work package in a project"}, r.createWorkPackage)
	server.AddTool(&mcp.Tool{Name: "update_work_package", Description: "Update an existing work package"}, r.updateWorkPackage)
	server.AddTool(&mcp.Tool{Name: "delete_work_package", Description: "Delete a work package"}, r.deleteWorkPackage)
	server.AddTool(&mcp.Tool{Name: "list_types", Description: "List work package types"}, r.listTypes)
	server.AddTool(&mcp.Tool{Name: "list_statuses", Description: "List all work package statuses"}, r.listStatuses)
	server.AddTool(&mcp.Tool{Name: "list_priorities", Description: "List all work package priorities"}, r.listPriorities)
}

func (r *Registry) listWorkPackages(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackagesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.ListWorkPackagesOptions{Offset: args.Offset, PageSize: args.PageSize, OrderBy: args.OrderBy}
	var list *openproject.WorkPackageList
	var err error

	if args.ProjectID > 0 {
		list, err = r.client.ListProjectWorkPackages(ctx, args.ProjectID, opts)
	} else {
		list, err = r.client.ListWorkPackages(ctx, opts)
	}
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list work packages: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d work packages:\n\n", list.Total)
	for _, wp := range list.Embedded.Elements {
		result += fmt.Sprintf("- **#%d %s**\n", wp.ID, wp.Subject)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) getWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	wp, err := r.client.GetWorkPackage(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get work package: %v", err)}}}, nil
	}

	result := fmt.Sprintf("# #%d %s\n\n", wp.ID, wp.Subject)
	result += fmt.Sprintf("- **ID:** %d\n", wp.ID)
	if wp.StartDate != "" {
		result += fmt.Sprintf("- **Start Date:** %s\n", wp.StartDate)
	}
	if wp.DueDate != "" {
		result += fmt.Sprintf("- **Due Date:** %s\n", wp.DueDate)
	}
	if wp.Description != "" {
		result += fmt.Sprintf("\n## Description\n%s\n", wp.Description)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) createWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.CreateWorkPackageOptions{
		Subject:     args.Subject,
		Description: args.Description,
		StartDate:   args.StartDate,
		DueDate:     args.DueDate,
		Links:       &openproject.CreateWorkPackageLinks{},
	}
	if args.TypeID > 0 {
		opts.Links.Type = &openproject.WorkPackageLink{Href: fmt.Sprintf("/api/v3/types/%d", args.TypeID)}
	}
	if args.StatusID > 0 {
		opts.Links.Status = &openproject.WorkPackageLink{Href: fmt.Sprintf("/api/v3/statuses/%d", args.StatusID)}
	}
	if args.PriorityID > 0 {
		opts.Links.Priority = &openproject.WorkPackageLink{Href: fmt.Sprintf("/api/v3/priorities/%d", args.PriorityID)}
	}
	if args.AssigneeID > 0 {
		opts.Links.Assignee = &openproject.WorkPackageLink{Href: fmt.Sprintf("/api/v3/users/%d", args.AssigneeID)}
	}

	wp, err := r.client.CreateWorkPackage(ctx, args.ProjectID, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create work package: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Work package #%d created successfully!\n\nSubject: %s", wp.ID, wp.Subject)}}}, nil
}

func (r *Registry) updateWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.UpdateWorkPackageOptions{
		Subject:     args.Subject,
		Description: args.Description,
		StartDate:   args.StartDate,
		DueDate:     args.DueDate,
		Links:       &openproject.UpdateWorkPackageLinks{},
	}
	if args.StatusID > 0 {
		opts.Links.Status = &openproject.WorkPackageLink{Href: fmt.Sprintf("/api/v3/statuses/%d", args.StatusID)}
	}
	if args.PriorityID > 0 {
		opts.Links.Priority = &openproject.WorkPackageLink{Href: fmt.Sprintf("/api/v3/priorities/%d", args.PriorityID)}
	}
	if args.AssigneeID > 0 {
		opts.Links.Assignee = &openproject.WorkPackageLink{Href: fmt.Sprintf("/api/v3/users/%d", args.AssigneeID)}
	}

	wp, err := r.client.UpdateWorkPackage(ctx, args.ID, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update work package: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Work package #%d updated successfully!\n\nSubject: %s", wp.ID, wp.Subject)}}}, nil
}

func (r *Registry) deleteWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	if err := r.client.DeleteWorkPackage(ctx, args.ID); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete work package: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Work package #%d deleted successfully!", args.ID)}}}, nil
}

func (r *Registry) listTypes(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListTypesArgs
	parseArgs(req.Params.Arguments, &args)

	var list *openproject.TypeList
	var err error
	if args.ProjectID > 0 {
		list, err = r.client.ListProjectTypes(ctx, args.ProjectID)
	} else {
		list, err = r.client.ListTypes(ctx)
	}
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list types: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d work package types:\n\n", list.Total)
	for _, t := range list.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", t.Name, t.ID)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) listStatuses(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	list, err := r.client.ListStatuses(ctx)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list statuses: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d statuses:\n\n", list.Total)
	for _, s := range list.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", s.Name, s.ID)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) listPriorities(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	list, err := r.client.ListPriorities(ctx)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list priorities: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d priorities:\n\n", list.Total)
	for _, p := range list.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", p.Name, p.ID)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}
