package openproject

import (
	"context"
	"fmt"
	"strconv"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	external "github.com/pinealctx/openproject"
)

// WorkPackageListInput describes a work package collection request.
type WorkPackageListInput struct {
	ProjectID int
	Offset    int
	PageSize  int
	SortBy    string
	Filters   string
}

// WorkPackageCreateInput describes fields supported by both the CLI and MCP tool.
type WorkPackageCreateInput struct {
	ProjectID     int
	Subject       string
	Description   string
	TypeID        string
	StatusID      string
	PriorityID    string
	AssigneeID    int
	StartDate     string
	DueDate       string
	EstimatedTime string
}

// WorkPackageUpdateInput describes a work package patch.
type WorkPackageUpdateInput struct {
	ID             int
	Subject        string
	Description    string
	StatusID       string
	PriorityID     string
	AssigneeID     int
	StartDate      string
	DueDate        string
	EstimatedTime  string
	PercentageDone *int
	ParentID       *int
	ClearParent    bool
}

func (c *Client) ListWorkPackages(ctx context.Context, input WorkPackageListInput) (*external.WorkPackagesModel, error) {
	var list external.WorkPackagesModel
	if input.ProjectID > 0 {
		params := &external.GetProjectWorkPackageCollectionParams{}
		applyWorkPackageListParams(&params.Offset, &params.PageSize, &params.SortBy, &params.Filters, input)
		resp, err := c.apiClient.GetProjectWorkPackageCollection(ctx, input.ProjectID, params)
		if resp != nil {
			defer func() { _ = resp.Body.Close() }()
		}
		return &list, DecodeResponse(resp, err, &list)
	}

	params := &external.ListWorkPackagesParams{}
	applyWorkPackageListParams(&params.Offset, &params.PageSize, &params.SortBy, &params.Filters, input)
	resp, err := c.apiClient.ListWorkPackages(ctx, params)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &list, DecodeResponse(resp, err, &list)
}

func (c *Client) GetWorkPackage(ctx context.Context, id int) (*external.WorkPackageModel, error) {
	var wp external.WorkPackageModel
	resp, err := c.apiClient.ViewWorkPackage(ctx, id, nil)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &wp, DecodeResponse(resp, err, &wp)
}

func (c *Client) CreateWorkPackage(ctx context.Context, input WorkPackageCreateInput) (*external.WorkPackageModel, error) {
	body := external.WorkPackageModel{
		Subject: input.Subject,
	}
	if input.Description != "" {
		format := external.FormattableFormat("markdown")
		body.Description = &external.Formattable{Format: &format, Raw: ptr(input.Description)}
	}
	if input.StartDate != "" {
		body.StartDate = parseDate(input.StartDate)
	}
	if input.DueDate != "" {
		body.DueDate = parseDate(input.DueDate)
	}
	if input.EstimatedTime != "" {
		body.EstimatedTime = ptr(input.EstimatedTime)
	}
	if input.TypeID != "" {
		body.UnderscoreLinks.Type = external.Link{Href: ptr("/api/v3/types/" + input.TypeID)}
	}
	if input.StatusID != "" {
		body.UnderscoreLinks.Status = external.Link{Href: ptr("/api/v3/statuses/" + input.StatusID)}
	}
	if input.PriorityID != "" {
		body.UnderscoreLinks.Priority = external.Link{Href: ptr("/api/v3/priorities/" + input.PriorityID)}
	}
	if input.AssigneeID > 0 {
		body.UnderscoreLinks.Assignee = &external.Link{Href: ptr(fmt.Sprintf("/api/v3/users/%d", input.AssigneeID))}
	}

	var wp external.WorkPackageModel
	resp, err := c.apiClient.CreateProjectWorkPackage(ctx, input.ProjectID, nil, body)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &wp, DecodeResponse(resp, err, &wp)
}

func (c *Client) UpdateWorkPackage(ctx context.Context, input WorkPackageUpdateInput) (*external.WorkPackageModel, error) {
	current, err := c.GetWorkPackage(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("fetch work package lock version: %w", err)
	}

	if current.LockVersion == nil {
		return nil, fmt.Errorf("fetch work package lock version: response missing lockVersion")
	}

	body := input.patchBody(*current.LockVersion)
	var wp external.WorkPackageModel
	if err := c.Patch(ctx, fmt.Sprintf("/work_packages/%d", input.ID), body, &wp); err != nil {
		return nil, err
	}
	return &wp, nil
}

func (c *Client) SetWorkPackageParent(ctx context.Context, workPackageID, parentID int) (*external.WorkPackageModel, error) {
	return c.UpdateWorkPackage(ctx, WorkPackageUpdateInput{
		ID:       workPackageID,
		ParentID: &parentID,
	})
}

