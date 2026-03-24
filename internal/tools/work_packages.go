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
	SortBy    string `json:"sortBy,omitempty"`
	Filters   string `json:"filters,omitempty"`
}

type GetWorkPackageArgs struct {
	ID int `json:"id"`
}

type CreateWorkPackageArgs struct {
	ProjectID     int    `json:"projectId"`
	Subject       string `json:"subject"`
	Description   string `json:"description,omitempty"`
	TypeID        int    `json:"typeId,omitempty"`
	StatusID      int    `json:"statusId,omitempty"`
	PriorityID    int    `json:"priorityId,omitempty"`
	AssigneeID    int    `json:"assigneeId,omitempty"`
	StartDate     string `json:"startDate,omitempty"`
	DueDate       string `json:"dueDate,omitempty"`
	EstimatedTime string `json:"estimatedTime,omitempty"`
}

type UpdateWorkPackageArgs struct {
	ID             int    `json:"id"`
	Subject        string `json:"subject,omitempty"`
	Description    string `json:"description,omitempty"`
	StatusID       int    `json:"statusId,omitempty"`
	PriorityID     int    `json:"priorityId,omitempty"`
	AssigneeID     int    `json:"assigneeId,omitempty"`
	StartDate      string `json:"startDate,omitempty"`
	DueDate        string `json:"dueDate,omitempty"`
	EstimatedTime  string `json:"estimatedTime,omitempty"`
	PercentageDone *int   `json:"percentageDone,omitempty"`
}

type DeleteWorkPackageArgs struct {
	ID int `json:"id"`
}

type ListTypesArgs struct {
	ProjectID int `json:"projectId,omitempty"`
}

type ListStatusesArgs struct{}
type ListPrioritiesArgs struct{}

type ListAvailableAssigneesArgs struct {
	WorkPackageID int `json:"workPackageId"`
}

// registerWorkPackageTools registers work package-related tools.
func (r *Registry) registerWorkPackageTools(server *mcp.Server) {
	addTool(server, "list_work_packages", "List work packages, optionally filtered by project",
		newSchema(schemaProps{
			"projectId": schemaInt("Filter by project ID"),
			"offset":    schemaInt("Pagination offset (default: 0)"),
			"pageSize":  schemaInt("Number of items per page (default: 20)"),
			"sortBy":    schemaStr(`Sort criteria, e.g. "updatedAt:desc"`),
			"filters":   schemaStr("Raw JSON filter string"),
		}),
		r.listWorkPackages)

	addTool(server, "get_work_package", "Get details of a specific work package by ID",
		newSchema(schemaProps{
			"id": schemaInt("Work package ID"),
		}, "id"),
		r.getWorkPackage)

	addTool(server, "create_work_package", "Create a new work package in a project",
		newSchema(schemaProps{
			"projectId":     schemaInt("Project ID"),
			"subject":       schemaStr("Work package title / subject"),
			"description":   schemaStr("Work package description"),
			"typeId":        schemaInt("Type ID (task, bug, feature, etc.)"),
			"statusId":      schemaInt("Status ID"),
			"priorityId":    schemaInt("Priority ID"),
			"assigneeId":    schemaInt("Assignee user ID"),
			"startDate":     schemaStr("Start date (YYYY-MM-DD)"),
			"dueDate":       schemaStr("Due date (YYYY-MM-DD)"),
			"estimatedTime": schemaStr(`Estimated time in ISO 8601 duration, e.g. "PT4H" for 4 hours`),
		}, "projectId", "subject"),
		r.createWorkPackage)

	addTool(server, "update_work_package", "Update an existing work package",
		newSchema(schemaProps{
			"id":             schemaInt("Work package ID"),
			"subject":        schemaStr("New title / subject"),
			"description":    schemaStr("New description"),
			"statusId":       schemaInt("New status ID"),
			"priorityId":     schemaInt("New priority ID"),
			"assigneeId":     schemaInt("New assignee user ID"),
			"startDate":      schemaStr("New start date (YYYY-MM-DD)"),
			"dueDate":        schemaStr("New due date (YYYY-MM-DD)"),
			"estimatedTime":  schemaStr(`New estimated time, e.g. "PT8H"`),
			"percentageDone": schemaInt("Completion percentage (0-100)"),
		}, "id"),
		r.updateWorkPackage)

	addTool(server, "delete_work_package", "Delete a work package",
		newSchema(schemaProps{
			"id": schemaInt("Work package ID"),
		}, "id"),
		r.deleteWorkPackage)

	addTool(server, "list_types", "List work package types, optionally for a specific project",
		newSchema(schemaProps{
			"projectId": schemaInt("Project ID (omit to list all global types)"),
		}),
		r.listTypes)

	addTool(server, "list_statuses", "List all work package statuses",
		noSchema, r.listStatuses)

	addTool(server, "list_priorities", "List all work package priorities",
		noSchema, r.listPriorities)

	addTool(server, "list_available_assignees", "List users who can be assigned to a work package",
		newSchema(schemaProps{
			"workPackageId": schemaInt("Work package ID"),
		}, "workPackageId"),
		r.listAvailableAssignees)
}

