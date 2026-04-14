package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
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
		api := getClient().APIClient()
		params := &external.ListMembershipsParams{}

		if membershipListProjectID > 0 {
			filter := fmt.Sprintf(`[{"project":{"operator":"=","values":["%d"]}}]`, membershipListProjectID)
			params.Filters = ptr(filter)
		}

		resp, err := api.ListMemberships(getContext(), params)
		if err != nil {
			return err
		}
		var result external.MembershipCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
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
		api := getClient().APIClient()
		resp, err := api.GetMembership(getContext(), id)
		if err != nil {
			return err
		}
		var result external.MembershipReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var membershipCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add user to project",
	RunE: func(cmd *cobra.Command, args []string) error {
		roleIDs := parseRoleIDs(membershipCreateRoles)
		roleLinks := make([]external.Link, len(roleIDs))
		for i, rid := range roleIDs {
			roleLinks[i] = external.Link{Href: ptr(fmt.Sprintf("/api/v3/roles/%d", rid))}
		}

		body := external.MembershipWriteModel{}
		body.UnderscoreLinks.Principal = &external.Link{Href: ptr(fmt.Sprintf("/api/v3/users/%d", membershipCreateUserID))}
		body.UnderscoreLinks.Project = &external.Link{Href: ptr(fmt.Sprintf("/api/v3/projects/%d", membershipCreateProjectID))}
		body.UnderscoreLinks.Roles = &roleLinks

		api := getClient().APIClient()
		resp, err := api.CreateMembership(getContext(), body)
		if err != nil {
			return err
		}
		var result external.MembershipReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
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
		roleLinks := make([]external.Link, len(roleIDs))
		for i, rid := range roleIDs {
			roleLinks[i] = external.Link{Href: ptr(fmt.Sprintf("/api/v3/roles/%d", rid))}
		}

		body := external.MembershipWriteModel{}
		body.UnderscoreLinks.Roles = &roleLinks

		api := getClient().APIClient()
		resp, err := api.UpdateMembership(getContext(), id, body)
		if err != nil {
			return err
		}
		var result external.MembershipReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
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
		api := getClient().APIClient()
		resp, err := api.DeleteMembership(getContext(), id)
		if err != nil {
			return err
		}
		if err := openproject.ReadResponse(resp, nil); err != nil {
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
