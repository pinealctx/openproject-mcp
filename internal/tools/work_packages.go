package tools

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
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
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListWorkPackagesParams{}
	if args.Offset > 0 {
		params.Offset = intPtr(args.Offset)
	}
	if args.PageSize > 0 {
		params.PageSize = intPtr(args.PageSize)
	}
	if args.SortBy != "" {
		params.SortBy = strPtr(normalizeSortBy(args.SortBy))
	}
	if args.Filters != "" {
		params.Filters = strPtr(args.Filters)
	}

	var list external.WorkPackagesModel
	var resp *http.Response
	var err error

	if args.ProjectID > 0 {
		pParams := &external.GetProjectWorkPackageCollectionParams{}
		if args.Offset > 0 {
			pParams.Offset = intPtr(args.Offset)
		}
		if args.PageSize > 0 {
			pParams.PageSize = intPtr(args.PageSize)
		}
		if args.SortBy != "" {
			pParams.SortBy = strPtr(normalizeSortBy(args.SortBy))
		}
		if args.Filters != "" {
			pParams.Filters = strPtr(args.Filters)
		}
		resp, err = r.client.APIClient().GetProjectWorkPackageCollection(ctx, args.ProjectID, pParams)
	} else {
		resp, err = r.client.APIClient().ListWorkPackages(ctx, params)
	}
	if err != nil {
		return errorResult("Failed to list work packages: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list work packages: %v", err), nil
	}

	result := fmt.Sprintf("Found %d work packages:\n\n", list.Total)
	for _, wp := range list.UnderscoreEmbedded.Elements {
		status, assignee := "", ""
		if wp.UnderscoreLinks.Status.Title != nil {
			status = *wp.UnderscoreLinks.Status.Title
		}
		if wp.UnderscoreLinks.Assignee != nil && wp.UnderscoreLinks.Assignee.Title != nil {
			assignee = *wp.UnderscoreLinks.Assignee.Title
		}
		result += fmt.Sprintf("- **#%d %s** — Status: %s, Assignee: %s\n",
			derefInt(wp.Id), wp.Subject,
			firstNonEmpty(status, "Unknown"),
			firstNonEmpty(assignee, "Unassigned"))
	}
	return textResult(result), nil
}

func (r *Registry) getWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().ViewWorkPackage(ctx, args.ID, nil)
	if err != nil {
		return errorResult("Failed to get work package: %v", err), nil
	}
	var wp external.WorkPackageModel
	if err := openproject.ReadResponse(resp, &wp); err != nil {
		return errorResult("Failed to get work package: %v", err), nil
	}

	result := fmt.Sprintf("# %s\n\n", wp.Subject)
	result += fmt.Sprintf("- **ID:** %d\n", derefInt(wp.Id))
	result += fmt.Sprintf("- **Project:** %s\n", derefStr(wp.UnderscoreLinks.Project.Title))
	result += fmt.Sprintf("- **Type:** %s\n", derefStr(wp.UnderscoreLinks.Type.Title))
	result += fmt.Sprintf("- **Status:** %s\n", derefStr(wp.UnderscoreLinks.Status.Title))
	result += fmt.Sprintf("- **Priority:** %s\n", derefStr(wp.UnderscoreLinks.Priority.Title))
	if wp.UnderscoreLinks.Assignee != nil {
		result += fmt.Sprintf("- **Assignee:** %s\n", derefStr(wp.UnderscoreLinks.Assignee.Title))
	} else {
		result += "- **Assignee:** Unassigned\n"
	}
	if wp.PercentageDone != nil {
		result += fmt.Sprintf("- **Progress:** %d%%\n", *wp.PercentageDone)
	}
	if wp.EstimatedTime != nil {
		result += fmt.Sprintf("- **Estimated Time:** %s\n", *wp.EstimatedTime)
	}
	if wp.StartDate != nil {
		result += fmt.Sprintf("- **Start Date:** %s\n", wp.StartDate.String())
	}
	if wp.DueDate != nil {
		result += fmt.Sprintf("- **Due Date:** %s\n", wp.DueDate.String())
	}
	if wp.LockVersion != nil {
		result += fmt.Sprintf("- **Lock Version:** %d\n", *wp.LockVersion)
	}
	if wp.Description != nil && wp.Description.Raw != nil && *wp.Description.Raw != "" {
		result += fmt.Sprintf("\n## Description\n%s\n", *wp.Description.Raw)
	}
	return textResult(result), nil
}

