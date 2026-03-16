package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type SetWorkPackageParentArgs struct {
	WorkPackageID int `json:"workPackageId"`
	ParentID      int `json:"parentId"`
}
type RemoveWorkPackageParentArgs struct{ WorkPackageID int }
type ListWorkPackageChildrenArgs struct{ WorkPackageID int }
type CreateWorkPackageRelationArgs struct {
	FromID      int    `json:"fromId"`
	ToID        int    `json:"toId"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Delay       int    `json:"delay,omitempty"`
}
type ListWorkPackageRelationsArgs struct{ WorkPackageID int }
type GetWorkPackageRelationArgs struct{ ID int }
type UpdateWorkPackageRelationArgs struct {
	ID          int    `json:"id"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Delay       int    `json:"delay,omitempty"`
}
type DeleteWorkPackageRelationArgs struct{ ID int }

// registerRelationTools registers relation-related tools.
func (r *Registry) registerRelationTools(server *mcp.Server) {
	addTool(server, "set_work_package_parent", "Set the parent of a work package",
		newSchema(schemaProps{
			"workPackageId": schemaInt("Child work package ID"),
			"parentId":      schemaInt("Parent work package ID"),
		}, "workPackageId", "parentId"),
		r.setWorkPackageParent)

	addTool(server, "remove_work_package_parent", "Remove the parent relationship from a work package",
		newSchema(schemaProps{"workPackageId": schemaInt("Work package ID")}, "workPackageId"),
		r.removeWorkPackageParent)

	addTool(server, "list_work_package_children", "List all child work packages of a parent",
		newSchema(schemaProps{"workPackageId": schemaInt("Parent work package ID")}, "workPackageId"),
		r.listWorkPackageChildren)

	addTool(server, "create_work_package_relation", "Create a relation between two work packages",
		newSchema(schemaProps{
			"fromId":      schemaInt("Source work package ID"),
			"toId":        schemaInt("Target work package ID"),
			"type":        schemaEnum("Relation type", "follows", "precedes", "blocks", "blocked_by", "includes", "part_of", "duplicates", "duplicated"),
			"description": schemaStr("Optional description"),
			"delay":       schemaInt("Delay in days (for follows/precedes)"),
		}, "fromId", "toId", "type"),
		r.createWorkPackageRelation)

	addTool(server, "list_work_package_relations", "List all relations for a work package",
		newSchema(schemaProps{"workPackageId": schemaInt("Work package ID")}, "workPackageId"),
		r.listWorkPackageRelations)

	addTool(server, "get_work_package_relation", "Get details of a specific relation",
		newSchema(schemaProps{"id": schemaInt("Relation ID")}, "id"),
		r.getWorkPackageRelation)

	addTool(server, "update_work_package_relation", "Update an existing relation",
		newSchema(schemaProps{
			"id":          schemaInt("Relation ID"),
			"type":        schemaEnum("New relation type", "follows", "precedes", "blocks", "blocked_by", "includes", "part_of", "duplicates", "duplicated"),
			"description": schemaStr("New description"),
			"delay":       schemaInt("New delay in days"),
		}, "id"),
		r.updateWorkPackageRelation)

	addTool(server, "delete_work_package_relation", "Delete a relation between work packages",
		newSchema(schemaProps{"id": schemaInt("Relation ID")}, "id"),
		r.deleteWorkPackageRelation)
}

func (r *Registry) setWorkPackageParent(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args SetWorkPackageParentArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	_, err := r.client.SetWorkPackageParent(ctx, args.WorkPackageID, args.ParentID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to set parent: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Work package #%d parent set to #%d", args.WorkPackageID, args.ParentID)}}}, nil
}

func (r *Registry) removeWorkPackageParent(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args RemoveWorkPackageParentArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	_, err := r.client.RemoveWorkPackageParent(ctx, args.WorkPackageID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to remove parent: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Parent removed from work package #%d", args.WorkPackageID)}}}, nil
}

func (r *Registry) listWorkPackageChildren(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackageChildrenArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	list, err := r.client.ListWorkPackageChildren(ctx, args.WorkPackageID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list children: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d children:\n\n", list.Total)
	for _, wp := range list.Embedded.Elements {
		result += fmt.Sprintf("- #%d %s\n", wp.ID, wp.Subject)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) createWorkPackageRelation(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateWorkPackageRelationArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.CreateRelationOptions{FromID: args.FromID, ToID: args.ToID, Type: args.Type, Description: args.Description, Delay: args.Delay}
	rel, err := r.client.CreateRelation(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create relation: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Relation #%d created: %s (#%d → #%d)", rel.ID, args.Type, args.FromID, args.ToID)}}}, nil
}

func (r *Registry) listWorkPackageRelations(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackageRelationsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	list, err := r.client.ListWorkPackageRelations(ctx, args.WorkPackageID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list relations: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d relations:\n\n", list.Total)
	for _, rel := range list.Embedded.Elements {
		result += fmt.Sprintf("- #%d: %s\n", rel.ID, rel.Type)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) getWorkPackageRelation(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetWorkPackageRelationArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	rel, err := r.client.GetRelation(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get relation: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Relation #%d: %s", rel.ID, rel.Type)}}}, nil
}

func (r *Registry) updateWorkPackageRelation(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateWorkPackageRelationArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.UpdateRelationOptions{Type: args.Type, Description: args.Description, Delay: args.Delay}
	rel, err := r.client.UpdateRelation(ctx, args.ID, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update relation: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Relation #%d updated successfully!", rel.ID)}}}, nil
}

func (r *Registry) deleteWorkPackageRelation(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteWorkPackageRelationArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	if err := r.client.DeleteRelation(ctx, args.ID); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete relation: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Relation #%d deleted successfully!", args.ID)}}}, nil
}
