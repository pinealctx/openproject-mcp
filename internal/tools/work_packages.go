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
	AssigneeID    *int   `json:"assigneeId,omitempty"`
	AccountableID *int   `json:"accountableId,omitempty"`
	StartDate     string `json:"startDate,omitempty"`
	DueDate       string `json:"dueDate,omitempty"`
	EstimatedTime string `json:"estimatedTime,omitempty"`
}

type UpdateWorkPackageArgs struct {
	ID               int    `json:"id"`
	Subject          string `json:"subject,omitempty"`
	Description      string `json:"description,omitempty"`
	StatusID         int    `json:"statusId,omitempty"`
	PriorityID       int    `json:"priorityId,omitempty"`
	AssigneeID       *int   `json:"assigneeId,omitempty"`
	AccountableID    *int   `json:"accountableId,omitempty"`
	ClearAssignee    bool   `json:"clearAssignee,omitempty"`
	ClearAccountable bool   `json:"clearAccountable,omitempty"`
	StartDate        string `json:"startDate,omitempty"`
	DueDate          string `json:"dueDate,omitempty"`
	EstimatedTime    string `json:"estimatedTime,omitempty"`
	PercentageDone   *int   `json:"percentageDone,omitempty"`
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
			"assigneeId":    schemaInt("Assignee user ID (the person currently working on the work package)"),
			"accountableId": schemaInt("Accountable user ID (the person responsible for delivery)"),
			"startDate":     schemaStr("Start date (YYYY-MM-DD)"),
			"dueDate":       schemaStr("Due date (YYYY-MM-DD)"),
			"estimatedTime": schemaStr(`Estimated time in ISO 8601 duration, e.g. "PT4H" for 4 hours`),
		}, "projectId", "subject"),
		r.createWorkPackage)

	addTool(server, "update_work_package", "Update an existing work package",
		newSchema(schemaProps{
			"id":               schemaInt("Work package ID"),
			"subject":          schemaStr("New title / subject"),
			"description":      schemaStr("New description"),
			"statusId":         schemaInt("New status ID"),
			"priorityId":       schemaInt("New priority ID"),
			"assigneeId":       schemaInt("New Assignee user ID (current worker)"),
			"accountableId":    schemaInt("New Accountable user ID (delivery owner)"),
			"clearAssignee":    schemaBool("Remove the current Assignee; cannot be combined with assigneeId"),
			"clearAccountable": schemaBool("Remove the current Accountable; cannot be combined with accountableId"),
			"startDate":        schemaStr("New start date (YYYY-MM-DD)"),
			"dueDate":          schemaStr("New due date (YYYY-MM-DD)"),
			"estimatedTime":    schemaStr(`New estimated time, e.g. "PT8H"`),
			"percentageDone":   schemaInt("Completion percentage (0-100)"),
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

	list, err := r.client.ListWorkPackages(ctx, openproject.WorkPackageListInput{
		ProjectID: args.ProjectID,
		Offset:    args.Offset,
		PageSize:  args.PageSize,
		SortBy:    normalizeSortBy(args.SortBy),
		Filters:   args.Filters,
	})
	if err != nil {
		return errorResult("Failed to list work packages: %v", err), nil
	}

	result := fmt.Sprintf("Found %d work packages:\n\n", list.Total)
	for _, wp := range list.UnderscoreEmbedded.Elements {
		status, assignee, accountable := "", "", ""
		if wp.UnderscoreLinks.Status.Title != nil {
			status = *wp.UnderscoreLinks.Status.Title
		}
		if wp.UnderscoreLinks.Assignee != nil && wp.UnderscoreLinks.Assignee.Title != nil {
			assignee = *wp.UnderscoreLinks.Assignee.Title
		}
		if wp.UnderscoreLinks.Responsible != nil && wp.UnderscoreLinks.Responsible.Title != nil {
			accountable = *wp.UnderscoreLinks.Responsible.Title
		}
		result += fmt.Sprintf("- **#%d %s** — Status: %s, Assignee: %s, Accountable: %s\n",
			derefInt(wp.Id), wp.Subject,
			firstNonEmpty(status, "Unknown"),
			firstNonEmpty(assignee, "Unassigned"),
			firstNonEmpty(accountable, "Unassigned"))
	}
	return textResult(result), nil
}

func (r *Registry) getWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	wp, err := r.client.GetWorkPackage(ctx, args.ID)
	if err != nil {
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
	if wp.UnderscoreLinks.Responsible != nil {
		result += fmt.Sprintf("- **Accountable:** %s\n", derefStr(wp.UnderscoreLinks.Responsible.Title))
	} else {
		result += "- **Accountable:** Unassigned\n"
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

	wp, err := r.client.CreateWorkPackage(ctx, openproject.WorkPackageCreateInput{
		ProjectID:     args.ProjectID,
		Subject:       args.Subject,
		Description:   args.Description,
		TypeID:        openproject.IntID(args.TypeID),
		StatusID:      openproject.IntID(args.StatusID),
		PriorityID:    openproject.IntID(args.PriorityID),
		AssigneeID:    args.AssigneeID,
		AccountableID: args.AccountableID,
		StartDate:     args.StartDate,
		DueDate:       args.DueDate,
		EstimatedTime: args.EstimatedTime,
	})
	if err != nil {
		return errorResult("Failed to create work package: %v", err), nil
	}
	return textResult(fmt.Sprintf("Work package #%d created successfully!\n\nSubject: %s", derefInt(wp.Id), wp.Subject)), nil
}

func (r *Registry) updateWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	wp, err := r.client.UpdateWorkPackage(ctx, openproject.WorkPackageUpdateInput{
		ID:               args.ID,
		Subject:          args.Subject,
		Description:      args.Description,
		StatusID:         openproject.IntID(args.StatusID),
		PriorityID:       openproject.IntID(args.PriorityID),
		AssigneeID:       args.AssigneeID,
		AccountableID:    args.AccountableID,
		ClearAssignee:    args.ClearAssignee,
		ClearAccountable: args.ClearAccountable,
		StartDate:        args.StartDate,
		DueDate:          args.DueDate,
		EstimatedTime:    args.EstimatedTime,
		PercentageDone:   args.PercentageDone,
	})
	if err != nil {
		return errorResult("Failed to update work package: %v", err), nil
	}
	return textResult(fmt.Sprintf("Work package #%d updated successfully!\n\nSubject: %s", derefInt(wp.Id), wp.Subject)), nil
}

func (r *Registry) deleteWorkPackage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteWorkPackageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	if err := r.client.DeleteWorkPackage(ctx, args.ID); err != nil {
		return errorResult("Failed to delete work package: %v", err), nil
	}
	return textResult(fmt.Sprintf("Work package #%d deleted successfully!", args.ID)), nil
}

func (r *Registry) listTypes(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListTypesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	list, err := r.client.ListWorkPackageTypes(ctx, args.ProjectID)
	if err != nil {
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
	list, err := r.client.ListWorkPackageStatuses(ctx)
	if err != nil {
		return errorResult("Failed to list statuses: %v", err), nil
	}

	result := fmt.Sprintf("Found %d statuses:\n\n", list.Total)
	for _, s := range list.UnderscoreEmbedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", derefStr(s.Name), derefInt(s.Id))
	}
	return textResult(result), nil
}

func (r *Registry) listPriorities(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	list, err := r.client.ListWorkPackagePriorities(ctx)
	if err != nil {
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

	list, err := r.client.ListWorkPackageAvailableAssignees(ctx, args.WorkPackageID)
	if err != nil {
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
