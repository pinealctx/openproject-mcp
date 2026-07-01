package openproject

import (
	"context"

	external "github.com/pinealctx/openproject"
)

type GroupListInput struct {
	SortBy string
}

type GroupCreateInput struct {
	Name string
}

type GroupUpdateInput struct {
	ID   int
	Name string
}

func (c *Client) ListGroups(ctx context.Context, input GroupListInput) (*external.GroupCollectionModel, error) {
	params := &external.ListGroupsParams{}
	if input.SortBy != "" {
		params.SortBy = ptr(input.SortBy)
	}

	var groups external.GroupCollectionModel
	resp, err := c.apiClient.ListGroups(ctx, params)
	return &groups, DecodeResponse(resp, err, &groups)
}

func (c *Client) GetGroup(ctx context.Context, id int) (*external.GroupModel, error) {
	var group external.GroupModel
	resp, err := c.apiClient.GetGroup(ctx, id)
	return &group, DecodeResponse(resp, err, &group)
}

func (c *Client) CreateGroup(ctx context.Context, input GroupCreateInput) (*external.GroupModel, error) {
	body := external.GroupWriteModel{}
	if input.Name != "" {
		body.Name = ptr(input.Name)
	}

	var group external.GroupModel
	resp, err := c.apiClient.CreateGroup(ctx, body)
	return &group, DecodeResponse(resp, err, &group)
}

func (c *Client) UpdateGroup(ctx context.Context, input GroupUpdateInput) (*external.GroupModel, error) {
	body := external.GroupWriteModel{}
	if input.Name != "" {
		body.Name = ptr(input.Name)
	}

	var group external.GroupModel
	resp, err := c.apiClient.UpdateGroup(ctx, input.ID, body)
	return &group, DecodeResponse(resp, err, &group)
}

func (c *Client) DeleteGroup(ctx context.Context, id int) error {
	resp, err := c.apiClient.DeleteGroup(ctx, id)
	return DecodeResponse(resp, err, nil)
}
