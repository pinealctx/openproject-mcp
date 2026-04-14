package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

type SearchArgs struct {
	Query string `json:"query"`
	Type  string `json:"type"`
	Limit int    `json:"limit,omitempty"`
}

// registerSearchTools registers the cross-resource search tool.
func (r *Registry) registerSearchTools(server *mcp.Server) {
	addTool(server, "search",
		"Search for projects, work packages, or users by keyword",
		newSchema(schemaProps{
			"query": schemaStr("Search keyword"),
			"type":  schemaEnum("Resource type to search", "project", "work_package", "user"),
			"limit": schemaInt("Maximum number of results (default: 10)"),
		}, "query", "type"),
		r.search)
}

func (r *Registry) search(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args SearchArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 10
	}

	switch args.Type {
	case "project":
		params := &external.ListProjectsParams{
			Filters: strPtr(fmt.Sprintf(`[{"name_and_identifier":{"operator":"~","values":["%s"]}}]`, args.Query)),
		}
		resp, err := r.client.APIClient().ListProjects(ctx, params)
		if err != nil {
			return errorResult("Search failed: %v", err), nil
		}
		var list external.ProjectCollectionModel
		if err := openproject.ReadResponse(resp, &list); err != nil {
			return errorResult("Search failed: %v", err), nil
		}
		result := fmt.Sprintf("Found %d projects matching \"%s\":\n\n", list.Total, args.Query)
		for _, p := range list.UnderscoreEmbedded.Elements {
			result += fmt.Sprintf("- **%s** (ID: %d) — %s\n", derefStr(p.Name), derefInt(p.Id), derefStr(p.Identifier))
		}
		return textResult(result), nil

	case "work_package":
		params := &external.ListWorkPackagesParams{
			PageSize: intPtr(limit),
			Filters:  strPtr(fmt.Sprintf(`[{"subject":{"operator":"~","values":["%s"]}}]`, args.Query)),
		}
		resp, err := r.client.APIClient().ListWorkPackages(ctx, params)
		if err != nil {
			return errorResult("Search failed: %v", err), nil
		}
		var list external.WorkPackagesModel
		if err := openproject.ReadResponse(resp, &list); err != nil {
			return errorResult("Search failed: %v", err), nil
		}
		result := fmt.Sprintf("Found %d work packages matching \"%s\":\n\n", list.Total, args.Query)
		for _, wp := range list.UnderscoreEmbedded.Elements {
			status := ""
			if wp.UnderscoreLinks.Status.Title != nil {
				status = " — " + *wp.UnderscoreLinks.Status.Title
			}
			result += fmt.Sprintf("- **#%d %s**%s\n", derefInt(wp.Id), wp.Subject, status)
		}
		return textResult(result), nil

	case "user":
		params := &external.ListUsersParams{
			PageSize: intPtr(limit),
			Filters:  strPtr(fmt.Sprintf(`[{"name":{"operator":"~","values":["%s"]}}]`, args.Query)),
		}
		resp, err := r.client.APIClient().ListUsers(ctx, params)
		if err != nil {
			return errorResult("Search failed: %v", err), nil
		}
		var list external.UserCollectionModel
		if err := openproject.ReadResponse(resp, &list); err != nil {
			return errorResult("Search failed: %v", err), nil
		}
		result := fmt.Sprintf("Found %d users matching \"%s\":\n\n", list.Total, args.Query)
		for _, u := range list.UnderscoreEmbedded.Elements {
			result += fmt.Sprintf("- **%s** (ID: %d) — %s\n", u.Name, u.Id, derefStr(u.Email))
		}
		return textResult(result), nil

	default:
		return errorResult("Unknown type %q; must be project, work_package, or user", args.Type), nil
	}
}
