package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

type ListTimeEntriesArgs struct {
	ProjectID     int    `json:"projectId,omitempty"`
	WorkPackageID int    `json:"workPackageId,omitempty"`
	UserID        int    `json:"userId,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	PageSize      int    `json:"pageSize,omitempty"`
	SortBy        string `json:"sortBy,omitempty"`
}

type CreateTimeEntryArgs struct {
	Hours         string `json:"hours"`
	Comment       string `json:"comment,omitempty"`
	SpentOn       string `json:"spentOn,omitempty"`
	ProjectID     int    `json:"projectId,omitempty"`
	WorkPackageID int    `json:"workPackageId,omitempty"`
	ActivityID    int    `json:"activityId,omitempty"`
	UserID        int    `json:"userId,omitempty"`
}

type UpdateTimeEntryArgs struct {
	ID         int    `json:"id"`
	Hours      string `json:"hours,omitempty"`
	Comment    string `json:"comment,omitempty"`
	SpentOn    string `json:"spentOn,omitempty"`
	ActivityID int    `json:"activityId,omitempty"`
}

type DeleteTimeEntryArgs struct{ ID int }
type ListTimeEntryActivitiesArgs struct{}

// registerTimeEntryTools registers time entry-related tools.
func (r *Registry) registerTimeEntryTools(server *mcp.Server) {
	addTool(server, "list_time_entries", "List time entries with optional filters",
		newSchema(schemaProps{
			"projectId":     schemaInt("Filter by project ID"),
			"workPackageId": schemaInt("Filter by work package ID"),
			"userId":        schemaInt("Filter by user ID"),
			"offset":        schemaInt("Pagination offset"),
			"pageSize":      schemaInt("Items per page"),
			"sortBy":        schemaStr(`Sort criteria, e.g. "spentOn:desc"`),
		}),
		r.listTimeEntries)

	addTool(server, "create_time_entry", "Create a new time entry to log work",
		newSchema(schemaProps{
			"hours":         schemaStr(`Hours spent, e.g. "8.5" or "PT8H30M"`),
			"projectId":     schemaInt("Project ID"),
			"workPackageId": schemaInt("Work package ID (optional)"),
			"activityId":    schemaInt("Activity type ID"),
			"comment":       schemaStr("Optional comment / description"),
			"spentOn":       schemaStr("Date (YYYY-MM-DD, defaults to today)"),
			"userId":        schemaInt("User ID (defaults to current user)"),
		}, "hours"),
		r.createTimeEntry)

	addTool(server, "update_time_entry", "Update an existing time entry",
		newSchema(schemaProps{
			"id":         schemaInt("Time entry ID"),
			"hours":      schemaStr("New hours value"),
			"comment":    schemaStr("New comment"),
			"spentOn":    schemaStr("New date (YYYY-MM-DD)"),
			"activityId": schemaInt("New activity type ID"),
		}, "id"),
		r.updateTimeEntry)

	addTool(server, "delete_time_entry", "Delete a time entry",
		newSchema(schemaProps{"id": schemaInt("Time entry ID")}, "id"),
		r.deleteTimeEntry)

	addTool(server, "list_time_entry_activities", "List available time entry activity types",
		noSchema, r.listTimeEntryActivities)
}

func (r *Registry) listTimeEntries(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListTimeEntriesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListTimeEntriesParams{}
	if args.Offset > 0 {
		params.Offset = intPtr(args.Offset)
	}
	if args.PageSize > 0 {
		params.PageSize = intPtr(args.PageSize)
	}
	if args.SortBy != "" {
		params.SortBy = strPtr(normalizeSortBy(args.SortBy))
	}

	var filters []string
	if args.ProjectID > 0 {
		filters = append(filters, fmt.Sprintf(`{"project_id":{"operator":"=","values":["%d"]}}`, args.ProjectID))
	}
	if args.WorkPackageID > 0 {
		filters = append(filters,
			`{"entity_type":{"operator":"=","values":["WorkPackage"]}}`,
			fmt.Sprintf(`{"entity_id":{"operator":"=","values":["%d"]}}`, args.WorkPackageID),
		)
	}
	if args.UserID > 0 {
		filters = append(filters, fmt.Sprintf(`{"user_id":{"operator":"=","values":["%d"]}}`, args.UserID))
	}
	if len(filters) > 0 {
		params.Filters = strPtr("[" + joinStrings(filters, ",") + "]")
	}

	resp, err := r.client.APIClient().ListTimeEntries(ctx, params)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list time entries: %v", err), nil
	}

	var list external.TimeEntryCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list time entries: %v", err), nil
	}

	result := fmt.Sprintf("Found %d time entries:\n\n", list.Total)
	for _, e := range list.UnderscoreEmbedded.Elements {
		spentOn := ""
		if e.SpentOn != nil {
			spentOn = e.SpentOn.String()
		}
		result += fmt.Sprintf("- #%d: %s on %s\n", derefInt(e.Id), derefStr(e.Hours), spentOn)
	}
	return textResult(result), nil
}

func (r *Registry) createTimeEntry(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateTimeEntryArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.TimeEntryModel{}
	if args.Hours != "" {
		body.Hours = strPtr(args.Hours)
	}
	if args.Comment != "" {
		fmt := external.FormattableFormat("markdown")
		body.Comment = &external.Formattable{Format: &fmt, Raw: strPtr(args.Comment)}
	}
	if args.SpentOn != "" {
		body.SpentOn = parseDatePtr(args.SpentOn)
	}

	// Set links for project, activity, user, entity
	body.UnderscoreLinks.Project = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/projects/%d", args.ProjectID))}
	if args.ActivityID > 0 {
		body.UnderscoreLinks.Activity = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/time_entries/activities/%d", args.ActivityID))}
	}
	if args.UserID > 0 {
		body.UnderscoreLinks.User = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/users/%d", args.UserID))}
	}
	if args.WorkPackageID > 0 {
		body.UnderscoreLinks.Entity = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/work_packages/%d", args.WorkPackageID))}
	} else {
		body.UnderscoreLinks.Entity = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/projects/%d", args.ProjectID))}
	}

	resp, err := r.client.APIClient().CreateTimeEntry(ctx, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to create time entry: %v", err), nil
	}
	var entry external.TimeEntryModel
	if err := openproject.ReadResponse(resp, &entry); err != nil {
		return errorResult("Failed to create time entry: %v", err), nil
	}
	spentOn := ""
	if entry.SpentOn != nil {
		spentOn = entry.SpentOn.String()
	}
	return textResult(fmt.Sprintf("Time entry #%d created: %s on %s", derefInt(entry.Id), derefStr(entry.Hours), spentOn)), nil
}

func (r *Registry) updateTimeEntry(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateTimeEntryArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// Use raw Patch via client for time entry update since generated UpdateTimeEntry has no body
	body := map[string]interface{}{}
	if args.Hours != "" {
		body["hours"] = args.Hours
	}
	if args.Comment != "" {
		body["comment"] = map[string]interface{}{
			"format": "markdown",
			"raw":    args.Comment,
		}
	}
	if args.SpentOn != "" {
		body["spentOn"] = args.SpentOn
	}
	if args.ActivityID > 0 {
		body["_links"] = map[string]interface{}{
			"activity": map[string]interface{}{
				"href": fmt.Sprintf("/api/v3/time_entries/activities/%d", args.ActivityID),
			},
		}
	}

	var result external.TimeEntryModel
	if err := r.client.Patch(ctx, fmt.Sprintf("/time_entries/%d", args.ID), body, &result); err != nil {
		return errorResult("Failed to update time entry: %v", err), nil
	}
	return textResult(fmt.Sprintf("Time entry #%d updated successfully!", derefInt(result.Id))), nil
}

func (r *Registry) deleteTimeEntry(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteTimeEntryArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().DeleteTimeEntry(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to delete time entry: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to delete time entry: %v", err), nil
	}
	return textResult(fmt.Sprintf("Time entry #%d deleted successfully!", args.ID)), nil
}

func (r *Registry) listTimeEntryActivities(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Time entry activities are available via raw GET since there's no direct generated method
	var result struct {
		Total     int `json:"total"`
		Embedded  struct {
			Elements []external.TimeEntryActivityModel `json:"elements"`
		} `json:"_embedded"`
	}
	if err := r.client.Get(ctx, "/time_entries/activities", &result); err != nil {
		return errorResult("Failed to list activities: %v", err), nil
	}

	out := fmt.Sprintf("Found %d activities:\n\n", result.Total)
	for _, a := range result.Embedded.Elements {
		out += fmt.Sprintf("- **%s** (ID: %d)\n", a.Name, a.Id)
	}
	return textResult(out), nil
}

// joinStrings joins string slices with a separator.
func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
