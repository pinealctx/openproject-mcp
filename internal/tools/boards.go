package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

type GetBoardsArgs struct {
	ProjectID int `json:"projectId,omitempty"`
	Offset    int `json:"offset,omitempty"`
	PageSize  int `json:"pageSize,omitempty"`
}

type GetBoardArgs struct {
	ID int `json:"id"`
}

type CreateBoardArgs struct {
	ProjectID   int `json:"projectId"`
	RowCount    int `json:"rowCount,omitempty"`
	ColumnCount int `json:"columnCount,omitempty"`
}

type UpdateBoardArgs struct {
	ID          int `json:"id"`
	RowCount    int `json:"rowCount,omitempty"`
	ColumnCount int `json:"columnCount,omitempty"`
}

type DeleteBoardArgs struct {
	ID int `json:"id"`
}

type AddBoardWidgetArgs struct {
	BoardID     int    `json:"boardId"`
	Identifier  string `json:"identifier"`
	StartRow    int    `json:"startRow"`
	EndRow      int    `json:"endRow"`
	StartColumn int    `json:"startColumn"`
	EndColumn   int    `json:"endColumn"`
	QueryID     int    `json:"queryId,omitempty"`
}

type RemoveBoardWidgetArgs struct {
	BoardID  int `json:"boardId"`
	WidgetID int `json:"widgetId"`
}

// registerBoardTools registers board (grid) tools.
func (r *Registry) registerBoardTools(server *mcp.Server) {
	addTool(server, "get_boards", "List Kanban boards, optionally filtered by project",
		newSchema(schemaProps{
			"projectId": schemaInt("Filter boards by project ID"),
			"offset":    schemaInt("Pagination offset"),
			"pageSize":  schemaInt("Items per page"),
		}),
		r.getBoards)

	addTool(server, "get_board", "Get a specific Kanban board by ID",
		newSchema(schemaProps{"id": schemaInt("Board ID")}, "id"),
		r.getBoard)

	addTool(server, "create_board", "Create a new Kanban board in a project",
		newSchema(schemaProps{
			"projectId":   schemaInt("Project ID"),
			"rowCount":    schemaInt("Number of rows (default: 1)"),
			"columnCount": schemaInt("Number of columns (default: 3)"),
		}, "projectId"),
		r.createBoard)

	addTool(server, "update_board", "Update an existing Kanban board",
		newSchema(schemaProps{
			"id":          schemaInt("Board ID"),
			"rowCount":    schemaInt("New number of rows"),
			"columnCount": schemaInt("New number of columns"),
		}, "id"),
		r.updateBoard)

	addTool(server, "delete_board", "Delete a Kanban board",
		newSchema(schemaProps{"id": schemaInt("Board ID")}, "id"),
		r.deleteBoard)

	addTool(server, "add_board_widget", "Add a widget (column) to a Kanban board",
		newSchema(schemaProps{
			"boardId":     schemaInt("Board ID"),
			"identifier":  schemaStr(`Widget identifier, e.g. "work_package_query"`),
			"startRow":    schemaInt("Starting row position (1-based)"),
			"endRow":      schemaInt("Ending row position (exclusive)"),
			"startColumn": schemaInt("Starting column position (1-based)"),
			"endColumn":   schemaInt("Ending column position (exclusive)"),
			"queryId":     schemaInt("Query ID for filtering work packages"),
		}, "boardId", "identifier", "startRow", "endRow", "startColumn", "endColumn"),
		r.addBoardWidget)

	addTool(server, "remove_board_widget", "Remove a widget from a Kanban board",
		newSchema(schemaProps{
			"boardId":  schemaInt("Board ID"),
			"widgetId": schemaInt("Widget ID to remove"),
		}, "boardId", "widgetId"),
		r.removeBoardWidget)
}

