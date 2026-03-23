package openproject

import (
	"context"
	"encoding/json"
	"fmt"
)

// SetWorkPackageParent sets the parent of a work package.
// It automatically fetches the current lockVersion to satisfy OpenProject's optimistic concurrency check.
func (c *Client) SetWorkPackageParent(ctx context.Context, workPackageID, parentID int) (*WorkPackage, error) {
	// Fetch current work package to get lockVersion
	current, err := c.GetWorkPackage(ctx, workPackageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current work package for lockVersion: %w", err)
	}

	payload := map[string]interface{}{
		"lockVersion": current.LockVersion,
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
// It automatically fetches the current lockVersion to satisfy OpenProject's optimistic concurrency check.
func (c *Client) RemoveWorkPackageParent(ctx context.Context, workPackageID int) (*WorkPackage, error) {
	// Fetch current work package to get lockVersion
	current, err := c.GetWorkPackage(ctx, workPackageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current work package for lockVersion: %w", err)
	}

	payload := map[string]interface{}{
		"lockVersion": current.LockVersion,
		"_links": map[string]interface{}{
			"parent": map[string]interface{}{"href": nil},
		},
	}

	var wp WorkPackage
	if err := c.Patch(ctx, fmt.Sprintf("/work_packages/%d", workPackageID), payload, &wp); err != nil {
		return nil, err
	}
	return &wp, nil
}

// ListWorkPackageChildren retrieves the children of a work package.
// It fetches the parent work package and parses its _links.children to get child IDs,
// then fetches each child work package individually.
func (c *Client) ListWorkPackageChildren(ctx context.Context, workPackageID int) (*WorkPackageList, error) {
	// Get the parent work package to access its _links.children
	parent, err := c.GetWorkPackage(ctx, workPackageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch work package: %w", err)
	}

	result := &WorkPackageList{}

	// Check if there are children links
	if parent.Links == nil || len(parent.Links.Children) == 0 {
		return result, nil
	}

	// Parse children links - they're stored as json.RawMessage
	var childLinks []struct {
		Href string `json:"href"`
	}
	if err := json.Unmarshal(parent.Links.Children, &childLinks); err != nil {
		return nil, fmt.Errorf("failed to parse children links: %w", err)
	}

	// Fetch each child work package
	for _, link := range childLinks {
		// Extract ID from href like "/api/v3/work_packages/123"
		var childID int
		if _, err := fmt.Sscanf(link.Href, "/api/v3/work_packages/%d", &childID); err != nil {
			continue // Skip malformed links
		}

		child, err := c.GetWorkPackage(ctx, childID)
		if err != nil {
			continue // Skip if child can't be fetched
		}
		result.Embedded.Elements = append(result.Embedded.Elements, *child)
	}

	result.Count = len(result.Embedded.Elements)
	result.Total = result.Count
	return result, nil
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
