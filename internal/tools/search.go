package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type SearchArgs struct {
	Query string `json:"query"`
	Type  string `json:"type"`
	Limit int    `json:"limit,omitempty"`
}

// registerSearchTools registers the cross-resource search tool.
func (r *Registry) registerSearchTools(server *mcp.Server) {
	addTool(server, "search",
		"Search projects, work packages, or users by keyword using supported OpenProject collection filters",
		newSchema(schemaProps{
			"query": schemaStr("Search keyword"),
			"type":  schemaEnum("Optional resource type; omit to search all accessible types", "project", "work_package", "user"),
			"limit": schemaInt("Maximum results per resource type (default: 10, maximum: 100)"),
		}, "query"),
		r.search)
}

func (r *Registry) search(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args SearchArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	results, err := r.client.Search(ctx, openproject.SearchInput{
		Query: args.Query,
		Type:  args.Type,
		Limit: args.Limit,
	})
	if err != nil {
		return errorResult("Search failed: %v", err), nil
	}
	return textResult(formatSearchResults(results)), nil
}

func formatSearchResults(results *openproject.SearchResults) string {
	text := fmt.Sprintf("Found %d of %d matching resources for %q:\n\n", results.Count, results.Total, results.Query)
	for _, result := range results.Results {
		detail := result.Identifier
		if result.Status != "" {
			detail = result.Status
		}
		if detail != "" {
			detail = " - " + detail
		}
		text += fmt.Sprintf("- **[%s] #%d %s**%s\n", result.Type, result.ID, result.Title, detail)
	}
	for _, warning := range results.Warnings {
		text += fmt.Sprintf("\nWarning: %s\n", warning.Message)
	}
	return text
}
