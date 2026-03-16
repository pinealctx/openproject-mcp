package tools

import (
	"context"
	"fmt"
	"strings"

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
	addTool(server, "list_memberships", "List all memberships, optionally filtered by project",
		newSchema(schemaProps{
			"projectId": schemaInt("Filter by project ID"),
			"offset":    schemaInt("Pagination offset"),
			"pageSize":  schemaInt("Items per page"),
		}),
		r.listMemberships)

	addTool(server, "get_membership", "Get details of a specific membership",
		newSchema(schemaProps{"id": schemaInt("Membership ID")}, "id"),
		r.getMembership)

	addTool(server, "create_membership", "Add a user to a project with specified roles",
		newSchema(schemaProps{
			"projectId": schemaInt("Project ID"),
			"principal": schemaInt("User ID to add"),
			"roleIds":   {Type: "array", Description: "List of role IDs to assign", Items: schemaInt("Role ID")},
		}, "projectId", "principal", "roleIds"),
		r.createMembership)

	addTool(server, "update_membership", "Update roles for a membership",
		newSchema(schemaProps{
			"id":      schemaInt("Membership ID"),
			"roleIds": {Type: "array", Description: "New list of role IDs", Items: schemaInt("Role ID")},
		}, "id", "roleIds"),
		r.updateMembership)

	addTool(server, "delete_membership", "Remove a user from a project",
		newSchema(schemaProps{"id": schemaInt("Membership ID")}, "id"),
		r.deleteMembership)

	addTool(server, "list_project_members", "List all members of a project",
		newSchema(schemaProps{"projectId": schemaInt("Project ID")}, "projectId"),
		r.listProjectMembers)

	addTool(server, "list_roles", "List all available roles",
		noSchema, r.listRoles)

	addTool(server, "get_role", "Get details of a specific role",
		newSchema(schemaProps{"id": schemaInt("Role ID")}, "id"),
		r.getRole)
}

func (r *Registry) listMemberships(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListMembershipsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

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
		principal := "Unknown"
		if m.Links != nil && m.Links.Principal != nil {
			principal = m.Links.Principal.Title
		}
		roles := []string{}
		if m.Links != nil {
			for _, r := range m.Links.Roles {
				if r != nil {
					roles = append(roles, r.Title)
				}
			}
		}
		rolesStr := strings.Join(roles, ", ")
		if rolesStr == "" {
			rolesStr = "No roles"
		}
		result += fmt.Sprintf("- #%d **%s** — Roles: %s\n", m.ID, principal, rolesStr)
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

	principal, project := "Unknown", "Unknown"
	if m.Links != nil {
		if m.Links.Principal != nil {
			principal = m.Links.Principal.Title
		}
		if m.Links.Project != nil {
			project = m.Links.Project.Title
		}
	}
	roles := []string{}
	if m.Links != nil {
		for _, role := range m.Links.Roles {
			if role != nil {
				roles = append(roles, role.Title)
			}
		}
	}

	result := fmt.Sprintf("# Membership #%d\n\n", m.ID)
	result += fmt.Sprintf("- **Principal:** %s\n", principal)
	result += fmt.Sprintf("- **Project:** %s\n", project)
	result += fmt.Sprintf("- **Roles:** %s\n", strings.Join(roles, ", "))
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
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
			roles := []string{}
			for _, role := range m.Links.Roles {
				if role != nil {
					roles = append(roles, role.Title)
				}
			}
			result += fmt.Sprintf("- **%s** — %s\n", m.Links.Principal.Title, strings.Join(roles, ", "))
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
