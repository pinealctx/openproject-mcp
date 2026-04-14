package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

type ListPlaceholderUsersArgs struct {
	Filters string `json:"filters,omitempty"`
}

type GetPlaceholderUserArgs struct {
	ID string `json:"id"`
}

type CreatePlaceholderUserArgs struct {
	Name string `json:"name"`
}

// placeholderUserCollection is a minimal HAL collection for placeholder users.
type placeholderUserCollection struct {
	Embedded struct {
		Elements []external.PlaceholderUserModel `json:"elements"`
	} `json:"_embedded"`
	Count int `json:"count"`
	Total int `json:"total"`
}

func (r *Registry) registerPlaceholderTools(server *mcp.Server) {
	addTool(server, "list_placeholder_users",
		"List all placeholder users in OpenProject",
		newSchema(schemaProps{
			"filters": schemaStr(`JSON filter string, e.g. [{"name":{"operator":"~","values":["test"]}}]`),
		}),
		r.listPlaceholderUsers)

	addTool(server, "get_placeholder_user",
		"Get details of a specific placeholder user by ID",
		newSchema(schemaProps{
			"id": schemaStr("Placeholder user ID"),
		}, "id"),
		r.getPlaceholderUser)

	addTool(server, "create_placeholder_user",
		"Create a new placeholder user",
		newSchema(schemaProps{
			"name": schemaStr("Name for the placeholder user"),
		}, "name"),
		r.createPlaceholderUser)
}

func (r *Registry) listPlaceholderUsers(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListPlaceholderUsersArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListPlaceholderUsersParams{}
	if args.Filters != "" {
		params.Filters = strPtr(args.Filters)
	}

	resp, err := r.client.APIClient().ListPlaceholderUsers(ctx, params)
	if err != nil {
		return errorResult("Failed to list placeholder users: %v", err), nil
	}
	var list placeholderUserCollection
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list placeholder users: %v", err), nil
	}

	if list.Total == 0 {
		return textResult("No placeholder users found."), nil
	}

	result := fmt.Sprintf("Found %d placeholder users:\n\n", list.Total)
	for _, p := range list.Embedded.Elements {
		createdAt := ""
		if p.CreatedAt != nil {
			createdAt = p.CreatedAt.Format("2006-01-02")
		}
		result += fmt.Sprintf("- **%s** (ID: %d) — Created: %s\n", p.Name, p.Id, createdAt)
	}
	return textResult(result), nil
}

func (r *Registry) getPlaceholderUser(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetPlaceholderUserArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().ViewPlaceholderUser(ctx, args.ID)
	if err != nil {
		return errorResult("Failed to get placeholder user: %v", err), nil
	}
	var p external.PlaceholderUserModel
	if err := openproject.ReadResponse(resp, &p); err != nil {
		return errorResult("Failed to get placeholder user: %v", err), nil
	}

	result := fmt.Sprintf("# %s\n\n", p.Name)
	result += fmt.Sprintf("- **ID:** %d\n", p.Id)
	result += fmt.Sprintf("- **Type:** %s\n", p.UnderscoreType)
	if p.CreatedAt != nil {
		result += fmt.Sprintf("- **Created At:** %s\n", p.CreatedAt.Format("2006-01-02 15:04"))
	}
	return textResult(result), nil
}

func (r *Registry) createPlaceholderUser(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreatePlaceholderUserArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.PlaceholderUserCreateModel{
		Name: strPtr(args.Name),
	}

	resp, err := r.client.APIClient().CreatePlaceholderUser(ctx, body)
	if err != nil {
		return errorResult("Failed to create placeholder user: %v", err), nil
	}
	var p external.PlaceholderUserModel
	if err := openproject.ReadResponse(resp, &p); err != nil {
		return errorResult("Failed to create placeholder user: %v", err), nil
	}

	return textResult(fmt.Sprintf("Placeholder user **%s** created (ID: %d).", p.Name, p.Id)), nil
}
