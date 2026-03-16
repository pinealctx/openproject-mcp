package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	"github.com/spf13/cobra"
)

var (
	membershipListProjectID int
	membershipListPageSize  int

	membershipCreateProjectID int
	membershipCreateUserID    int
	membershipCreateRoles     string

	membershipUpdateRoles string
)

// membershipCmd represents the membership command.
var membershipCmd = &cobra.Command{
	Use:   "membership",
	Short: "Manage project memberships",
	Long: `Manage project memberships and user roles.

Memberships define which users have access to which projects and what roles
they have within those projects. Roles determine permissions and capabilities.

Available subcommands:
  list    List memberships (optionally filtered by project)
  get     Get details of a specific membership
  create  Add a user to a project with specified roles
  update  Change roles for an existing membership
  delete  Remove a user from a project

Examples:
  # List all memberships
  openproject-mcp membership list

  # List members of a specific project
  openproject-mcp membership list -p 42

  # Get membership details
  openproject-mcp membership get 123

  # Add user 5 to project 42 with roles 3 and 4
  openproject-mcp membership create -p 42 -u 5 -r "3,4"

  # Update membership roles
  openproject-mcp membership update 123 -r "1,2,3"

  # Remove user from project
  openproject-mcp membership delete 123

  # First, list available roles to find role IDs
  openproject-mcp role`,
	Aliases: []string{"member", "m"},
}

var membershipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List memberships",
	RunE: func(cmd *cobra.Command, args []string) error {
		if membershipListProjectID > 0 {
			list, err := getClient().ListProjectMemberships(getContext(), membershipListProjectID)
			if err != nil {
				return err
			}
			return output(list)
		}
		opts := &openproject.ListMembershipsOptions{
			PageSize: membershipListPageSize,
		}
		list, err := getClient().ListMemberships(getContext(), opts)
		if err != nil {
			return err
		}
		return output(list)
	},
}

var membershipGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get membership details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid membership ID: %s", args[0])
		}
		membership, err := getClient().GetMembership(getContext(), id)
		if err != nil {
			return err
		}
		return output(membership)
	},
}

var membershipCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add user to project",
	RunE: func(cmd *cobra.Command, args []string) error {
		roleIDs := parseRoleIDs(membershipCreateRoles)
		opts := &openproject.CreateMembershipOptions{
			ProjectID: membershipCreateProjectID,
			Principal: membershipCreateUserID,
			RoleIDs:   roleIDs,
		}
		membership, err := getClient().CreateMembership(getContext(), opts)
		if err != nil {
			return err
		}
		return output(membership)
	},
}

var membershipUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update membership roles",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid membership ID: %s", args[0])
		}
		roleIDs := parseRoleIDs(membershipUpdateRoles)
		opts := &openproject.UpdateMembershipOptions{
			RoleIDs: roleIDs,
		}
		membership, err := getClient().UpdateMembership(getContext(), id, opts)
		if err != nil {
			return err
		}
		return output(membership)
	},
}

var membershipDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove membership",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid membership ID: %s", args[0])
		}
		if err := getClient().DeleteMembership(getContext(), id); err != nil {
			return err
		}
		fmt.Println("Membership removed successfully")
		return nil
	},
}

func parseRoleIDs(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if id, err := strconv.Atoi(p); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func init() {
	rootCmd.AddCommand(membershipCmd)
	membershipCmd.AddCommand(membershipListCmd)
	membershipCmd.AddCommand(membershipGetCmd)
	membershipCmd.AddCommand(membershipCreateCmd)
	membershipCmd.AddCommand(membershipUpdateCmd)
	membershipCmd.AddCommand(membershipDeleteCmd)

	// List flags
	membershipListCmd.Flags().IntVarP(&membershipListProjectID, "project", "p", 0, "Filter by project ID")
	membershipListCmd.Flags().IntVarP(&membershipListPageSize, "page-size", "s", 20, "Number of results per page")

	// Create flags
	membershipCreateCmd.Flags().IntVarP(&membershipCreateProjectID, "project", "p", 0, "Project ID (required)")
	membershipCreateCmd.Flags().IntVarP(&membershipCreateUserID, "user", "u", 0, "User ID (required)")
	membershipCreateCmd.Flags().StringVarP(&membershipCreateRoles, "roles", "r", "", "Role IDs (comma-separated, required)")
	_ = membershipCreateCmd.MarkFlagRequired("project")
	_ = membershipCreateCmd.MarkFlagRequired("user")
	_ = membershipCreateCmd.MarkFlagRequired("roles")

	// Update flags
	membershipUpdateCmd.Flags().StringVarP(&membershipUpdateRoles, "roles", "r", "", "Role IDs (comma-separated)")
	_ = membershipUpdateCmd.MarkFlagRequired("roles")
}