func (r *Registry) createWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.WorkPackageModel{
		Subject: args.Subject,
	}
	if args.Description != "" {
		fmt := external.FormattableFormat("markdown")
		body.Description = &external.Formattable{Format: &fmt, Raw: strPtr(args.Description)}
	}
	if args.StartDate != "" {
		body.StartDate = parseDatePtr(args.StartDate)
	}
	if args.DueDate != "" {
		body.DueDate = parseDatePtr(args.DueDate)
	}
	if args.EstimatedTime != "" {
		body.EstimatedTime = strPtr(args.EstimatedTime)
	}
	if args.TypeID > 0 {
		body.UnderscoreLinks.Type = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/types/%d", args.TypeID))}
	}
	if args.StatusID > 0 {
		body.UnderscoreLinks.Status = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/statuses/%d", args.StatusID))}
	}
	if args.PriorityID > 0 {
		body.UnderscoreLinks.Priority = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/priorities/%d", args.PriorityID))}
	}
	if args.AssigneeID > 0 {
		body.UnderscoreLinks.Assignee = &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/users/%d", args.AssigneeID))}
	}

	params := &external.CreateProjectWorkPackageParams{}
	resp, err := r.client.APIClient().CreateProjectWorkPackage(ctx, args.ProjectID, params, body)
	if err != nil {
		return errorResult("Failed to create work package: %v", err), nil
	}
	var wp external.WorkPackageModel
	if err := openproject.ReadResponse(resp, &wp); err != nil {
		return errorResult("Failed to create work package: %v", err), nil
	}
	return textResult(fmt.Sprintf("Work package #%d created successfully!\n\nSubject: %s", derefInt(wp.Id), wp.Subject)), nil
}

func (r *Registry) updateWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// First fetch the work package to get LockVersion (required for patch)
	resp, err := r.client.APIClient().ViewWorkPackage(ctx, args.ID, nil)
	if err != nil {
		return errorResult("Failed to fetch work package for lock version: %v", err), nil
	}
	var current external.WorkPackageModel
	if err := openproject.ReadResponse(resp, &current); err != nil {
		return errorResult("Failed to fetch work package for lock version: %v", err), nil
	}

	lockVersion := 0
	if current.LockVersion != nil {
		lockVersion = *current.LockVersion
	}

	body := external.WorkPackagePatchModel{
		LockVersion: lockVersion,
	}
	if args.Subject != "" {
		body.Subject = strPtr(args.Subject)
	}
	if args.Description != "" {
		fmt := external.FormattableFormat("markdown")
		body.Description = &external.Formattable{Format: &fmt, Raw: strPtr(args.Description)}
	}
	if args.StartDate != "" {
		body.StartDate = parseDatePtr(args.StartDate)
	}
	if args.DueDate != "" {
		body.DueDate = parseDatePtr(args.DueDate)
	}
	if args.EstimatedTime != "" {
		body.EstimatedTime = strPtr(args.EstimatedTime)
	}
	if args.PercentageDone != nil {
		// PercentageDone is not in WorkPackagePatchModel; it must be set via custom fields or omitted
		// The external client uses WorkPackagePatchModel which doesn't have percentageDone
	}

	body.UnderscoreLinks = &struct {
		Assignee    *external.Link `json:"assignee,omitempty"`
		Category    *external.Link `json:"category,omitempty"`
		Parent      *external.Link `json:"parent,omitempty"`
		Priority    *external.Link `json:"priority,omitempty"`
		Project     *external.Link `json:"project,omitempty"`
		Responsible *external.Link `json:"responsible,omitempty"`
		Status      *external.Link `json:"status,omitempty"`
		Type        *external.Link `json:"type,omitempty"`
		Version     *external.Link `json:"version,omitempty"`
	}{}
	if args.StatusID > 0 {
		body.UnderscoreLinks.Status = &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/statuses/%d", args.StatusID))}
	}
	if args.PriorityID > 0 {
		body.UnderscoreLinks.Priority = &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/priorities/%d", args.PriorityID))}
	}
	if args.AssigneeID > 0 {
		body.UnderscoreLinks.Assignee = &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/users/%d", args.AssigneeID))}
	}

	resp, err = r.client.APIClient().UpdateWorkPackage(ctx, args.ID, nil, body)
	if err != nil {
		return errorResult("Failed to update work package: %v", err), nil
	}
	var wp external.WorkPackageModel
	if err := openproject.ReadResponse(resp, &wp); err != nil {
		return errorResult("Failed to update work package: %v", err), nil
	}
	return textResult(fmt.Sprintf("Work package #%d updated successfully!\n\nSubject: %s", derefInt(wp.Id), wp.Subject)), nil
}

