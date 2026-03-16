package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type ListNotificationsArgs struct {
	Unread   bool `json:"unread,omitempty"`
	PageSize int  `json:"pageSize,omitempty"`
	Offset   int  `json:"offset,omitempty"`
}

type MarkNotificationReadArgs struct {
	ID int `json:"id"`
}

func (r *Registry) registerNotificationTools(server *mcp.Server) {
	addTool(server, "list_notifications",
		"List notifications for the current user (mentions, status changes, watched items, etc.)",
		newSchema(schemaProps{
			"unread":   schemaBool("When true, return only unread notifications (default: false = all)"),
			"pageSize": schemaInt("Number of results per page"),
			"offset":   schemaInt("Pagination offset"),
		}),
		r.listNotifications)

	addTool(server, "mark_notification_read",
		"Mark a specific notification as read",
		newSchema(schemaProps{
			"id": schemaInt("Notification ID"),
		}, "id"),
		r.markNotificationRead)

	addTool(server, "mark_all_notifications_read",
		"Mark all notifications as read",
		noSchema,
		r.markAllNotificationsRead)
}

func (r *Registry) listNotifications(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListNotificationsArgs
	parseArgs(req.Params.Arguments, &args)

	opts := &openproject.ListNotificationsOptions{
		PageSize: args.PageSize,
		Offset:   args.Offset,
	}
	if args.Unread {
		opts.ReadIAN = "false"
	}

	list, err := r.client.ListNotifications(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list notifications: %v", err)}}}, nil
	}

	if list.Total == 0 {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "No notifications found."}}}, nil
	}

	label := "all"
	if args.Unread {
		label = "unread"
	}
	result := fmt.Sprintf("Found %d %s notifications:\n\n", list.Total, label)

	for _, n := range list.Embedded.Elements {
		read := "unread"
		if n.ReadIAN != nil && *n.ReadIAN {
			read = "read"
		}
		resource := ""
		if n.Links != nil && n.Links.Resource != nil {
			resource = " — " + n.Links.Resource.Title
		}
		project := ""
		if n.Links != nil && n.Links.Project != nil && n.Links.Project.Title != "" {
			project = " [" + n.Links.Project.Title + "]"
		}
		createdAt := ""
		if n.CreatedAt != nil {
			createdAt = " (" + n.CreatedAt.Format("2006-01-02 15:04") + ")"
		}
		result += fmt.Sprintf("- **#%d** [%s] reason: `%s`%s%s%s\n",
			n.ID, read, n.Reason, resource, project, createdAt)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) markNotificationRead(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args MarkNotificationReadArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	if err := r.client.MarkNotificationRead(ctx, args.ID); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to mark notification as read: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Notification #%d marked as read.", args.ID)}}}, nil
}

func (r *Registry) markAllNotificationsRead(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := r.client.MarkAllNotificationsRead(ctx); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to mark all notifications as read: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "All notifications marked as read."}}}, nil
}
