package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
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
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListMembershipsParams{}
	if args.ProjectID > 0 {
		params.Filters = strPtr(fmt.Sprintf(`[{"project":{"operator":"=","values":["%d"]}}]`, args.ProjectID))
	}

	resp, err := r.client.APIClient().ListMemberships(ctx, params)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list memberships: %v", err), nil
	}

	var list external.MembershipCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list memberships: %v", err), nil
	}

	result := fmt.Sprintf("Found %d memberships:\n\n", list.Total)
	for _, m := range list.UnderscoreEmbedded.Elements {
		principal := "Unknown"
		if m.UnderscoreLinks.Principal.Title != nil {
			principal = *m.UnderscoreLinks.Principal.Title
		}
		roles := []string{}
		for _, rl := range m.UnderscoreLinks.Roles {
			if rl.Title != nil {
				roles = append(roles, *rl.Title)
			}
		}
		rolesStr := strings.Join(roles, ", ")
		if rolesStr == "" {
			rolesStr = "No roles"
		}
		result += fmt.Sprintf("- #%d **%s** — Roles: %s\n", m.Id, principal, rolesStr)
	}
	return textResult(result), nil
}

func (r *Registry) getMembership(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetMembershipArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().GetMembership(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get membership: %v", err), nil
	}
	var m external.MembershipReadModel
	if err := openproject.ReadResponse(resp, &m); err != nil {
		return errorResult("Failed to get membership: %v", err), nil
	}

	principal, project := "Unknown", "Unknown"
	if m.UnderscoreLinks.Principal.Title != nil {
		principal = *m.UnderscoreLinks.Principal.Title
	}
	if m.UnderscoreLinks.Project.Title != nil {
		project = *m.UnderscoreLinks.Project.Title
	}
	roles := []string{}
	for _, role := range m.UnderscoreLinks.Roles {
		if role.Title != nil {
			roles = append(roles, *role.Title)
		}
	}

	result := fmt.Sprintf("# Membership #%d\n\n", m.Id)
	result += fmt.Sprintf("- **Principal:** %s\n", principal)
	result += fmt.Sprintf("- **Project:** %s\n", project)
	result += fmt.Sprintf("- **Roles:** %s\n", strings.Join(roles, ", "))
	return textResult(result), nil
}

func (r *Registry) createMembership(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateMembershipArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	roleLinks := make([]external.Link, len(args.RoleIDs))
	for i, rid := range args.RoleIDs {
		roleLinks[i] = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/roles/%d", rid))}
	}

	body := external.MembershipWriteModel{
		UnderscoreLinks: struct {
			Principal *external.Link   `json:"principal,omitempty"`
			Project   *external.Link   `json:"project,omitempty"`
			Roles     *[]external.Link `json:"roles,omitempty"`
		}{
			Principal: &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/users/%d", args.Principal))},
			Project:   &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/projects/%d", args.ProjectID))},
			Roles:     &roleLinks,
		},
	}

	resp, err := r.client.APIClient().CreateMembership(ctx, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to create membership: %v", err), nil
	}
	var m external.MembershipReadModel
	if err := openproject.ReadResponse(resp, &m); err != nil {
		return errorResult("Failed to create membership: %v", err), nil
	}
	return textResult(fmt.Sprintf("Membership #%d created successfully!", m.Id)), nil
}

func (r *Registry) updateMembership(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateMembershipArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	roleLinks := make([]external.Link, len(args.RoleIDs))
	for i, rid := range args.RoleIDs {
		roleLinks[i] = external.Link{Href: strPtr(fmt.Sprintf("/api/v3/roles/%d", rid))}
	}

	body := external.MembershipWriteModel{
		UnderscoreLinks: struct {
			Principal *external.Link   `json:"principal,omitempty"`
			Project   *external.Link   `json:"project,omitempty"`
			Roles     *[]external.Link `json:"roles,omitempty"`
		}{
			Roles: &roleLinks,
		},
	}

	resp, err := r.client.APIClient().UpdateMembership(ctx, args.ID, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to update membership: %v", err), nil
	}
	var m external.MembershipReadModel
	if err := openproject.ReadResponse(resp, &m); err != nil {
		return errorResult("Failed to update membership: %v", err), nil
	}
	return textResult(fmt.Sprintf("Membership #%d updated successfully!", m.Id)), nil
}

func (r *Registry) deleteMembership(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteMembershipArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().DeleteMembership(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to delete membership: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to delete membership: %v", err), nil
	}
	return textResult(fmt.Sprintf("Membership #%d deleted successfully!", args.ID)), nil
}

func (r *Registry) listProjectMembers(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListProjectMembersArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListMembershipsParams{
		Filters: strPtr(fmt.Sprintf(`[{"project":{"operator":"=","values":["%d"]}}]`, args.ProjectID)),
	}

	resp, err := r.client.APIClient().ListMemberships(ctx, params)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list project members: %v", err), nil
	}

	var list external.MembershipCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list project members: %v", err), nil
	}

	result := fmt.Sprintf("Found %d members:\n\n", list.Total)
	for _, m := range list.UnderscoreEmbedded.Elements {
		if m.UnderscoreLinks.Principal.Title != nil {
			roles := []string{}
			for _, role := range m.UnderscoreLinks.Roles {
				if role.Title != nil {
					roles = append(roles, *role.Title)
				}
			}
			result += fmt.Sprintf("- **%s** — %s\n", *m.UnderscoreLinks.Principal.Title, strings.Join(roles, ", "))
		}
	}
	return textResult(result), nil
}

func (r *Registry) listRoles(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := r.client.APIClient().ListRoles(ctx, nil)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list roles: %v", err), nil
	}

	// ListRoles returns an indeterminate model; read raw and try to parse
	var raw map[string]interface{}
	if err := openproject.ReadResponseRawTo(resp, &raw); err != nil {
		return errorResult("Failed to list roles: %v", err), nil
	}

	// Try to extract embedded elements
	result := "Available roles:\n\n"
	if embedded, ok := raw["_embedded"].(map[string]interface{}); ok {
		if elements, ok := embedded["elements"].([]interface{}); ok {
			for _, e := range elements {
				if elem, ok := e.(map[string]interface{}); ok {
					name, _ := elem["name"].(string)
					id, _ := elem["id"].(float64)
					result += fmt.Sprintf("- **%s** (ID: %d)\n", name, int(id))
				}
			}
		}
	}
	return textResult(result), nil
}

func (r *Registry) getRole(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetRoleArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().ViewRole(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get role: %v", err), nil
	}
	var role external.RoleModel
	if err := openproject.ReadResponse(resp, &role); err != nil {
		return errorResult("Failed to get role: %v", err), nil
	}

	result := fmt.Sprintf("# %s\n\n- **ID:** %d\n", role.Name, derefInt(role.Id))
	return textResult(result), nil
}
