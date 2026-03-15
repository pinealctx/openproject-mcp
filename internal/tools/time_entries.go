package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type ListTimeEntriesArgs struct {
	ProjectID     int `json:"projectId,omitempty"`
	WorkPackageID int `json:"workPackageId,omitempty"`
	UserID        int `json:"userId,omitempty"`
	Offset        int `json:"offset,omitempty"`
	PageSize      int `json:"pageSize,omitempty"`
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
	server.AddTool(&mcp.Tool{Name: "list_time_entries", Description: "List time entries with optional filters"}, r.listTimeEntries)
	server.AddTool(&mcp.Tool{Name: "create_time_entry", Description: "Create a new time entry to log work"}, r.createTimeEntry)
	server.AddTool(&mcp.Tool{Name: "update_time_entry", Description: "Update an existing time entry"}, r.updateTimeEntry)
	server.AddTool(&mcp.Tool{Name: "delete_time_entry", Description: "Delete a time entry"}, r.deleteTimeEntry)
	server.AddTool(&mcp.Tool{Name: "list_time_entry_activities", Description: "List available time entry activity types"}, r.listTimeEntryActivities)
}

func (r *Registry) listTimeEntries(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListTimeEntriesArgs
	parseArgs(req.Params.Arguments, &args)

	opts := &openproject.ListTimeEntriesOptions{Offset: args.Offset, PageSize: args.PageSize}
	var filters []openproject.TimeEntryFilter
	if args.ProjectID > 0 {
		filters = append(filters, openproject.TimeEntryFilter{Name: "project", Values: []string{fmt.Sprintf("%d", args.ProjectID)}})
	}
	if args.WorkPackageID > 0 {
		filters = append(filters, openproject.TimeEntryFilter{Name: "workPackage", Values: []string{fmt.Sprintf("%d", args.WorkPackageID)}})
	}
	if args.UserID > 0 {
		filters = append(filters, openproject.TimeEntryFilter{Name: "user", Values: []string{fmt.Sprintf("%d", args.UserID)}})
	}
	opts.Filters = filters

	list, err := r.client.ListTimeEntries(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list time entries: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d time entries:\n\n", list.Total)
	for _, e := range list.Embedded.Elements {
		result += fmt.Sprintf("- #%d: %s on %s\n", e.ID, e.Hours, e.SpentOn)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) createTimeEntry(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateTimeEntryArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.CreateTimeEntryOptions{
		Hours: args.Hours, Comment: args.Comment, SpentOn: args.SpentOn,
		ProjectID: args.ProjectID, WorkPackage: args.WorkPackageID, ActivityID: args.ActivityID, UserID: args.UserID,
	}
	entry, err := r.client.CreateTimeEntry(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create time entry: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Time entry #%d created: %s on %s", entry.ID, entry.Hours, entry.SpentOn)}}}, nil
}

func (r *Registry) updateTimeEntry(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateTimeEntryArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.UpdateTimeEntryOptions{Hours: args.Hours, Comment: args.Comment, SpentOn: args.SpentOn, ActivityID: args.ActivityID}
	entry, err := r.client.UpdateTimeEntry(ctx, args.ID, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update time entry: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Time entry #%d updated successfully!", entry.ID)}}}, nil
}

func (r *Registry) deleteTimeEntry(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteTimeEntryArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	if err := r.client.DeleteTimeEntry(ctx, args.ID); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete time entry: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Time entry #%d deleted successfully!", args.ID)}}}, nil
}

func (r *Registry) listTimeEntryActivities(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	list, err := r.client.ListTimeEntryActivities(ctx)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list activities: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d activities:\n\n", list.Total)
	for _, a := range list.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", a.Name, a.ID)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}
