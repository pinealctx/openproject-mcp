package openproject

import (
	"context"
	"fmt"

	external "github.com/pinealctx/openproject"
)

type MembershipListInput struct {
	ProjectID int
	Filters   string
	SortBy    string
}

type MembershipCreateInput struct {
	ProjectID   int
	PrincipalID int
	RoleIDs     []int
}

type MembershipUpdateInput struct {
	ID      int
	RoleIDs []int
}

func (c *Client) ListMemberships(ctx context.Context, input MembershipListInput) (*external.MembershipCollectionModel, error) {
	params := &external.ListMembershipsParams{}
	if input.Filters != "" {
		params.Filters = ptr(input.Filters)
	}
	if input.ProjectID > 0 {
		params.Filters = ptr(fmt.Sprintf(`[{"project":{"operator":"=","values":["%d"]}}]`, input.ProjectID))
	}
	if input.SortBy != "" {
		params.SortBy = ptr(input.SortBy)
	}

	var memberships external.MembershipCollectionModel
	resp, err := c.apiClient.ListMemberships(ctx, params)
	return &memberships, DecodeResponse(resp, err, &memberships)
}

func (c *Client) GetMembership(ctx context.Context, id int) (*external.MembershipReadModel, error) {
	var membership external.MembershipReadModel
	resp, err := c.apiClient.GetMembership(ctx, id)
	return &membership, DecodeResponse(resp, err, &membership)
}

func (c *Client) CreateMembership(ctx context.Context, input MembershipCreateInput) (*external.MembershipReadModel, error) {
	body := membershipWriteBody(input.ProjectID, input.PrincipalID, input.RoleIDs)

	var membership external.MembershipReadModel
	resp, err := c.apiClient.CreateMembership(ctx, body)
	return &membership, DecodeResponse(resp, err, &membership)
}

func (c *Client) UpdateMembership(ctx context.Context, input MembershipUpdateInput) (*external.MembershipReadModel, error) {
	body := membershipWriteBody(0, 0, input.RoleIDs)

	var membership external.MembershipReadModel
	resp, err := c.apiClient.UpdateMembership(ctx, input.ID, body)
	return &membership, DecodeResponse(resp, err, &membership)
}

func (c *Client) DeleteMembership(ctx context.Context, id int) error {
	resp, err := c.apiClient.DeleteMembership(ctx, id)
	return DecodeResponse(resp, err, nil)
}

func (c *Client) ListRolesRaw(ctx context.Context) (map[string]any, error) {
	resp, err := c.apiClient.ListRoles(ctx, nil)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := ReadResponseRawTo(resp, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (c *Client) GetRole(ctx context.Context, id int) (*external.RoleModel, error) {
	var role external.RoleModel
	resp, err := c.apiClient.ViewRole(ctx, id)
	return &role, DecodeResponse(resp, err, &role)
}

func membershipWriteBody(projectID, principalID int, roleIDs []int) external.MembershipWriteModel {
	roleLinks := make([]external.Link, len(roleIDs))
	for i, roleID := range roleIDs {
		roleLinks[i] = external.Link{Href: ptr(fmt.Sprintf("/api/v3/roles/%d", roleID))}
	}

	body := external.MembershipWriteModel{}
	if principalID > 0 {
		body.UnderscoreLinks.Principal = &external.Link{Href: ptr(fmt.Sprintf("/api/v3/users/%d", principalID))}
	}
	if projectID > 0 {
		body.UnderscoreLinks.Project = &external.Link{Href: ptr(fmt.Sprintf("/api/v3/projects/%d", projectID))}
	}
	body.UnderscoreLinks.Roles = &roleLinks
	return body
}
