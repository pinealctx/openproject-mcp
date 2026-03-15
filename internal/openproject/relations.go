package openproject

import (
	"context"
	"fmt"
)

// SetWorkPackageParent sets the parent of a work package.
func (c *Client) SetWorkPackageParent(ctx context.Context, workPackageID, parentID int) (*WorkPackage, error) {
	payload := map[string]interface{}{
		"_links": map[string]interface{}{
			"parent": map[string]string{
				"href": fmt.Sprintf("/api/v3/work_packages/%d", parentID),
			},
		},
	}

	var wp WorkPackage
	if err := c.Patch(ctx, fmt.Sprintf("/work_packages/%d", workPackageID), payload, &wp); err != nil {
		return nil, err
	}
	return &wp, nil
}

// RemoveWorkPackageParent removes the parent of a work package.
func (c *Client) RemoveWorkPackageParent(ctx context.Context, workPackageID int) (*WorkPackage, error) {
	payload := map[string]interface{}{
		"_links": map[string]interface{}{
			"parent": nil,
		},
	}

	var wp WorkPackage
	if err := c.Patch(ctx, fmt.Sprintf("/work_packages/%d", workPackageID), payload, &wp); err != nil {
		return nil, err
	}
	return &wp, nil
}

// ListWorkPackageChildren retrieves the children of a work package.
func (c *Client) ListWorkPackageChildren(ctx context.Context, workPackageID int) (*WorkPackageList, error) {
	var result WorkPackageList
	if err := c.Get(ctx, fmt.Sprintf("/work_packages/%d/children", workPackageID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateRelationOptions contains options for creating a relation.
type CreateRelationOptions struct {
	FromID      int    `json:"-"`
	ToID        int    `json:"-"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Delay       int    `json:"delay,omitempty"`
}

// CreateRelation creates a new relation between work packages.
func (c *Client) CreateRelation(ctx context.Context, opts *CreateRelationOptions) (*Relation, error) {
	payload := map[string]interface{}{
		"type":        opts.Type,
		"description": opts.Description,
		"delay":       opts.Delay,
		"_links": map[string]interface{}{
			"from": map[string]string{
				"href": fmt.Sprintf("/api/v3/work_packages/%d", opts.FromID),
			},
			"to": map[string]string{
				"href": fmt.Sprintf("/api/v3/work_packages/%d", opts.ToID),
			},
		},
	}

	var relation Relation
	if err := c.Post(ctx, "/relations", payload, &relation); err != nil {
		return nil, err
	}
	return &relation, nil
}

// ListWorkPackageRelations retrieves all relations for a work package.
func (c *Client) ListWorkPackageRelations(ctx context.Context, workPackageID int) (*RelationList, error) {
	var result RelationList
	if err := c.Get(ctx, fmt.Sprintf("/work_packages/%d/relations", workPackageID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRelation retrieves a specific relation by ID.
func (c *Client) GetRelation(ctx context.Context, id int) (*Relation, error) {
	var relation Relation
	if err := c.Get(ctx, fmt.Sprintf("/relations/%d", id), &relation); err != nil {
		return nil, err
	}
	return &relation, nil
}

// UpdateRelationOptions contains options for updating a relation.
type UpdateRelationOptions struct {
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Delay       int    `json:"delay,omitempty"`
}

// UpdateRelation updates an existing relation.
func (c *Client) UpdateRelation(ctx context.Context, id int, opts *UpdateRelationOptions) (*Relation, error) {
	payload := map[string]interface{}{}
	if opts.Type != "" {
		payload["type"] = opts.Type
	}
	if opts.Description != "" {
		payload["description"] = opts.Description
	}
	if opts.Delay != 0 {
		payload["delay"] = opts.Delay
	}

	var relation Relation
	if err := c.Patch(ctx, fmt.Sprintf("/relations/%d", id), payload, &relation); err != nil {
		return nil, err
	}
	return &relation, nil
}

// DeleteRelation deletes a relation.
func (c *Client) DeleteRelation(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/relations/%d", id))
}
