package openproject

import (
	"context"
	"time"

	external "github.com/pinealctx/openproject"
)

// WorkPackageCommentInput describes a work package comment request.
type WorkPackageCommentInput struct {
	WorkPackageID int
	Raw           string
	Internal      bool
}

// ActivityCollection is a minimal HAL collection for work package activities.
type ActivityCollection struct {
	Embedded struct {
		Elements []Activity `json:"elements"`
	} `json:"_embedded"`
	Count int `json:"count"`
	Total int `json:"total"`
}

// Activity represents a work package activity/comment entry.
type Activity struct {
	ID        *int                  `json:"id,omitempty"`
	Type      *string               `json:"_type,omitempty"`
	Comment   *external.Formattable `json:"comment,omitempty"`
	Details   []any                 `json:"details,omitempty"`
	CreatedAt *time.Time            `json:"createdAt,omitempty"`
	UpdatedAt *time.Time            `json:"updatedAt,omitempty"`
	Internal  *bool                 `json:"internal,omitempty"`
	Links     *struct {
		Self *external.Link `json:"self,omitempty"`
		User *external.Link `json:"user,omitempty"`
	} `json:"_links,omitempty"`
}

func (c *Client) ListWorkPackageActivities(ctx context.Context, workPackageID int) (*ActivityCollection, error) {
	var list ActivityCollection
	resp, err := c.apiClient.ListWorkPackageActivities(ctx, workPackageID)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &list, DecodeResponse(resp, err, &list)
}

func (c *Client) CreateWorkPackageComment(ctx context.Context, input WorkPackageCommentInput) (*external.ActivityModel, error) {
	body := external.ActivityCommentWriteModel{
		Comment: &struct {
			Raw *string `json:"raw,omitempty"`
		}{
			Raw: ptr(input.Raw),
		},
	}
	if input.Internal {
		internal := true
		body.Internal = &internal
	}

	var activity external.ActivityModel
	resp, err := c.apiClient.CommentWorkPackage(ctx, input.WorkPackageID, nil, body)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &activity, DecodeResponse(resp, err, &activity)
}