func (c *Client) RemoveWorkPackageParent(ctx context.Context, workPackageID int) (*external.WorkPackageModel, error) {
	return c.UpdateWorkPackage(ctx, WorkPackageUpdateInput{
		ID:          workPackageID,
		ClearParent: true,
	})
}

func (c *Client) ListWorkPackageChildren(ctx context.Context, workPackageID int) (*external.WorkPackagesModel, error) {
	return c.ListWorkPackages(ctx, WorkPackageListInput{
		Filters: fmt.Sprintf(`[{"parent":{"operator":"=","values":["%d"]}}]`, workPackageID),
	})
}

func (c *Client) DeleteWorkPackage(ctx context.Context, id int) error {
	resp, err := c.apiClient.DeleteWorkPackage(ctx, id)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return DecodeResponse(resp, err, nil)
}

func (c *Client) ListWorkPackageTypes(ctx context.Context, projectID int) (*external.TypesByWorkspaceModel, error) {
	var list external.TypesByWorkspaceModel
	if projectID > 0 {
		httpResp, err := c.apiClient.ListTypesAvailableInAProject(ctx, projectID)
		if httpResp != nil {
			defer func() { _ = httpResp.Body.Close() }()
		}
		return &list, DecodeResponse(httpResp, err, &list)
	}
	httpResp, err := c.apiClient.ListAllTypes(ctx)
	if httpResp != nil {
		defer func() { _ = httpResp.Body.Close() }()
	}
	return &list, DecodeResponse(httpResp, err, &list)
}

func (c *Client) ListWorkPackageStatuses(ctx context.Context) (*external.StatusCollectionModel, error) {
	var list external.StatusCollectionModel
	resp, err := c.apiClient.ListStatuses(ctx)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &list, DecodeResponse(resp, err, &list)
}

func (c *Client) ListWorkPackagePriorities(ctx context.Context) (*external.PriorityCollectionModel, error) {
	var list external.PriorityCollectionModel
	resp, err := c.apiClient.ListAllPriorities(ctx)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &list, DecodeResponse(resp, err, &list)
}

func (c *Client) ListWorkPackageAvailableAssignees(ctx context.Context, workPackageID int) (*external.AvailableAssigneesModel, error) {
	var list external.AvailableAssigneesModel
	resp, err := c.apiClient.WorkPackageAvailableAssignees(ctx, workPackageID)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &list, DecodeResponse(resp, err, &list)
}

func (input WorkPackageUpdateInput) patchBody(lockVersion int) map[string]any {
	body := map[string]any{"lockVersion": lockVersion}
	if input.Subject != "" {
		body["subject"] = input.Subject
	}
	if input.Description != "" {
		body["description"] = map[string]any{
			"format": "markdown",
			"raw":    input.Description,
		}
	}
	if input.StartDate != "" {
		body["startDate"] = input.StartDate
	}
	if input.DueDate != "" {
		body["dueDate"] = input.DueDate
	}
	if input.EstimatedTime != "" {
		body["estimatedTime"] = input.EstimatedTime
	}
	if input.PercentageDone != nil {
		body["percentageDone"] = *input.PercentageDone
	}

	links := map[string]any{}
	if input.StatusID != "" {
		links["status"] = map[string]any{"href": "/api/v3/statuses/" + input.StatusID}
	}
	if input.PriorityID != "" {
		links["priority"] = map[string]any{"href": "/api/v3/priorities/" + input.PriorityID}
	}
	if input.AssigneeID > 0 {
		links["assignee"] = map[string]any{"href": fmt.Sprintf("/api/v3/users/%d", input.AssigneeID)}
	}
	if input.ParentID != nil {
		links["parent"] = map[string]any{"href": fmt.Sprintf("/api/v3/work_packages/%d", *input.ParentID)}
	}
	if input.ClearParent {
		links["parent"] = map[string]any{"href": nil}
	}
	if len(links) > 0 {
		body["_links"] = links
	}

	return body
}

func applyWorkPackageListParams(offset, pageSize **int, sortBy, filters **string, input WorkPackageListInput) {
	if input.Offset > 0 {
		*offset = ptr(input.Offset)
	}
	if input.PageSize > 0 {
		*pageSize = ptr(input.PageSize)
	}
	if input.SortBy != "" {
		*sortBy = ptr(input.SortBy)
	}
	if input.Filters != "" {
		*filters = ptr(input.Filters)
	}
}

func parseDate(s string) *openapi_types.Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &openapi_types.Date{Time: t}
}

func IntID(id int) string {
	if id <= 0 {
		return ""
	}
	return strconv.Itoa(id)
}
