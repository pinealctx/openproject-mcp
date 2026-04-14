package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListWikiPagesArgs struct {
	ProjectID int `json:"projectId"`
}

type GetWikiPageArgs struct {
	ID int `json:"id"`
}

type UpdateWikiPageArgs struct {
	ID          int    `json:"id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// wikiPageCollection is a minimal HAL collection for wiki pages.
type wikiPageCollection struct {
	Embedded struct {
		Elements []wikiPageElement `json:"elements"`
	} `json:"_embedded"`
	Count int `json:"count"`
	Total int `json:"total"`
}

// wikiPageElement represents a wiki page in a collection.
type wikiPageElement struct {
	ID    *int   `json:"id,omitempty"`
	Title string `json:"title"`
	Links *struct {
		Self        *wikiLink `json:"self,omitempty"`
		AddAttachment *wikiLink `json:"addAttachment,omitempty"`
	} `json:"_links,omitempty"`
}

// wikiPageDetail represents a full wiki page with content.
type wikiPageDetail struct {
	ID          *int              `json:"id,omitempty"`
	Title       string            `json:"title"`
	Description *wikiFormattable `json:"description,omitempty"`
	Links       *struct {
		Self         *wikiLink `json:"self,omitempty"`
		AddAttachment *wikiLink `json:"addAttachment,omitempty"`
	} `json:"_links,omitempty"`
}

type wikiLink struct {
	Href  string `json:"href"`
	Title string `json:"title,omitempty"`
}

type wikiFormattable struct {
	Format *string `json:"format,omitempty"`
	Raw    *string `json:"raw,omitempty"`
	Html   *string `json:"html,omitempty"`
}

// wikiPageWrite is the body for updating a wiki page.
type wikiPageWrite struct {
	Title       *string              `json:"title,omitempty"`
	Description *wikiFormattable     `json:"description,omitempty"`
}

func (r *Registry) registerWikiTools(server *mcp.Server) {
	addTool(server, "list_wiki_pages",
		"List wiki pages for a project",
		newSchema(schemaProps{
			"projectId": schemaInt("Project ID"),
		}, "projectId"),
		r.listWikiPages)

	addTool(server, "get_wiki_page",
		"Get details of a specific wiki page by ID",
		newSchema(schemaProps{
			"id": schemaInt("Wiki page ID"),
		}, "id"),
		r.getWikiPage)

	addTool(server, "update_wiki_page",
		"Update an existing wiki page",
		newSchema(schemaProps{
			"id":          schemaInt("Wiki page ID"),
			"title":       schemaStr("New title for the wiki page"),
			"description": schemaStr("New content in markdown"),
		}, "id"),
		r.updateWikiPage)
}

func (r *Registry) listWikiPages(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWikiPagesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	var list wikiPageCollection
	path := fmt.Sprintf("/projects/%d/wiki_pages", args.ProjectID)
	if err := r.client.Get(ctx, path, &list); err != nil {
		return errorResult("Failed to list wiki pages: %v", err), nil
	}

	if list.Total == 0 {
		return textResult(fmt.Sprintf("No wiki pages found for project #%d.", args.ProjectID)), nil
	}

	result := fmt.Sprintf("Found %d wiki pages for project #%d:\n\n", list.Total, args.ProjectID)
	for _, p := range list.Embedded.Elements {
		href := ""
		if p.Links != nil && p.Links.Self != nil {
			href = p.Links.Self.Href
		}
		result += fmt.Sprintf("- **%s** (ID: %d) — %s\n", p.Title, derefInt(p.ID), href)
	}
	return textResult(result), nil
}

func (r *Registry) getWikiPage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetWikiPageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	var page wikiPageDetail
	path := fmt.Sprintf("/wiki_pages/%d", args.ID)
	if err := r.client.Get(ctx, path, &page); err != nil {
		return errorResult("Failed to get wiki page: %v", err), nil
	}

	result := fmt.Sprintf("# %s\n\n", page.Title)
	result += fmt.Sprintf("- **ID:** %d\n", derefInt(page.ID))
	if page.Description != nil && page.Description.Raw != nil && *page.Description.Raw != "" {
		result += fmt.Sprintf("\n## Content\n%s\n", *page.Description.Raw)
	}
	return textResult(result), nil
}

func (r *Registry) updateWikiPage(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateWikiPageArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := wikiPageWrite{}
	if args.Title != "" {
		body.Title = strPtr(args.Title)
	}
	if args.Description != "" {
		format := "markdown"
		body.Description = &wikiFormattable{
			Format: &format,
			Raw:    strPtr(args.Description),
		}
	}

	var page wikiPageDetail
	path := fmt.Sprintf("/wiki_pages/%d", args.ID)
	if err := r.client.Patch(ctx, path, body, &page); err != nil {
		return errorResult("Failed to update wiki page: %v", err), nil
	}

	return textResult(fmt.Sprintf("Wiki page #%d updated. Title: %s", args.ID, page.Title)), nil
}