func (r *Registry) listWorkPackages(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackagesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.ListWorkPackagesOptions{
		Offset:     args.Offset,
		PageSize:   args.PageSize,
		SortBy:     args.SortBy,
		RawFilters: args.Filters,
	}
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
		status, assignee := "", ""
		if wp.Links != nil {
			if wp.Links.Status != nil {
				status = wp.Links.Status.Title
			}
			if wp.Links.Assignee != nil {
				assignee = wp.Links.Assignee.Title
			}
		}
		result += fmt.Sprintf("- **#%d %s** — Status: %s, Assignee: %s\n",
			wp.ID, wp.Subject,
			firstNonEmpty(status, "Unknown"),
			firstNonEmpty(assignee, "Unassigned"))
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

	status, typeName, priority, assignee, project := "", "", "", "", ""
	if wp.Links != nil {
		if wp.Links.Status != nil {
			status = wp.Links.Status.Title
		}
		if wp.Links.Type != nil {
			typeName = wp.Links.Type.Title
		}
		if wp.Links.Priority != nil {
			priority = wp.Links.Priority.Title
		}
		if wp.Links.Assignee != nil {
			assignee = wp.Links.Assignee.Title
		}
		if wp.Links.Project != nil {
			project = wp.Links.Project.Title
		}
	}

	result := fmt.Sprintf("# #%d %s\n\n", wp.ID, wp.Subject)
	result += fmt.Sprintf("- **ID:** %d\n", wp.ID)
	result += fmt.Sprintf("- **Project:** %s\n", firstNonEmpty(project, "Unknown"))
	result += fmt.Sprintf("- **Type:** %s\n", firstNonEmpty(typeName, "Unknown"))
	result += fmt.Sprintf("- **Status:** %s\n", firstNonEmpty(status, "Unknown"))
	result += fmt.Sprintf("- **Priority:** %s\n", firstNonEmpty(priority, "Unknown"))
	result += fmt.Sprintf("- **Assignee:** %s\n", firstNonEmpty(assignee, "Unassigned"))
	result += fmt.Sprintf("- **Progress:** %d%%\n", wp.PercentageDone)
	if wp.EstimatedTime != "" {
		result += fmt.Sprintf("- **Estimated Time:** %s\n", wp.EstimatedTime)
	}
	if wp.StartDate != "" {
		result += fmt.Sprintf("- **Start Date:** %s\n", wp.StartDate)
	}
	if wp.DueDate != "" {
		result += fmt.Sprintf("- **Due Date:** %s\n", wp.DueDate)
	}
	result += fmt.Sprintf("- **Lock Version:** %d\n", wp.LockVersion)
	if wp.Description.Raw != "" {
		result += fmt.Sprintf("\n## Description\n%s\n", wp.Description.Raw)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) createWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.CreateWorkPackageOptions{
		Subject:       args.Subject,
		Description:   openproject.NewRichText(args.Description),
		StartDate:     args.StartDate,
		DueDate:       args.DueDate,
		EstimatedTime: args.EstimatedTime,
		Links:         &openproject.CreateWorkPackageLinks{},
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
		Subject:        args.Subject,
		Description:    openproject.NewRichText(args.Description),
		StartDate:      args.StartDate,
		DueDate:        args.DueDate,
		EstimatedTime:  args.EstimatedTime,
		PercentageDone: args.PercentageDone,
		Links:          &openproject.UpdateWorkPackageLinks{},
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
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

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

func (r *Registry) listAvailableAssignees(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListAvailableAssigneesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	list, err := r.client.ListAvailableAssignees(ctx, args.WorkPackageID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list available assignees: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d available assignees for work package #%d:\n\n", list.Total, args.WorkPackageID)
	for _, u := range list.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d, Email: %s)\n", u.Name, u.ID, u.Email)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}
