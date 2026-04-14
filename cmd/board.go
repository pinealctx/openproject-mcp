package cmd

import (
	"fmt"
	"strconv"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
	"github.com/spf13/cobra"
)

var (
	boardListProjectID int
	boardListPageSize  int

	boardCreateProjectID   int
	boardCreateRowCount    int
	boardCreateColumnCount int

	boardUpdateRowCount    int
	boardUpdateColumnCount int

	boardWidgetIdentifier string
	boardWidgetQueryID    int
	boardWidgetStartRow   int
	boardWidgetEndRow     int
	boardWidgetStartCol   int
	boardWidgetEndCol     int
)

// boardCmd represents the board command.
var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Manage Kanban boards",
	Long: `Manage Kanban boards in OpenProject.

Boards provide a visual way to organize and track work packages. They can be
customized with multiple columns and rows, and widgets can display filtered
work package queries.

Available subcommands:
  list    List boards (optionally filtered by project)
  get     Get details of a specific board
  create  Create a new board in a project
  update  Modify board layout (rows/columns)
  delete  Delete a board
  widget  Manage board widgets (add/remove)

Examples:
  # List all boards
  openproject-mcp board list

  # List boards in a specific project
  openproject-mcp board list -p 42

  # Get board details
  openproject-mcp board get 5

  # Create a new board with 3 columns
  openproject-mcp board create -p 42 -c 3

  # Create a board with custom layout
  openproject-mcp board create -p 42 -r 2 -c 4

  # Update board dimensions
  openproject-mcp board update 5 -r 3 -c 5

  # Add a widget to a board
  openproject-mcp board widget add 5 -i work_package_query -q 10

  # Remove a widget from a board
  openproject-mcp board widget remove 5 3`,
}

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List boards",
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		params := &openproject.ListGridsParams{}
		if boardListPageSize > 0 {
			params.PageSize = ptr(boardListPageSize)
		}
		if boardListProjectID > 0 {
			filter := fmt.Sprintf(`[{"page":{"operator":"=","values":["/api/v3/projects/%d"]}}]`, boardListProjectID)
			params.Filters = ptr(filter)
		}

		resp, err := api.ListGrids(getContext(), params)
		if err != nil {
			return err
		}
		var result openproject.GridCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var boardGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get board details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid board ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.GetGrid(getContext(), id)
		if err != nil {
			return err
		}
		var result openproject.GridReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var boardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new board",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := external.GridWriteModel{
			RowCount:    ptr(boardCreateRowCount),
			ColumnCount: ptr(boardCreateColumnCount),
		}
		// Set scope link to project
		body.UnderscoreLinks = &struct {
			Scope *external.Link `json:"scope,omitempty"`
		}{
			Scope: &external.Link{Href: ptr(fmt.Sprintf("/api/v3/projects/%d", boardCreateProjectID))},
		}

		api := getClient().APIClient()
		resp, err := api.CreateGrid(getContext(), body)
		if err != nil {
			return err
		}
		var result openproject.GridReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var boardUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a board",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid board ID: %s", args[0])
		}
		body := external.GridWriteModel{}
		if cmd.Flags().Changed("rows") {
			body.RowCount = ptr(boardUpdateRowCount)
		}
		if cmd.Flags().Changed("columns") {
			body.ColumnCount = ptr(boardUpdateColumnCount)
		}

		api := getClient().APIClient()
		resp, err := api.UpdateGrid(getContext(), id, body)
		if err != nil {
			return err
		}
		var result openproject.GridReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var boardDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a board",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid board ID: %s", args[0])
		}
		// No DeleteGrid method in generated client, use raw DELETE
		if err := getClient().Delete(getContext(), fmt.Sprintf("/grids/%d", id)); err != nil {
			return err
		}
		fmt.Println("Board deleted successfully")
		return nil
	},
}

// Widget subcommands
var boardWidgetCmd = &cobra.Command{
	Use:   "widget",
	Short: "Manage board widgets",
}

