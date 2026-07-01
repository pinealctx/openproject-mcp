package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type ListDocumentsArgs struct {
	Offset   int    `json:"offset,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	SortBy   string `json:"sortBy,omitempty"`
}

type GetDocumentArgs struct {
	ID int `json:"id"`
}

type UpdateDocumentArgs struct {
	ID          int    `json:"id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func (r *Registry) registerDocumentTools(server *mcp.Server) {
	addTool(server, "list_documents",
		"List documents in OpenProject",
		newSchema(schemaProps{
			"offset":   schemaInt("Pagination offset"),
			"pageSize": schemaInt("Number of items per page"),
			"sortBy":   schemaStr(`Sort criteria, e.g. "created_at:desc"`),
		}),
		r.listDocuments)

	addTool(server, "get_document",
		"Get details of a specific document",
		newSchema(schemaProps{
			"id": schemaInt("Document ID"),
		}, "id"),
		r.getDocument)

	addTool(server, "update_document",
		"Update an existing document",
		newSchema(schemaProps{
			"id":          schemaInt("Document ID"),
			"title":       schemaStr("New title for the document"),
			"description": schemaStr("New description in markdown"),
		}, "id"),
		r.updateDocument)
}

func (r *Registry) listDocuments(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListDocumentsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	list, err := r.client.ListDocuments(ctx, openproject.DocumentListInput{
		Offset:   args.Offset,
		PageSize: args.PageSize,
		SortBy:   normalizeSortBy(args.SortBy),
	})
	if err != nil {
		return errorResult("Failed to list documents: %v", err), nil
	}

	if list.Total == 0 {
		return textResult("No documents found."), nil
	}

	result := fmt.Sprintf("Found %d documents:\n\n", list.Total)
	for _, d := range list.Embedded.Elements {
		title := derefStr(d.Title)
		id := derefInt(d.Id)
		createdAt := ""
		if d.CreatedAt != nil {
			createdAt = d.CreatedAt.Format("2006-01-02")
		}
		project := ""
		if d.UnderscoreLinks != nil {
			project = derefStr(d.UnderscoreLinks.Project.Title)
		}
		result += fmt.Sprintf("- **#%d %s** — Project: %s, Created: %s\n", id, title, project, createdAt)
	}
	return textResult(result), nil
}

func (r *Registry) getDocument(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetDocumentArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	doc, err := r.client.GetDocument(ctx, args.ID)
	if err != nil {
		return errorResult("Failed to get document: %v", err), nil
	}

	result := fmt.Sprintf("# %s\n\n", derefStr(doc.Title))
	result += fmt.Sprintf("- **ID:** %d\n", derefInt(doc.Id))
	if doc.CreatedAt != nil {
		result += fmt.Sprintf("- **Created At:** %s\n", doc.CreatedAt.Format("2006-01-02 15:04"))
	}
	if doc.UnderscoreLinks != nil {
		result += fmt.Sprintf("- **Project:** %s\n", derefStr(doc.UnderscoreLinks.Project.Title))
	}
	if doc.Description != nil && doc.Description.Raw != nil && *doc.Description.Raw != "" {
		result += fmt.Sprintf("\n## Description\n%s\n", *doc.Description.Raw)
	}
	return textResult(result), nil
}

func (r *Registry) updateDocument(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateDocumentArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	doc, err := r.client.UpdateDocument(ctx, openproject.DocumentUpdateInput{
		ID:          args.ID,
		Title:       args.Title,
		Description: args.Description,
	})
	if err != nil {
		return errorResult("Failed to update document: %v", err), nil
	}

	return textResult(fmt.Sprintf("Document #%d updated. Title: %s", args.ID, derefStr(doc.Title))), nil
}
