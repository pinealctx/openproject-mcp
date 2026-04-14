package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

type ListWorkPackageActivitiesArgs struct {
	WorkPackageID int `json:"workPackageId"`
}

type CreateWorkPackageCommentArgs struct {
	WorkPackageID int    `json:"workPackageId"`
	Raw           string `json:"raw"`
	Internal      bool   `json:"internal,omitempty"`
}

// activityCollection is a minimal HAL collection for work package activities.
type activityCollection struct {
	Embedded struct {
		Elements []activityElement `json:"elements"`
	} `json:"_embedded"`
	Count int `json:"count"`
	Total int `json:"total"`
}

// activityElement represents a single activity/comment in a collection.
type activityElement struct {
	ID        *int              `json:"id,omitempty"`
	Type      *string           `json:"_type,omitempty"`
	Comment   *external.Formattable `json:"comment,omitempty"`
	Details   []any             `json:"details,omitempty"`
	CreatedAt *time.Time        `json:"createdAt,omitempty"`
	UpdatedAt *time.Time        `json:"updatedAt,omitempty"`
	Internal  *bool             `json:"internal,omitempty"`
	Links     *struct {
		Self   *external.Link `json:"self,omitempty"`
		User   *external.Link `json:"user,omitempty"`
	} `json:"_links,omitempty"`
}

func (r *Registry) registerCommentTools(server *mcp.Server) {
	addTool(server, "list_work_package_activities",
		"List activities/comments for a work package",
		newSchema(schemaProps{
			"workPackageId": schemaInt("Work package ID"),
		}, "workPackageId"),
		r.listWorkPackageActivities)

	addTool(server, "create_work_package_comment",
		"Add a comment to a work package",
		newSchema(schemaProps{
			"workPackageId": schemaInt("Work package ID"),
			"raw":           schemaStr("Comment text in markdown"),
			"internal":      schemaBool("Whether this is an internal comment (default: false)"),
		}, "workPackageId", "raw"),
		r.createWorkPackageComment)
}

func (r *Registry) listWorkPackageActivities(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackageActivitiesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().ListWorkPackageActivities(ctx, args.WorkPackageID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list activities: %v", err), nil
	}
	var list activityCollection
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list activities: %v", err), nil
	}

	if list.Total == 0 {
		return textResult(fmt.Sprintf("No activities found for work package #%d.", args.WorkPackageID)), nil
	}

	result := fmt.Sprintf("Found %d activities for work package #%d:\n\n", list.Total, args.WorkPackageID)
	for _, a := range list.Embedded.Elements {
		actType := derefStr(a.Type)
		user := ""
		if a.Links != nil && a.Links.User != nil {
			user = derefStr(a.Links.User.Title)
		}
		createdAt := ""
		if a.CreatedAt != nil {
			createdAt = a.CreatedAt.Format("2006-01-02 15:04")
		}
		internal := ""
		if derefBool(a.Internal) {
			internal = " [internal]"
		}
		result += fmt.Sprintf("- **#%d** %s by %s%s (%s)\n",
			derefInt(a.ID), actType, user, internal, createdAt)
		if a.Comment != nil && a.Comment.Raw != nil && *a.Comment.Raw != "" {
			result += fmt.Sprintf("  > %s\n", *a.Comment.Raw)
		}
	}
	return textResult(result), nil
}

func (r *Registry) createWorkPackageComment(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateWorkPackageCommentArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.ActivityCommentWriteModel{
		Comment: &struct {
			Raw *string `json:"raw,omitempty"`
		}{
			Raw: strPtr(args.Raw),
		},
	}
	if args.Internal {
		internal := true
		body.Internal = &internal
	}

	resp, err := r.client.APIClient().CommentWorkPackage(ctx, args.WorkPackageID, nil, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to create comment: %v", err), nil
	}
	var activity external.ActivityModel
	if err := openproject.ReadResponse(resp, &activity); err != nil {
		return errorResult("Failed to create comment: %v", err), nil
	}

	return textResult(fmt.Sprintf("Comment added to work package #%d (activity #%d).", args.WorkPackageID, derefInt(activity.Id))), nil
}
