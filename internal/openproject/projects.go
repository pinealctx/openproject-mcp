package openproject

import (
	"context"

	external "github.com/pinealctx/openproject"
)

type ProjectListInput struct {
	SortBy  string
	Filters string
}

type ProjectCreateInput struct {
	Name        string
	Identifier  string
	Description string
	Public      bool
}

type ProjectUpdateInput struct {
	ID          int
	Name        string
	Description string
	Public      *bool
	Active      *bool
}

func (c *Client) ListProjects(ctx context.Context, input ProjectListInput) (*external.ProjectCollectionModel, error) {
	params := &external.ListProjectsParams{}
	if input.SortBy != "" {
		params.SortBy = ptr(input.SortBy)
	}
	if input.Filters != "" {
		params.Filters = ptr(input.Filters)
	}

	var projects external.ProjectCollectionModel
	resp, err := c.apiClient.ListProjects(ctx, params)
	return &projects, DecodeResponse(resp, err, &projects)
}

func (c *Client) GetProject(ctx context.Context, id int) (*external.ProjectModel, error) {
	var project external.ProjectModel
	resp, err := c.apiClient.ViewProject(ctx, id)
	return &project, DecodeResponse(resp, err, &project)
}

func (c *Client) CreateProject(ctx context.Context, input ProjectCreateInput) (*external.ProjectModel, error) {
	body := external.ProjectModel{
		Identifier: ptr(input.Identifier),
		Name:       ptr(input.Name),
		Public:     ptr(input.Public),
	}
	if input.Description != "" {
		format := external.FormattableFormat("markdown")
		body.Description = &external.Formattable{Format: &format, Raw: ptr(input.Description)}
	}

	var project external.ProjectModel
	resp, err := c.apiClient.CreateProject(ctx, body)
	return &project, DecodeResponse(resp, err, &project)
}

func (c *Client) UpdateProject(ctx context.Context, input ProjectUpdateInput) (*external.ProjectModel, error) {
	body := external.ProjectModel{
		Public: input.Public,
		Active: input.Active,
	}
	if input.Name != "" {
		body.Name = ptr(input.Name)
	}
	if input.Description != "" {
		format := external.FormattableFormat("markdown")
		body.Description = &external.Formattable{Format: &format, Raw: ptr(input.Description)}
	}

	var project external.ProjectModel
	resp, err := c.apiClient.UpdateProject(ctx, input.ID, body)
	return &project, DecodeResponse(resp, err, &project)
}

func (c *Client) DeleteProject(ctx context.Context, id int) error {
	resp, err := c.apiClient.DeleteProject(ctx, id)
	return DecodeResponse(resp, err, nil)
}
