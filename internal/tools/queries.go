package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

type ListQueriesArgs struct {
	Filters string `json:"filters,omitempty"`
}

type GetQueryArgs struct {
	ID int `json:"id"`
}

// queryCollection is a minimal HAL collection for queries.
type queryCollection struct {
	Embedded struct {
		Elements []external.QueryModel `json:"elements"`
	} `json:"_embedded"`
	Count int `json:"count"`
	Total int `json:"total"`
}

func (r *Registry) registerQueryTools(server *mcp.Server) {
	addTool(server, "list_queries",
		"List queries in OpenProject",
		newSchema(schemaProps{
			"filters": schemaStr(`JSON filter string, e.g. [{"project":{"operator":"=","values":["1"]}}]`),
		}),
		r.listQueries)

	addTool(server, "get_query",
		"Get details of a specific query by ID",
		newSchema(schemaProps{
			"id": schemaInt("Query ID"),
		}, "id"),
		r.getQuery)
}

func (r *Registry) listQueries(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListQueriesArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListQueriesParams{}
	if args.Filters != "" {
		params.Filters = strPtr(args.Filters)
	}

	resp, err := r.client.APIClient().ListQueries(ctx, params)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list queries: %v", err), nil
	}
	var list queryCollection
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list queries: %v", err), nil
	}

	if list.Total == 0 {
		return textResult("No queries found."), nil
	}

	result := fmt.Sprintf("Found %d queries:\n\n", list.Total)
	for _, q := range list.Embedded.Elements {
		name := derefStr(q.Name)
		id := derefInt(q.Id)
		public := ""
		if q.Public != nil && *q.Public {
			public = " [public]"
		}
		starred := ""
		if q.Starred != nil && *q.Starred {
			starred = " *"
		}
		result += fmt.Sprintf("- **#%d %s**%s%s\n", id, name, public, starred)
	}
	return textResult(result), nil
}

func (r *Registry) getQuery(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetQueryArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().ViewQuery(ctx, args.ID, nil)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get query: %v", err), nil
	}
	var q external.QueryModel
	if err := openproject.ReadResponse(resp, &q); err != nil {
		return errorResult("Failed to get query: %v", err), nil
	}

	result := fmt.Sprintf("# %s\n\n", derefStr(q.Name))
	result += fmt.Sprintf("- **ID:** %d\n", derefInt(q.Id))
	result += fmt.Sprintf("- **Public:** %v\n", derefBool(q.Public))
	result += fmt.Sprintf("- **Starred:** %v\n", derefBool(q.Starred))
	result += fmt.Sprintf("- **Hidden:** %v\n", derefBool(q.Hidden))
	result += fmt.Sprintf("- **Sums:** %v\n", derefBool(q.Sums))
	if q.CreatedAt != nil {
		result += fmt.Sprintf("- **Created At:** %s\n", q.CreatedAt.Format("2006-01-02 15:04"))
	}
	if q.UpdatedAt != nil {
		result += fmt.Sprintf("- **Updated At:** %s\n", q.UpdatedAt.Format("2006-01-02 15:04"))
	}
	if q.Filters != nil && len(*q.Filters) > 0 {
		result += fmt.Sprintf("\n## Filters (%d)\n\n", len(*q.Filters))
		for _, f := range *q.Filters {
			result += fmt.Sprintf("- %s\n", f.Name)
		}
	}
	return textResult(result), nil
}
