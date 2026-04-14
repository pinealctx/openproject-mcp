package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
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
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListNotificationsParams{}
	if args.Offset > 0 {
		params.Offset = intPtr(args.Offset)
	}
	if args.PageSize > 0 {
		params.PageSize = intPtr(args.PageSize)
	}
	if args.Unread {
		params.Filters = strPtr(`[{"readIAN":{"operator":"!","values":["t"]}}]`)
	}

	resp, err := r.client.APIClient().ListNotifications(ctx, params)
	if err != nil {
		return errorResult("Failed to list notifications: %v", err), nil
	}

	var list external.NotificationCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list notifications: %v", err), nil
	}

	if list.Total == 0 {
		return textResult("No notifications found."), nil
	}

	label := "all"
	if args.Unread {
		label = "unread"
	}
	result := fmt.Sprintf("Found %d %s notifications:\n\n", list.Total, label)

	for _, n := range list.UnderscoreEmbedded.Elements {
		read := "unread"
		if n.ReadIAN != nil && *n.ReadIAN {
			read = "read"
		}
		resource := ""
		if n.UnderscoreLinks != nil && n.UnderscoreLinks.Resource.Title != nil {
			resource = " — " + *n.UnderscoreLinks.Resource.Title
		}
		project := ""
		if n.UnderscoreLinks != nil && n.UnderscoreLinks.Project.Title != nil {
			project = " [" + *n.UnderscoreLinks.Project.Title + "]"
		}
		createdAt := ""
		if n.CreatedAt != nil {
			createdAt = " (" + n.CreatedAt.Format("2006-01-02 15:04") + ")"
		}
		reason := ""
		if n.Reason != nil {
			reason = string(*n.Reason)
		}
		result += fmt.Sprintf("- **#%d** [%s] reason: `%s`%s%s%s\n",
			derefInt(n.Id), read, reason, resource, project, createdAt)
	}
	return textResult(result), nil
}

func (r *Registry) markNotificationRead(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args MarkNotificationReadArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().ReadNotification(ctx, args.ID)
	if err != nil {
		return errorResult("Failed to mark notification as read: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to mark notification as read: %v", err), nil
	}
	return textResult(fmt.Sprintf("Notification #%d marked as read.", args.ID)), nil
}

func (r *Registry) markAllNotificationsRead(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := r.client.APIClient().ReadNotifications(ctx, nil)
	if err != nil {
		return errorResult("Failed to mark all notifications as read: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to mark all notifications as read: %v", err), nil
	}
	return textResult("All notifications marked as read."), nil
}
