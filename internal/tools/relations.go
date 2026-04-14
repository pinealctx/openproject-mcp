package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
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
		return errorResult("Invalid arguments: %v", err), nil
	}

	// Need to fetch current WP to get lockVersion
	resp, err := r.client.APIClient().ViewWorkPackage(ctx, args.WorkPackageID, nil)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to fetch work package: %v", err), nil
	}
	var current external.WorkPackageModel
	if err := openproject.ReadResponse(resp, &current); err != nil {
		return errorResult("Failed to fetch work package: %v", err), nil
	}

	lockVersion := 0
	if current.LockVersion != nil {
		lockVersion = *current.LockVersion
	}

	body := external.WorkPackagePatchModel{
		LockVersion: lockVersion,
		UnderscoreLinks: &struct {
			Assignee    *external.Link `json:"assignee,omitempty"`
			Category    *external.Link `json:"category,omitempty"`
			Parent      *external.Link `json:"parent,omitempty"`
			Priority    *external.Link `json:"priority,omitempty"`
			Project     *external.Link `json:"project,omitempty"`
			Responsible *external.Link `json:"responsible,omitempty"`
			Status      *external.Link `json:"status,omitempty"`
			Type        *external.Link `json:"type,omitempty"`
			Version     *external.Link `json:"version,omitempty"`
		}{
			Parent: &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/work_packages/%d", args.ParentID))},
		},
	}

	resp, err = r.client.APIClient().UpdateWorkPackage(ctx, args.WorkPackageID, nil, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to set parent: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to set parent: %v", err), nil
	}
	return textResult(fmt.Sprintf("Work package #%d parent set to #%d", args.WorkPackageID, args.ParentID)), nil
}

func (r *Registry) removeWorkPackageParent(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args RemoveWorkPackageParentArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// Need to fetch current WP to get lockVersion
	resp, err := r.client.APIClient().ViewWorkPackage(ctx, args.WorkPackageID, nil)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to fetch work package: %v", err), nil
	}
	var current external.WorkPackageModel
	if err := openproject.ReadResponse(resp, &current); err != nil {
		return errorResult("Failed to fetch work package: %v", err), nil
	}

	lockVersion := 0
	if current.LockVersion != nil {
		lockVersion = *current.LockVersion
	}

	body := external.WorkPackagePatchModel{
		LockVersion: lockVersion,
		UnderscoreLinks: &struct {
			Assignee    *external.Link `json:"assignee,omitempty"`
			Category    *external.Link `json:"category,omitempty"`
			Parent      *external.Link `json:"parent,omitempty"`
			Priority    *external.Link `json:"priority,omitempty"`
			Project     *external.Link `json:"project,omitempty"`
			Responsible *external.Link `json:"responsible,omitempty"`
			Status      *external.Link `json:"status,omitempty"`
			Type        *external.Link `json:"type,omitempty"`
			Version     *external.Link `json:"version,omitempty"`
		}{
			Parent: &external.Link{},
		},
	}

	resp, err = r.client.APIClient().UpdateWorkPackage(ctx, args.WorkPackageID, nil, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to remove parent: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to remove parent: %v", err), nil
	}
	return textResult(fmt.Sprintf("Parent removed from work package #%d", args.WorkPackageID)), nil
}

func (r *Registry) listWorkPackageChildren(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackageChildrenArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// Use filters on work packages list with parent filter
	params := &external.ListWorkPackagesParams{
		Filters: strPtr(fmt.Sprintf(`[{"parent":{"operator":"=","values":["%d"]}}]`, args.WorkPackageID)),
	}
	resp, err := r.client.APIClient().ListWorkPackages(ctx, params)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list children: %v", err), nil
	}

	var list external.WorkPackagesModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list children: %v", err), nil
	}

	result := fmt.Sprintf("Found %d children:\n\n", list.Total)
	for _, wp := range list.UnderscoreEmbedded.Elements {
		result += fmt.Sprintf("- #%d %s\n", derefInt(wp.Id), wp.Subject)
	}
	return textResult(result), nil
}

func (r *Registry) createWorkPackageRelation(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateWorkPackageRelationArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.RelationWriteModel{
		Type: external.RelationWriteModelType(args.Type),
		UnderscoreLinks: struct {
			To *external.Link `json:"to,omitempty"`
		}{
			To: &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/work_packages/%d", args.ToID))},
		},
	}
	if args.Description != "" {
		body.Description = strPtr(args.Description)
	}
	if args.Delay > 0 {
		body.Lag = intPtr(args.Delay)
	}

	resp, err := r.client.APIClient().CreateRelation(ctx, args.FromID, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to create relation: %v", err), nil
	}
	var rel external.RelationReadModel
	if err := openproject.ReadResponse(resp, &rel); err != nil {
		return errorResult("Failed to create relation: %v", err), nil
	}
	return textResult(fmt.Sprintf("Relation #%d created: %s (#%d → #%d)", derefInt(rel.Id), args.Type, args.FromID, args.ToID)), nil
}

func (r *Registry) listWorkPackageRelations(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackageRelationsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListRelationsParams{
		Filters: strPtr(fmt.Sprintf(`[{"involved":{"operator":"=","values":["%d"]}}]`, args.WorkPackageID)),
	}
	resp, err := r.client.APIClient().ListRelations(ctx, params)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list relations: %v", err), nil
	}

	var list external.RelationCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list relations: %v", err), nil
	}

	result := fmt.Sprintf("Found %d relations:\n\n", list.Total)
	for _, rel := range list.UnderscoreEmbedded.Elements {
		result += fmt.Sprintf("- #%d: %s\n", derefInt(rel.Id), string(derefStr((*string)(rel.Type))))
	}
	return textResult(result), nil
}

func (r *Registry) getWorkPackageRelation(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetWorkPackageRelationArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().GetRelation(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get relation: %v", err), nil
	}
	var rel external.RelationReadModel
	if err := openproject.ReadResponse(resp, &rel); err != nil {
		return errorResult("Failed to get relation: %v", err), nil
	}
	return textResult(fmt.Sprintf("Relation #%d: %s", derefInt(rel.Id), string(derefStr((*string)(rel.Type))))), nil
}

func (r *Registry) updateWorkPackageRelation(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateWorkPackageRelationArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.RelationWriteModel{}
	if args.Type != "" {
		body.Type = external.RelationWriteModelType(args.Type)
	}
	if args.Description != "" {
		body.Description = strPtr(args.Description)
	}
	if args.Delay > 0 {
		body.Lag = intPtr(args.Delay)
	}

	resp, err := r.client.APIClient().UpdateRelation(ctx, args.ID, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to update relation: %v", err), nil
	}
	var rel external.RelationReadModel
	if err := openproject.ReadResponse(resp, &rel); err != nil {
		return errorResult("Failed to update relation: %v", err), nil
	}
	return textResult(fmt.Sprintf("Relation #%d updated successfully!", derefInt(rel.Id))), nil
}

func (r *Registry) deleteWorkPackageRelation(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteWorkPackageRelationArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().DeleteRelation(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to delete relation: %v", err), nil
	}
	if err := openproject.ReadResponse(resp, nil); err != nil {
		return errorResult("Failed to delete relation: %v", err), nil
	}
	return textResult(fmt.Sprintf("Relation #%d deleted successfully!", args.ID)), nil
}