var boardWidgetAddCmd = &cobra.Command{
	Use:   "add <board-id>",
	Short: "Add a widget to a board",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		boardID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid board ID: %s", args[0])
		}

		// First get the current grid to preserve existing widgets
		api := getClient().APIClient()
		resp, err := api.GetGrid(getContext(), boardID)
		if err != nil {
			return err
		}
		var grid openproject.GridReadModel
		if err := openproject.ReadResponse(resp, &grid); err != nil {
			return err
		}

		// Build new widget
		newWidget := external.GridWidgetModel{
			Identifier:  boardWidgetIdentifier,
			StartRow:    boardWidgetStartRow,
			EndRow:      boardWidgetEndRow,
			StartColumn: boardWidgetStartCol,
			EndColumn:   boardWidgetEndCol,
		}

		// Append widget and update grid
		widgets := append(grid.Widgets, newWidget)
		body := external.GridWriteModel{
			RowCount:    ptr(grid.RowCount),
			ColumnCount: ptr(grid.ColumnCount),
			Widgets:     &widgets,
		}

		resp, err = api.UpdateGrid(getContext(), boardID, body)
		if err != nil {
			return err
		}
		var result openproject.GridReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var boardWidgetRemoveCmd = &cobra.Command{
	Use:   "remove <board-id> <widget-id>",
	Short: "Remove a widget from a board",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		boardID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid board ID: %s", args[0])
		}
		widgetID, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid widget ID: %s", args[1])
		}

		// First get the current grid
		api := getClient().APIClient()
		resp, err := api.GetGrid(getContext(), boardID)
		if err != nil {
			return err
		}
		var grid openproject.GridReadModel
		if err := openproject.ReadResponse(resp, &grid); err != nil {
			return err
		}

		// Filter out the widget with the given ID
		var widgets []external.GridWidgetModel
		for _, w := range grid.Widgets {
			if w.Id == nil || *w.Id != widgetID {
				widgets = append(widgets, w)
			}
		}

		body := external.GridWriteModel{
			RowCount:    ptr(grid.RowCount),
			ColumnCount: ptr(grid.ColumnCount),
			Widgets:     &widgets,
		}

		resp, err = api.UpdateGrid(getContext(), boardID, body)
		if err != nil {
			return err
		}
		var result openproject.GridReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

func init() {
	rootCmd.AddCommand(boardCmd)
	boardCmd.AddCommand(boardListCmd)
	boardCmd.AddCommand(boardGetCmd)
	boardCmd.AddCommand(boardCreateCmd)
	boardCmd.AddCommand(boardUpdateCmd)
	boardCmd.AddCommand(boardDeleteCmd)
	boardCmd.AddCommand(boardWidgetCmd)

	boardWidgetCmd.AddCommand(boardWidgetAddCmd)
	boardWidgetCmd.AddCommand(boardWidgetRemoveCmd)

	// List flags
	boardListCmd.Flags().IntVarP(&boardListProjectID, "project", "p", 0, "Filter by project ID")
	boardListCmd.Flags().IntVarP(&boardListPageSize, "page-size", "s", 20, "Number of results per page")

	// Create flags
	boardCreateCmd.Flags().IntVarP(&boardCreateProjectID, "project", "p", 0, "Project ID (required)")
	boardCreateCmd.Flags().IntVarP(&boardCreateRowCount, "rows", "r", 1, "Number of rows")
	boardCreateCmd.Flags().IntVarP(&boardCreateColumnCount, "columns", "c", 3, "Number of columns")
	_ = boardCreateCmd.MarkFlagRequired("project")

	// Update flags
	boardUpdateCmd.Flags().IntVarP(&boardUpdateRowCount, "rows", "r", 0, "Number of rows")
	boardUpdateCmd.Flags().IntVarP(&boardUpdateColumnCount, "columns", "c", 0, "Number of columns")

	// Widget add flags
	boardWidgetAddCmd.Flags().StringVarP(&boardWidgetIdentifier, "identifier", "i", "work_package_query", "Widget identifier")
	boardWidgetAddCmd.Flags().IntVar(&boardWidgetQueryID, "query", 0, "Query ID")
	boardWidgetAddCmd.Flags().IntVar(&boardWidgetStartRow, "start-row", 1, "Start row position (1-based)")
	boardWidgetAddCmd.Flags().IntVar(&boardWidgetEndRow, "end-row", 2, "End row position (exclusive)")
	boardWidgetAddCmd.Flags().IntVar(&boardWidgetStartCol, "start-col", 1, "Start column position (1-based)")
	boardWidgetAddCmd.Flags().IntVar(&boardWidgetEndCol, "end-col", 2, "End column position (exclusive)")
}
