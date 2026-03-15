package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type ListMembershipsArgs struct {
	ProjectID int `json:"projectId,omitempty"`
	Offset    int `json:"offset,omitempty"`
	PageSize  int `json:"pageSize,omitempty"`
}

type GetMembershipArgs struct{ ID int }
type DeleteMembershipArgs struct{ ID int }
type CreateMembershipArgs struct {
	ProjectID int   `json:"projectId"`
	Principal int   `json:"principal"`
	RoleIDs   []int `json:"roleIds"`
}
type UpdateMembershipArgs struct {
	ID      int   `json:"id"`
	RoleIDs []int `json:"roleIds"`
}
type ListProjectMembersArgs struct{ ProjectID int }
type ListRolesArgs struct{}
type GetRoleArgs struct{ ID int }

// registerMembershipTools registers membership-related tools.
func (r *Registry) registerMembershipTools(server *mcp.Server) {
	server.AddTool(&mcp.Tool{Name: "list_memberships", Description: "List all memberships"}, r.listMemberships)
	server.AddTool(&mcp.Tool{Name: "get_membership", Description: "Get details of a specific membership"}, r.getMembership)
	server.AddTool(&mcp.Tool{Name: "create_membership", Description: "Add a user to a project with specified roles"}, r.createMembership)
	server.AddTool(&mcp.Tool{Name: "update_membership", Description: "Update roles for a membership"}, r.updateMembership)
	server.AddTool(&mcp.Tool{Name: "delete_membership", Description: "Remove a user from a project"}, r.deleteMembership)
	server.AddTool(&mcp.Tool{Name: "list_project_members", Description: "List all members of a project"}, r.listProjectMembers)
	server.AddTool(&mcp.Tool{Name: "list_roles", Description: "List all available roles"}, r.listRoles)
	server.AddTool(&mcp.Tool{Name: "get_role", Description: "Get details of a specific role"}, r.getRole)
}

func (r *Registry) listMemberships(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListMembershipsArgs
	parseArgs(req.Params.Arguments, &args)

	opts := &openproject.ListMembershipsOptions{Offset: args.Offset, PageSize: args.PageSize}
	var list *openproject.MembershipList
	var err error
	if args.ProjectID > 0 {
		list, err = r.client.ListProjectMemberships(ctx, args.ProjectID)
	} else {
		list, err = r.client.ListMemberships(ctx, opts)
	}
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list memberships: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d memberships:\n\n", list.Total)
	for _, m := range list.Embedded.Elements {
		result += fmt.Sprintf("- Membership #%d\n", m.ID)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) getMembership(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetMembershipArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	m, err := r.client.GetMembership(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get membership: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Membership #%d", m.ID)}}}, nil
}

func (r *Registry) createMembership(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateMembershipArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.CreateMembershipOptions{ProjectID: args.ProjectID, Principal: args.Principal, RoleIDs: args.RoleIDs}
	m, err := r.client.CreateMembership(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create membership: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Membership #%d created successfully!", m.ID)}}}, nil
}

func (r *Registry) updateMembership(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateMembershipArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.UpdateMembershipOptions{RoleIDs: args.RoleIDs}
	m, err := r.client.UpdateMembership(ctx, args.ID, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update membership: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Membership #%d updated successfully!", m.ID)}}}, nil
}

func (r *Registry) deleteMembership(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteMembershipArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	if err := r.client.DeleteMembership(ctx, args.ID); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete membership: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Membership #%d deleted successfully!", args.ID)}}}, nil
}

func (r *Registry) listProjectMembers(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListProjectMembersArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	list, err := r.client.ListProjectMemberships(ctx, args.ProjectID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list project members: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d members:\n\n", list.Total)
	for _, m := range list.Embedded.Elements {
		if m.Links != nil && m.Links.Principal != nil {
			result += fmt.Sprintf("- %s\n", m.Links.Principal.Title)
		}
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) listRoles(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	list, err := r.client.ListRoles(ctx)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list roles: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d roles:\n\n", list.Total)
	for _, role := range list.Embedded.Elements {
		result += fmt.Sprintf("- **%s** (ID: %d)\n", role.Name, role.ID)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) getRole(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetRoleArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	role, err := r.client.GetRole(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get role: %v", err)}}}, nil
	}

	result := fmt.Sprintf("# %s\n\n- **ID:** %d\n", role.Name, role.ID)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}
