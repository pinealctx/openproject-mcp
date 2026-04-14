package cmd

import (
	"fmt"
	"strconv"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List work package statuses",
	Long: `List all available work package statuses.

Statuses represent the current state of a work package in its lifecycle.
Common statuses include: New, In Progress, Resolved, Closed, etc.

The status IDs returned can be used when creating or updating work packages
with the --status flag.

Examples:
  # List all statuses
  openproject-mcp status

  # Output as JSON to find status IDs
  openproject-mcp status -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		resp, err := api.ListStatuses(getContext())
		if err != nil {
			return err
		}
		var result openproject.StatusCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

// priorityCmd represents the priority command.
var priorityCmd = &cobra.Command{
	Use:   "priority",
	Short: "List work package priorities",
	Long: `List all available work package priorities.

Priorities indicate the importance or urgency of a work package.
Common priorities include: Low, Normal, High, Urgent, Immediate.

The priority IDs returned can be used when creating or updating work packages
with the --priority flag.

Examples:
  # List all priorities
  openproject-mcp priority

  # Output as JSON to find priority IDs
  openproject-mcp priority -o json`,
	Aliases: []string{"priorities"},
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		resp, err := api.ListAllPriorities(getContext())
		if err != nil {
			return err
		}
		var result openproject.PriorityCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

// typeCmd represents the type command.
var typeCmd = &cobra.Command{
	Use:   "type",
	Short: "List work package types",
	Long: `List all available work package types.

Types categorize work packages and determine their behavior and available attributes.
Common types include: Task, Bug, Feature, Milestone, User Story.

Types can be global or project-specific. Use without arguments for global types,
or provide a project ID to see types available in that project.

The type IDs returned can be used when creating work packages with the --type flag.

Examples:
  # List all global types
  openproject-mcp type

  # List types available in a specific project
  openproject-mcp type 42

  # Output as JSON to find type IDs
  openproject-mcp type -o json`,
	Aliases: []string{"types"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		var projectID int
		if len(args) > 0 {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid project ID: %s", args[0])
			}
			projectID = id
		}

		if projectID > 0 {
			resp, err := api.ListTypesAvailableInAProject(getContext(), projectID)
			if err != nil {
				return err
			}
			var result openproject.TypesByWorkspaceModel
			if err := openproject.ReadResponse(resp, &result); err != nil {
				return err
			}
			return output(&result)
		}

		resp, err := api.ListAllTypes(getContext())
		if err != nil {
			return err
		}
		var result openproject.TypesByWorkspaceModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

// roleCmd represents the role command.
var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "List user roles",
	Long: `List all available roles.

Roles define what permissions a user has within a project. Each role contains
a set of permissions that control what actions the user can perform.

Common roles include: Project Admin, Member, Reader, etc.

The role IDs returned are used when adding users to projects (membership create)
or updating membership roles.

Examples:
  # List all roles
  openproject-mcp role

  # Output as JSON to find role IDs
  openproject-mcp role -o json

  # Use role IDs when adding a user to a project
  openproject-mcp membership create -p 42 -u 5 -r "3,4"`,
	Aliases: []string{"roles"},
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		resp, err := api.ListRoles(getContext(), nil)
		if err != nil {
			return err
		}
		var collection struct {
			UnderscoreEmbedded struct {
				Elements []openproject.RoleModel `json:"elements"`
			} `json:"_embedded"`
		}
		if err := openproject.ReadResponse(resp, &collection); err != nil {
			return err
		}
		return outputRoleList(collection.UnderscoreEmbedded.Elements)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(priorityCmd)
	rootCmd.AddCommand(typeCmd)
	rootCmd.AddCommand(roleCmd)
}
