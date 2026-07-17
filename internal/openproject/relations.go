package openproject

import (
	"context"
	"fmt"

	external "github.com/pinealctx/openproject"
)

type RelationListInput struct {
	Filters string
	SortBy  string
}

type RelationCreateInput struct {
	FromID      int
	ToID        int
	Type        string
	Description string
	Delay       int
}

type RelationUpdateInput struct {
	ID          int
	Type        string
	Description string
	Delay       int
}

func (c *Client) ListRelations(ctx context.Context, input RelationListInput) (*external.RelationCollectionModel, error) {
	params := &external.ListRelationsParams{}
	if input.Filters != "" {
		params.Filters = ptr(input.Filters)
	}
	if input.SortBy != "" {
		params.SortBy = ptr(input.SortBy)
	}

	var relations external.RelationCollectionModel
	resp, err := c.apiClient.ListRelations(ctx, params)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &relations, DecodeResponse(resp, err, &relations)
}

func (c *Client) ListRelationsFromWorkPackage(ctx context.Context, workPackageID int) (*external.RelationCollectionModel, error) {
	return c.ListRelations(ctx, RelationListInput{
		Filters: fmt.Sprintf(`[{"from":{"operator":"=","values":["%d"]}}]`, workPackageID),
	})
}

func (c *Client) ListRelationsInvolvingWorkPackage(ctx context.Context, workPackageID int) (*external.RelationCollectionModel, error) {
	return c.ListRelations(ctx, RelationListInput{
		Filters: fmt.Sprintf(`[{"involved":{"operator":"=","values":["%d"]}}]`, workPackageID),
	})
}

func (c *Client) GetRelation(ctx context.Context, id int) (*external.RelationReadModel, error) {
	var relation external.RelationReadModel
	resp, err := c.apiClient.GetRelation(ctx, id)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &relation, DecodeResponse(resp, err, &relation)
}

func (c *Client) CreateRelation(ctx context.Context, input RelationCreateInput) (*external.RelationReadModel, error) {
	body := external.RelationWriteModel{
		Type: external.RelationWriteModelType(input.Type),
	}
	body.UnderscoreLinks.To = &external.Link{Href: ptr(fmt.Sprintf("/api/v3/work_packages/%d", input.ToID))}
	if input.Description != "" {
		body.Description = ptr(input.Description)
	}
	if input.Delay > 0 {
		body.Lag = ptr(input.Delay)
	}

	var relation external.RelationReadModel
	resp, err := c.apiClient.CreateRelation(ctx, input.FromID, body)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &relation, DecodeResponse(resp, err, &relation)
}

func (c *Client) UpdateRelation(ctx context.Context, input RelationUpdateInput) (*external.RelationReadModel, error) {
	body := map[string]any{}
	if input.Type != "" {
		body["type"] = input.Type
	}
	if input.Description != "" {
		body["description"] = input.Description
	}
	if input.Delay > 0 {
		body["lag"] = input.Delay
	}

	var relation external.RelationReadModel
	if err := c.Patch(ctx, fmt.Sprintf("/relations/%d", input.ID), body, &relation); err != nil {
		return nil, err
	}
	return &relation, nil
}

func (c *Client) DeleteRelation(ctx context.Context, id int) error {
	resp, err := c.apiClient.DeleteRelation(ctx, id)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return DecodeResponse(resp, err, nil)
}
