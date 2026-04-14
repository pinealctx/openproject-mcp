package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
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
		return errorResult("Invalid arguments: %v", err), nil
	}

	params := &external.ListGridsParams{}
	if args.Offset > 0 {
		params.Offset = intPtr(args.Offset)
	}
	if args.PageSize > 0 {
		params.PageSize = intPtr(args.PageSize)
	}
	if args.ProjectID > 0 {
		params.Filters = strPtr(fmt.Sprintf(`[{"scope":{"operator":"=","values":["/api/v3/projects/%d"]}}]`, args.ProjectID))
	}

	resp, err := r.client.APIClient().ListGrids(ctx, params)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to list boards: %v", err), nil
	}
	defer func() { _ = resp.Body.Close() }()

	var list external.GridCollectionModel
	if err := openproject.ReadResponse(resp, &list); err != nil {
		return errorResult("Failed to list boards: %v", err), nil
	}

	result := fmt.Sprintf("Found %d boards:\n\n", list.Total)
	for _, g := range list.UnderscoreEmbedded.Elements {
		scope := ""
		if g.UnderscoreLinks.Scope.Href != nil {
			scope = " — Scope: " + *g.UnderscoreLinks.Scope.Href
		}
		result += fmt.Sprintf("- Board #%d (%dx%d)%s\n", g.Id, g.ColumnCount, g.RowCount, scope)
	}
	return textResult(result), nil
}

func (r *Registry) getBoard(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetBoardArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	resp, err := r.client.APIClient().GetGrid(ctx, args.ID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get board: %v", err), nil
	}
	defer func() { _ = resp.Body.Close() }()

	var grid external.GridReadModel
	if err := openproject.ReadResponse(resp, &grid); err != nil {
		return errorResult("Failed to get board: %v", err), nil
	}

	result := fmt.Sprintf("# Board #%d\n\n", grid.Id)
	result += fmt.Sprintf("- **Columns:** %d\n", grid.ColumnCount)
	result += fmt.Sprintf("- **Rows:** %d\n", grid.RowCount)
	if grid.UnderscoreLinks.Scope.Href != nil {
		result += fmt.Sprintf("- **Scope:** %s\n", *grid.UnderscoreLinks.Scope.Href)
	}
	if len(grid.Widgets) > 0 {
		result += fmt.Sprintf("\n## Widgets (%d):\n\n", len(grid.Widgets))
		for _, w := range grid.Widgets {
			widgetID := ""
			if w.Id != nil {
				widgetID = fmt.Sprintf(" #%d", *w.Id)
			}
			result += fmt.Sprintf("-%s **%s** col %d-%d, row %d-%d\n",
				widgetID, w.Identifier, w.StartColumn, w.EndColumn, w.StartRow, w.EndRow)
		}
	}
	return textResult(result), nil
}

func (r *Registry) createBoard(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args CreateBoardArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.GridWriteModel{
		UnderscoreLinks: &struct {
			Scope *external.Link `json:"scope,omitempty"`
		}{
			Scope: &external.Link{Href: strPtr(fmt.Sprintf("/api/v3/projects/%d", args.ProjectID))},
		},
	}
	if args.RowCount > 0 {
		body.RowCount = intPtr(args.RowCount)
	}
	if args.ColumnCount > 0 {
		body.ColumnCount = intPtr(args.ColumnCount)
	}

	resp, err := r.client.APIClient().CreateGrid(ctx, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to create board: %v", err), nil
	}
	defer func() { _ = resp.Body.Close() }()
	var grid external.GridReadModel
	if err := openproject.ReadResponse(resp, &grid); err != nil {
		return errorResult("Failed to create board: %v", err), nil
	}
	return textResult(fmt.Sprintf("Board #%d created (%dx%d) for project %d!", grid.Id, grid.ColumnCount, grid.RowCount, args.ProjectID)), nil
}

func (r *Registry) updateBoard(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args UpdateBoardArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	body := external.GridWriteModel{}
	if args.RowCount > 0 {
		body.RowCount = intPtr(args.RowCount)
	}
	if args.ColumnCount > 0 {
		body.ColumnCount = intPtr(args.ColumnCount)
	}

	resp, err := r.client.APIClient().UpdateGrid(ctx, args.ID, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to update board: %v", err), nil
	}
	var grid external.GridReadModel
	if err := openproject.ReadResponse(resp, &grid); err != nil {
		return errorResult("Failed to update board: %v", err), nil
	}
	return textResult(fmt.Sprintf("Board #%d updated (%dx%d)!", grid.Id, grid.ColumnCount, grid.RowCount)), nil
}

func (r *Registry) deleteBoard(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args DeleteBoardArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// DeleteGrid is not available in the generated client; use raw DELETE
	if err := r.client.Delete(ctx, fmt.Sprintf("/grids/%d", args.ID)); err != nil {
		return errorResult("Failed to delete board: %v", err), nil
	}
	return textResult(fmt.Sprintf("Board #%d deleted successfully!", args.ID)), nil
}

func (r *Registry) addBoardWidget(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args AddBoardWidgetArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// First get the current board to read its widgets
	resp, err := r.client.APIClient().GetGrid(ctx, args.BoardID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get board: %v", err), nil
	}
	var grid external.GridReadModel
	if err := openproject.ReadResponse(resp, &grid); err != nil {
		return errorResult("Failed to get board: %v", err), nil
	}

	// Create the new widget
	widget := external.GridWidgetModel{
		Identifier:  args.Identifier,
		StartRow:    args.StartRow,
		EndRow:      args.EndRow,
		StartColumn: args.StartColumn,
		EndColumn:   args.EndColumn,
	}

	widgets := make([]external.GridWidgetModel, len(grid.Widgets)+1)
	copy(widgets, grid.Widgets)
	widgets[len(grid.Widgets)] = widget

	body := external.GridWriteModel{
		Widgets: &widgets,
	}

	resp, err = r.client.APIClient().UpdateGrid(ctx, args.BoardID, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to add widget: %v", err), nil
	}
	var updated external.GridReadModel
	if err := openproject.ReadResponse(resp, &updated); err != nil {
		return errorResult("Failed to add widget: %v", err), nil
	}

	return textResult(fmt.Sprintf("Widget added to board #%d. Board now has %d widgets.", args.BoardID, len(updated.Widgets))), nil
}

func (r *Registry) removeBoardWidget(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args RemoveBoardWidgetArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}

	// First get the current board
	resp, err := r.client.APIClient().GetGrid(ctx, args.BoardID)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to get board: %v", err), nil
	}
	var grid external.GridReadModel
	if err := openproject.ReadResponse(resp, &grid); err != nil {
		return errorResult("Failed to get board: %v", err), nil
	}

	// Filter out the widget with the given ID
	var widgets []external.GridWidgetModel
	for _, w := range grid.Widgets {
		if w.Id == nil || *w.Id != args.WidgetID {
			widgets = append(widgets, w)
		}
	}

	body := external.GridWriteModel{
		Widgets: &widgets,
	}

	resp, err = r.client.APIClient().UpdateGrid(ctx, args.BoardID, body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to remove widget: %v", err), nil
	}
	var updated external.GridReadModel
	if err := openproject.ReadResponse(resp, &updated); err != nil {
		return errorResult("Failed to remove widget: %v", err), nil
	}

	return textResult(fmt.Sprintf("Widget #%d removed from board #%d. Board now has %d widgets.", args.WidgetID, args.BoardID, len(updated.Widgets))), nil
}
