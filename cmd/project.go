package cmd

import (
	"fmt"
	"strconv"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
	"github.com/spf13/cobra"
)

var (
	projectListPageSize int
	projectListSortBy   string
	projectListFilters  string

	projectCreateName        string
	projectCreateIdentifier  string
	projectCreateDescription string
	projectCreatePublic      bool

	projectUpdateName        string
	projectUpdateDescription string
	projectUpdatePublic      bool
	projectUpdateActive      bool
)

// projectCmd represents the project command.
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage OpenProject projects",
	Long: `Manage OpenProject projects.

Projects are the main organizational unit in OpenProject. Each project can contain
work packages (tasks, bugs, features, etc.), versions, members, and other resources.

Available subcommands:
  list      List all projects (with optional filtering)
  get       Get details of a specific project
  create    Create a new project
  update    Update project properties
  delete    Delete a project (requires confirmation)

Examples:
  # List all projects
  openproject-mcp project list

  # List projects sorted by creation date
  openproject-mcp project list --sort-by "createdAt:desc"

  # Get project with ID 42
  openproject-mcp project get 42

  # Create a new project
  openproject-mcp project create -n "Website Redesign" -i "website-redesign" -d "Redesign company website"

  # Update project name
  openproject-mcp project update 42 -n "New Project Name"

  # Archive (deactivate) a project
  openproject-mcp project update 42 --active=false

  # Delete a project
  openproject-mcp project delete 42`,
	Aliases: []string{"proj", "p"},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long: `List all projects visible to the authenticated user.

The output can be customized with sorting and filtering options.
Use -o json for machine-readable output.

Examples:
  # List all projects (default)
  openproject-mcp project list

  # List projects sorted by name descending
  openproject-mcp project list --sort-by "name:desc"

  # List only active projects using filters
  openproject-mcp project list --filters '[{"active":{"operator":"=","values":["true"]}}]'

  # Output as JSON
  openproject-mcp project list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		params := &openproject.ListProjectsParams{}
		if projectListSortBy != "" {
			params.SortBy = ptr(normalizeSortBy(projectListSortBy))
		}
		if projectListFilters != "" {
			params.Filters = ptr(projectListFilters)
		}

		resp, err := api.ListProjects(getContext(), params)
		if err != nil {
			return err
		}
		var result openproject.ProjectCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get project details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid project ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.ViewProject(getContext(), id)
		if err != nil {
			return err
		}
		var result openproject.ProjectModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := external.ProjectModel{
			Identifier: ptr(projectCreateIdentifier),
			Name:       ptr(projectCreateName),
			Public:     ptr(projectCreatePublic),
		}
		if projectCreateDescription != "" {
			fmt := external.FormattableFormat("markdown")
			body.Description = &external.Formattable{Format: &fmt, Raw: ptr(projectCreateDescription)}
		}

		api := getClient().APIClient()
		resp, err := api.CreateProject(getContext(), body)
		if err != nil {
			return err
		}
		var result openproject.ProjectModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid project ID: %s", args[0])
		}
		body := external.ProjectModel{}
		if projectUpdateName != "" {
			body.Name = ptr(projectUpdateName)
		}
		if projectUpdateDescription != "" {
			fmt := external.FormattableFormat("markdown")
			body.Description = &external.Formattable{Format: &fmt, Raw: ptr(projectUpdateDescription)}
		}
		if cmd.Flags().Changed("public") {
			body.Public = ptr(projectUpdatePublic)
		}
		if cmd.Flags().Changed("active") {
			body.Active = ptr(projectUpdateActive)
		}

		api := getClient().APIClient()
		resp, err := api.UpdateProject(getContext(), id, body)
		if err != nil {
			return err
		}
		var result openproject.ProjectModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid project ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.DeleteProject(getContext(), id)
		if err != nil {
			return err
		}
		if err := openproject.ReadResponse(resp, nil); err != nil {
			return err
		}
		fmt.Println("Project deleted successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectDeleteCmd)

	// List flags
	projectListCmd.Flags().IntVarP(&projectListPageSize, "page-size", "s", 20, "Number of results per page")
	projectListCmd.Flags().StringVar(&projectListSortBy, "sort-by", "name:asc", "Sort criteria")
	projectListCmd.Flags().StringVar(&projectListFilters, "filters", "", "Raw JSON filter string")

	// Create flags
	projectCreateCmd.Flags().StringVarP(&projectCreateName, "name", "n", "", "Project name (required)")
	projectCreateCmd.Flags().StringVarP(&projectCreateIdentifier, "identifier", "i", "", "Project identifier (required)")
	projectCreateCmd.Flags().StringVarP(&projectCreateDescription, "description", "d", "", "Project description")
	projectCreateCmd.Flags().BoolVarP(&projectCreatePublic, "public", "p", false, "Make project public")
	_ = projectCreateCmd.MarkFlagRequired("name")
	_ = projectCreateCmd.MarkFlagRequired("identifier")

	// Update flags
	projectUpdateCmd.Flags().StringVarP(&projectUpdateName, "name", "n", "", "Project name")
	projectUpdateCmd.Flags().StringVarP(&projectUpdateDescription, "description", "d", "", "Project description")
	projectUpdateCmd.Flags().BoolVarP(&projectUpdatePublic, "public", "p", false, "Make project public")
	projectUpdateCmd.Flags().BoolVarP(&projectUpdateActive, "active", "a", true, "Project active status")
}
