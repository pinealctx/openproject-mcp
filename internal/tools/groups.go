package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

type ListGroupsArgs struct {
	SortBy string `json:"sortBy,omitempty"`
}

type GetGroupArgs struct {
	ID int `json:"id"`
}

type CreateGroupArgs struct {
	Name string `json:"name"`
}

type UpdateGroupArgs struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

type DeleteGroupArgs struct {
	ID int `json:"id"`
}

func (r *Registry) registerGroupTools(server *mcp.Server) {
	addTool(server, "list_groups",
		"List all groups in OpenProject",
		newSchema(schemaProps{
			"sortBy": schemaStr(`Sort criteria, e.g. "created_at:asc"`),
		}),
		r.listGroups)

	addTool(server, "get_group",
		"Get details of a specific group by ID",
		newSchema(schemaProps{
			"id": schemaInt("Group ID"),
		}, "id"),
		r.getGroup)

	addTool(server, "create_group",
		"Create a new group",
		newSchema(schemaProps{
			"name": schemaStr("Group name"),
		}, "name"),
		r.createGroup)

	addTool(server, "update_group",
		"Update an existing group",
		newSchema(schemaProps{
			"id":   schemaInt("Group ID"),
			"name": schemaStr("New group name"),
		}, "id"),
		r.updateGroup)

	addTool(server, "delete_group",
		"Delete a group",
		newSchema(schemaProps{
			"id": schemaInt("Group ID"),
		}, "id"),
		r.deleteGroup)
}

func (r *Registry) listGroups(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListGroupsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListGroupsParams{}
	if args.SortBy != "" {
		params.SortBy = strPtr(normalizeSortBy(args.SortBy))
	}

	resp, err := r.client.APIClient().ListGroups(ctx, params)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list groups: %v", err), nil
	}
	var list external.GroupCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list groups: %v", err), nil
	}

	if list.Count == 0 {
		return textResult("No groups found."), nil
	}

	result := fmt.Sprintf("Found %d groups:\n\n", list.Count)
	for _, g := range list.UnderscoreEmbedded.Elements {
		memberCount := 0
		if g.UnderscoreEmbedded.Members != nil {
			memberCount = len(*g.UnderscoreEmbedded.Members)
		}
		result += fmt.Sprintf("- **%s** (ID: %d) — %d members\n", g.Name, g.Id, memberCount)
	}
	return textResult(result), nil
}

func (r *Registry) getGroup(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetGroupArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().GetGroup(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get group: %v", err), nil
	}
	var group external.GroupModel
	if err := openproject.ReadResponse(resp, &group); err != nil {
		return errorResult("Failed to get group: %v", err), nil
	}

	result := fmt.Sprintf("# %s\n\n", group.Name)
	result += fmt.Sprintf("- **ID:** %d\n", group.Id)
	if group.CreatedAt != nil {
		result += fmt.Sprintf("- **Created At:** %s\n", group.CreatedAt.Format("2006-01-02 15:04"))
	}
	if group.UpdatedAt != nil {
		result += fmt.Sprintf("- **Updated At:** %s\n", group.UpdatedAt.Format("2006-01-02 15:04"))
	}
	if group.UnderscoreEmbedded.Members != nil {
		result += fmt.Sprintf("\n## Members (%d)\n\n", len(*group.UnderscoreEmbedded.Members))
		for _, m := range *group.UnderscoreEmbedded.Members {
			result += fmt.Sprintf("- **%s** (ID: %d)\n", m.Name, m.Id)
		}
	}
	return textResult(result), nil
}

func (r *Registry) createGroup(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateGroupArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.GroupWriteModel{
		Name: strPtr(args.Name),
	}

	resp, err := r.client.APIClient().CreateGroup(ctx, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to create group: %v", err), nil
	}
	var group external.GroupModel
	if err := openproject.ReadResponse(resp, &group); err != nil {
		return errorResult("Failed to create group: %v", err), nil
	}

	return textResult(fmt.Sprintf("Group **%s** created (ID: %d).", group.Name, group.Id)), nil
}

func (r *Registry) updateGroup(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateGroupArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.GroupWriteModel{}
	if args.Name != "" {
		body.Name = strPtr(args.Name)
	}

	resp, err := r.client.APIClient().UpdateGroup(ctx, args.ID, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to update group: %v", err), nil
	}
	var group external.GroupModel
	if err := openproject.ReadResponse(resp, &group); err != nil {
		return errorResult("Failed to update group: %v", err), nil
	}

	return textResult(fmt.Sprintf("Group #%d updated. Name: %s", args.ID, group.Name)), nil
}

func (r *Registry) deleteGroup(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteGroupArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().DeleteGroup(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to delete group: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to delete group: %v", err), nil
	}

	return textResult(fmt.Sprintf("Group #%d deleted.", args.ID)), nil
}
