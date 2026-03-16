package openproject

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ListMembershipsOptions contains options for listing memberships.
type ListMembershipsOptions struct {
	Offset   int
	PageSize int
	SortBy   string
	OrderBy  string // Deprecated: use SortBy
	Select   []string
	Filters  []MembershipFilter
}

// MembershipFilter represents a filter for memberships.
type MembershipFilter struct {
	Name     string   `json:"name"`
	Values   []string `json:"values"`
	Operator string   `json:"operator,omitempty"`
}

// ListMemberships retrieves all memberships.
func (c *Client) ListMemberships(ctx context.Context, opts *ListMembershipsOptions) (*MembershipList, error) {
	if opts == nil {
		opts = &ListMembershipsOptions{}
	}

	params := url.Values{}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(opts.PageSize))
	}
	if opts.SortBy != "" {
		params.Set("sortBy", opts.SortBy)
	} else if opts.OrderBy != "" {
		params.Set("sortBy", opts.OrderBy)
	}
	if len(opts.Select) > 0 {
		params.Set("select", strings.Join(opts.Select, ","))
	}
	if len(opts.Filters) > 0 {
		filterJSON, err := jsonMarshalMembershipFilters(opts.Filters)
		if err != nil {
			return nil, err
		}
		params.Set("filters", filterJSON)
	}

	path := "/memberships"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result MembershipList
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMembership retrieves a specific membership by ID.
func (c *Client) GetMembership(ctx context.Context, id int) (*Membership, error) {
	var membership Membership
	if err := c.Get(ctx, fmt.Sprintf("/memberships/%d", id), &membership); err != nil {
		return nil, err
	}
	return &membership, nil
}

// CreateMembershipOptions contains options for creating a membership.
type CreateMembershipOptions struct {
	ProjectID int
	Principal int
	RoleIDs   []int
}

// CreateMembership creates a new project membership.
func (c *Client) CreateMembership(ctx context.Context, opts *CreateMembershipOptions) (*Membership, error) {
	roles := make([]map[string]string, len(opts.RoleIDs))
	for i, roleID := range opts.RoleIDs {
		roles[i] = map[string]string{
			"href": fmt.Sprintf("/api/v3/roles/%d", roleID),
		}
	}

	payload := map[string]interface{}{
		"_links": map[string]interface{}{
			"project": map[string]string{
				"href": fmt.Sprintf("/api/v3/projects/%d", opts.ProjectID),
			},
			"principal": map[string]string{
				"href": fmt.Sprintf("/api/v3/users/%d", opts.Principal),
			},
			"roles": roles,
		},
	}

	var membership Membership
	if err := c.Post(ctx, "/memberships", payload, &membership); err != nil {
		return nil, err
	}
	return &membership, nil
}

// UpdateMembershipOptions contains options for updating a membership.
type UpdateMembershipOptions struct {
	RoleIDs []int
}

// UpdateMembership updates an existing membership.
func (c *Client) UpdateMembership(ctx context.Context, id int, opts *UpdateMembershipOptions) (*Membership, error) {
	roles := make([]map[string]string, len(opts.RoleIDs))
	for i, roleID := range opts.RoleIDs {
		roles[i] = map[string]string{
			"href": fmt.Sprintf("/api/v3/roles/%d", roleID),
		}
	}

	payload := map[string]interface{}{
		"_links": map[string]interface{}{
			"roles": roles,
		},
	}

	var membership Membership
	if err := c.Patch(ctx, fmt.Sprintf("/memberships/%d", id), payload, &membership); err != nil {
		return nil, err
	}
	return &membership, nil
}

// DeleteMembership deletes a membership.
func (c *Client) DeleteMembership(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/memberships/%d", id))
}

// ListProjectMemberships retrieves memberships for a project.
func (c *Client) ListProjectMemberships(ctx context.Context, projectID int) (*MembershipList, error) {
	var result MembershipList
	if err := c.Get(ctx, fmt.Sprintf("/projects/%d/memberships", projectID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListRoles retrieves all roles.
func (c *Client) ListRoles(ctx context.Context) (*RoleList, error) {
	var result RoleList
	if err := c.Get(ctx, "/roles", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRole retrieves a specific role by ID.
func (c *Client) GetRole(ctx context.Context, id int) (*Role, error) {
	var role Role
	if err := c.Get(ctx, fmt.Sprintf("/roles/%d", id), &role); err != nil {
		return nil, err
	}
	return &role, nil
}

// jsonMarshalMembershipFilters marshals membership filters to JSON string.
func jsonMarshalMembershipFilters(filters []MembershipFilter) (string, error) {
	encoded := make([]map[string]map[string]interface{}, 0, len(filters))
	for _, f := range filters {
		if f.Name == "" {
			continue
		}
		op := f.Operator
		if op == "" {
			op = "="
		}
		encoded = append(encoded, map[string]map[string]interface{}{
			f.Name: {
				"operator": op,
				"values":   f.Values,
			},
		})
	}

	data, err := json.Marshal(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
