package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

type ListWorkPackageWatchersArgs struct {
	WorkPackageID int `json:"workPackageId"`
}

type AddWorkPackageWatcherArgs struct {
	WorkPackageID int `json:"workPackageId"`
	UserID        int `json:"userId"`
}

type RemoveWorkPackageWatcherArgs struct {
	WorkPackageID int `json:"workPackageId"`
	UserID        int `json:"userId"`
}

func (r *Registry) registerWatcherTools(server *mcp.Server) {
	addTool(server, "list_work_package_watchers",
		"List watchers for a work package",
		newSchema(schemaProps{
			"workPackageId": schemaInt("Work package ID"),
		}, "workPackageId"),
		r.listWorkPackageWatchers)

	addTool(server, "add_work_package_watcher",
		"Add a watcher to a work package",
		newSchema(schemaProps{
			"workPackageId": schemaInt("Work package ID"),
			"userId":        schemaInt("User ID to add as watcher"),
		}, "workPackageId", "userId"),
		r.addWorkPackageWatcher)

	addTool(server, "remove_work_package_watcher",
		"Remove a watcher from a work package",
		newSchema(schemaProps{
			"workPackageId": schemaInt("Work package ID"),
			"userId":        schemaInt("User ID to remove as watcher"),
		}, "workPackageId", "userId"),
		r.removeWorkPackageWatcher)
}

func (r *Registry) listWorkPackageWatchers(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackageWatchersArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().ListWatchers(ctx, args.WorkPackageID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list watchers: %v", err), nil
	}
	var watchers external.WatchersModel
	if err := openproject.ReadResponse(resp, &watchers); err != nil {
		return errorResult("Failed to list watchers: %v", err), nil
	}

	if watchers.Count == 0 {
		return textResult(fmt.Sprintf("No watchers found for work package #%d.", args.WorkPackageID)), nil
	}

	result := fmt.Sprintf("Found %d watchers for work package #%d:\n\n", watchers.Count, args.WorkPackageID)
	if watchers.UnderscoreEmbedded.Elements != nil {
		for _, u := range *watchers.UnderscoreEmbedded.Elements {
			result += fmt.Sprintf("- **%s** (ID: %d)\n", u.Name, u.Id)
		}
	}
	return textResult(result), nil
}

func (r *Registry) addWorkPackageWatcher(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args AddWorkPackageWatcherArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	href := fmt.Sprintf("/api/v3/users/%d", args.UserID)
	body := external.AddWatcherJSONRequestBody{
		User: &struct {
			Href *string `json:"href,omitempty"`
		}{
			Href: &href,
		},
	}

	resp, err := r.client.APIClient().AddWatcher(ctx, args.WorkPackageID, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to add watcher: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to add watcher: %v", err), nil
	}

	return textResult(fmt.Sprintf("User #%d added as watcher to work package #%d.", args.UserID, args.WorkPackageID)), nil
}

func (r *Registry) removeWorkPackageWatcher(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args RemoveWorkPackageWatcherArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().RemoveWatcher(ctx, args.WorkPackageID, args.UserID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to remove watcher: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to remove watcher: %v", err), nil
	}

	return textResult(fmt.Sprintf("User #%d removed as watcher from work package #%d.", args.UserID, args.WorkPackageID)), nil
}