func (r *Registry) getBoards(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetBoardsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	var filters string
	if args.ProjectID > 0 {
		filters = fmt.Sprintf(`[{"scope":{"operator":"=","values":["/api/v3/projects/%d"]}}]`, args.ProjectID)
	}
	list, err := r.client.ListGrids(ctx, &openproject.ListGridsOptions{
		Offset:   args.Offset,
		PageSize: args.PageSize,
		Filters:  filters,
	})
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list boards: %v", err)}}}, nil
	}

	result := fmt.Sprintf("Found %d boards:\n\n", list.Total)
	for _, g := range list.Embedded.Elements {
		scope := ""
		if g.Links != nil && g.Links.Scope != nil {
			scope = " — Scope: " + g.Links.Scope.Href
		}
		result += fmt.Sprintf("- Board #%d (%dx%d)%s\n", g.ID, g.ColumnCount, g.RowCount, scope)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) getBoard(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetBoardArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	grid, err := r.client.GetGrid(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get board: %v", err)}}}, nil
	}

	result := fmt.Sprintf("# Board #%d\n\n", grid.ID)
	result += fmt.Sprintf("- **Columns:** %d\n", grid.ColumnCount)
	result += fmt.Sprintf("- **Rows:** %d\n", grid.RowCount)
	if grid.Links != nil && grid.Links.Scope != nil {
		result += fmt.Sprintf("- **Scope:** %s\n", grid.Links.Scope.Href)
	}
	if grid.Embedded != nil && len(grid.Embedded.Widgets) > 0 {
		result += fmt.Sprintf("\n## Widgets (%d):\n\n", len(grid.Embedded.Widgets))
		for _, w := range grid.Embedded.Widgets {
			result += fmt.Sprintf("- #%d **%s** col %d-%d, row %d-%d\n",
				w.ID, w.Identifier, w.StartColumn, w.EndColumn, w.StartRow, w.EndRow)
		}
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil
}

func (r *Registry) createBoard(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateBoardArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.CreateGridOptions{
		RowCount:    args.RowCount,
		ColumnCount: args.ColumnCount,
	}
	grid, err := r.client.CreateGrid(ctx, args.ProjectID, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create board: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Board #%d created (%dx%d) for project %d!", grid.ID, grid.ColumnCount, grid.RowCount, args.ProjectID)}}}, nil
}

func (r *Registry) updateBoard(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateBoardArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	opts := &openproject.UpdateGridOptions{
		RowCount:    args.RowCount,
		ColumnCount: args.ColumnCount,
	}
	grid, err := r.client.UpdateGrid(ctx, args.ID, opts)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update board: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Board #%d updated (%dx%d)!", grid.ID, grid.ColumnCount, grid.RowCount)}}}, nil
}

func (r *Registry) deleteBoard(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteBoardArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	if err := r.client.DeleteGrid(ctx, args.ID); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete board: %v", err)}}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Board #%d deleted successfully!", args.ID)}}}, nil
}

func (r *Registry) addBoardWidget(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args AddBoardWidgetArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	widget := openproject.GridWidget{
		Identifier:  args.Identifier,
		StartRow:    args.StartRow,
		EndRow:      args.EndRow,
		StartColumn: args.StartColumn,
		EndColumn:   args.EndColumn,
	}
	if args.QueryID > 0 {
		widget.Links = &openproject.GridWidgetLinks{
			Query: &openproject.Link{Href: fmt.Sprintf("/api/v3/queries/%d", args.QueryID)},
		}
	}

	grid, err := r.client.AddGridWidget(ctx, args.BoardID, widget)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to add widget: %v", err)}}}, nil
	}

	widgetCount := 0
	if grid.Embedded != nil {
		widgetCount = len(grid.Embedded.Widgets)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Widget added to board #%d. Board now has %d widgets.", args.BoardID, widgetCount)}}}, nil
}

func (r *Registry) removeBoardWidget(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args RemoveBoardWidgetArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid arguments: %v", err)}}}, nil
	}

	grid, err := r.client.RemoveGridWidget(ctx, args.BoardID, args.WidgetID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to remove widget: %v", err)}}}, nil
	}

	widgetCount := 0
	if grid.Embedded != nil {
		widgetCount = len(grid.Embedded.Widgets)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Widget #%d removed from board #%d. Board now has %d widgets.", args.WidgetID, args.BoardID, widgetCount)}}}, nil
}
