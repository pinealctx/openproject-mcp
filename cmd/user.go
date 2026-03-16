package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

var (
	userListPageSize int
	userListSortBy   string
)

// userCmd represents the user command.
var userCmd = &cobra.Command{
	Use:     "user",
	Short:   "Manage users",
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
		opts := &openproject.ListUsersOptions{
			PageSize: userListPageSize,
			SortBy:   userListSortBy,
		}
		users, err := getClient().ListUsers(getContext(), opts)
		if err != nil {
			return err
		}
		return output(users)
	},
}

var userGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get user details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid user ID: %s", args[0])
		}
		user, err := getClient().GetUser(getContext(), id)
		if err != nil {
			return err
		}
		return output(user)
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
