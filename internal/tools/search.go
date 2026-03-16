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
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 10
	}

	switch args.Type {
	case "project":
		list, err := r.client.ListProjects(ctx, &openproject.ListProjectsOptions{
			PageSize: limit,
			Filters: []openproject.ProjectFilter{
				{Name: "name_and_identifier", Values: []string{args.Query}, Operator: "~"},
			},
		})
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Search failed: %v", err)}}}, nil
		}
		result := fmt.Sprintf("Found %d projects matching \"%s\":\n\n", list.Total, args.Query)
		for _, p := range list.Embedded.Elements {
			result += fmt.Sprintf("- **%s** (ID: %d) — %s\n", p.Name, p.ID, p.Identifier)
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil

	case "work_package":
		list, err := r.client.ListWorkPackages(ctx, &openproject.ListWorkPackagesOptions{
			PageSize: limit,
			Filters: []openproject.WorkPackageFilter{
				{Name: "subject", Values: []interface{}{args.Query}, Operator: "~"},
			},
		})
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Search failed: %v", err)}}}, nil
		}
		result := fmt.Sprintf("Found %d work packages matching \"%s\":\n\n", list.Total, args.Query)
		for _, wp := range list.Embedded.Elements {
			status := ""
			if wp.Links != nil && wp.Links.Status != nil {
				status = " — " + wp.Links.Status.Title
			}
			result += fmt.Sprintf("- **#%d %s**%s\n", wp.ID, wp.Subject, status)
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil

	case "user":
		list, err := r.client.ListUsers(ctx, &openproject.ListUsersOptions{
			PageSize: limit,
			Filters: []openproject.UserFilter{
				{Name: "name", Values: []string{args.Query}, Operator: "~"},
			},
		})
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Search failed: %v", err)}}}, nil
		}
		result := fmt.Sprintf("Found %d users matching \"%s\":\n\n", list.Total, args.Query)
		for _, u := range list.Embedded.Elements {
			result += fmt.Sprintf("- **%s** (ID: %d) — %s\n", u.Name, u.ID, u.Email)
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil

	default:
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Unknown type %q; must be project, work_package, or user", args.Type)}}}, nil
	}
}
