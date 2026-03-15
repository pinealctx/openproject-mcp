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
	server.AddTool(&mcp.Tool{Name: "set_work_package_parent", Description: "Set the parent of a work package"}, r.setWorkPackageParent)
	server.AddTool(&mcp.Tool{Name: "remove_work_package_parent", Description: "Remove the parent relationship from a work package"}, r.removeWorkPackageParent)
	server.AddTool(&mcp.Tool{Name: "list_work_package_children", Description: "List all child work packages of a parent"}, r.listWorkPackageChildren)
	server.AddTool(&mcp.Tool{Name: "create_work_package_relation", Description: "Create a relation between two work packages"}, r.createWorkPackageRelation)
	server.AddTool(&mcp.Tool{Name: "list_work_package_relations", Description: "List all relations for a work package"}, r.listWorkPackageRelations)
	server.AddTool(&mcp.Tool{Name: "get_work_package_relation", Description: "Get details of a specific relation"}, r.getWorkPackageRelation)
	server.AddTool(&mcp.Tool{Name: "update_work_package_relation", Description: "Update an existing relation"}, r.updateWorkPackageRelation)
	server.AddTool(&mcp.Tool{Name: "delete_work_package_relation", Description: "Delete a relation between work packages"}, r.deleteWorkPackageRelation)
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
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Relation #%d created: %s (#%d -> #%d)", rel.ID, args.Type, args.FromID, args.ToID)}}}, nil
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