func (r *Registry) deleteWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().DeleteWorkPackage(ctx, args.ID)
	if err != nil {
		return errorResult("Failed to delete work package: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to delete work package: %v", err), nil
	}
	return textResult(fmt.Sprintf("Work package #%d deleted successfully!", args.ID)), nil
}

func (r *Registry) listTypes(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListTypesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	var resp *http.Response
	var err error
	if args.ProjectID > 0 {
		resp, err = r.client.APIClient().ListTypesAvailableInAProject(ctx, args.ProjectID)
	} else {
		resp, err = r.client.APIClient().ListAllTypes(ctx)
	}
	if err != nil {
		return errorResult("Failed to list types: %v", err), nil
	}

	var list external.TypesByWorkspaceModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list types: %v", err), nil
	}

	result := fmt.Sprintf("Found %d work package types:\n\n", list.Total)
	if list.UnderscoreEmbedded.Elements != nil {
		for _, t := range *list.UnderscoreEmbedded.Elements {
			result += fmt.Sprintf("- **%s** (ID: %d)\n", derefStr(t.Name), derefInt(t.Id))
		}
	}
	return textResult(result), nil
}

func (r *Registry) listStatuses(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := r.client.APIClient().ListStatuses(ctx)
	if err != nil {
		return errorResult("Failed to list statuses: %v", err), nil
	}

	var list external.StatusCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list statuses: %v", err), nil
	}

	result := fmt.Sprintf("Found %d statuses:\n\n", list.Total)
	for _, s := range list.UnderscoreEmbedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", derefStr(s.Name), derefInt(s.Id))
	}
	return textResult(result), nil
}

func (r *Registry) listPriorities(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := r.client.APIClient().ListAllPriorities(ctx)
	if err != nil {
		return errorResult("Failed to list priorities: %v", err), nil
	}

	var list external.PriorityCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list priorities: %v", err), nil
	}

	result := fmt.Sprintf("Found %d priorities:\n\n", list.Total)
	for _, p := range list.UnderscoreEmbedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", derefStr(p.Name), derefInt(p.Id))
	}
	return textResult(result), nil
}

func (r *Registry) listAvailableAssignees(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListAvailableAssigneesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().WorkPackageAvailableAssignees(ctx, args.WorkPackageID)
	if err != nil {
		return errorResult("Failed to list available assignees: %v", err), nil
	}

	var list external.AvailableAssigneesModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list available assignees: %v", err), nil
	}

	result := fmt.Sprintf("Found %d available assignees for work package #%d:\n\n", list.Total, args.WorkPackageID)
	if list.UnderscoreEmbedded.Elements != nil {
		for _, u := range *list.UnderscoreEmbedded.Elements {
			result += fmt.Sprintf("- **%s** (ID: %d, Email: %s)\n", u.Name, u.Id, derefStr(u.Email))
		}
	}
	return textResult(result), nil
}
