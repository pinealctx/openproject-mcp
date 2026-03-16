package cmd

import (
	"fmt"
	"strconv"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
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
		opts := &openproject.ListGridsOptions{
			PageSize: boardListPageSize,
		}
		if boardListProjectID > 0 {
			opts.Filters = fmt.Sprintf(`[{"scope":{"operator":"=","values":["/api/v3/projects/%d"]}}]`, boardListProjectID)
		}
		list, err := getClient().ListGrids(getContext(), opts)
		if err != nil {
			return err
		}
		return output(list)
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
		grid, err := getClient().GetGrid(getContext(), id)
		if err != nil {
			return err
		}
		return output(grid)
	},
}

var boardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new board",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := &openproject.CreateGridOptions{
			RowCount:    boardCreateRowCount,
			ColumnCount: boardCreateColumnCount,
		}
		grid, err := getClient().CreateGrid(getContext(), boardCreateProjectID, opts)
		if err != nil {
			return err
		}
		return output(grid)
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
		opts := &openproject.UpdateGridOptions{}
		if cmd.Flags().Changed("rows") {
			opts.RowCount = boardUpdateRowCount
		}
		if cmd.Flags().Changed("columns") {
			opts.ColumnCount = boardUpdateColumnCount
		}
		grid, err := getClient().UpdateGrid(getContext(), id, opts)
		if err != nil {
			return err
		}
		return output(grid)
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
		if err := getClient().DeleteGrid(getContext(), id); err != nil {
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
		widget := openproject.GridWidget{
			Identifier:  boardWidgetIdentifier,
			StartRow:    boardWidgetStartRow,
			EndRow:      boardWidgetEndRow,
			StartColumn: boardWidgetStartCol,
			EndColumn:   boardWidgetEndCol,
		}
		if boardWidgetQueryID > 0 && widget.Links == nil {
			widget.Links = &openproject.GridWidgetLinks{
				Query: &openproject.Link{Href: fmt.Sprintf("/api/v3/queries/%d", boardWidgetQueryID)},
			}
		}
		grid, err := getClient().AddGridWidget(getContext(), boardID, widget)
		if err != nil {
			return err
		}
		return output(grid)
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
		grid, err := getClient().RemoveGridWidget(getContext(), boardID, widgetID)
		if err != nil {
			return err
		}
		return output(grid)
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
