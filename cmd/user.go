package cmd

import (
	"fmt"
	"strconv"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
	"github.com/spf13/cobra"
)

var (
	userListPageSize int
	userListSortBy   string
)

// userCmd represents the user command.
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long: `Manage OpenProject users.

Users are the people who interact with OpenProject. Each user has a profile,
authentication credentials, and can be assigned to work packages and projects.

Available subcommands:
  list    List all users (with optional sorting)
  get     Get details of a specific user by ID
  me      Get details of the currently authenticated user

Examples:
  # List all users
  openproject-mcp user list

  # List users sorted by login
  openproject-mcp user list --sort-by "login:asc"

  # Get user by ID
  openproject-mcp user get 5

  # Get current user info
  openproject-mcp user me

  # Output in JSON format
  openproject-mcp user list -o json`,
	Aliases: []string{"u"},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		params := &external.ListUsersParams{}
		if userListPageSize > 0 {
			params.PageSize = ptr(userListPageSize)
		}
		if userListSortBy != "" {
			params.SortBy = ptr(normalizeSortBy(userListSortBy))
		}

		resp, err := api.ListUsers(getContext(), params)
		if err != nil {
			return err
		}
		var result external.UserCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var userGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get user details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		// Validate it's numeric
		if _, err := strconv.Atoi(id); err != nil {
			return fmt.Errorf("invalid user ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.ViewUser(getContext(), id)
		if err != nil {
			return err
		}
		var result external.UserModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var userMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current user details",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := getClient().GetCurrentUser(getContext())
		if err != nil {
			return err
		}
		return output(user)
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userGetCmd)
	userCmd.AddCommand(userMeCmd)

	// List flags
	userListCmd.Flags().IntVarP(&userListPageSize, "page-size", "s", 20, "Number of results per page")
	userListCmd.Flags().StringVar(&userListSortBy, "sort-by", "name:asc", "Sort criteria")
}
