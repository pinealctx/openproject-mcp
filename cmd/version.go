package cmd

import (
	"fmt"
	"strconv"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
	"github.com/spf13/cobra"
)

var (
	versionCreateProjectID   int
	versionCreateName        string
	versionCreateDescription string
	versionCreateStartDate   string
	versionCreateEndDate     string
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Manage project versions/milestones",
	Aliases: []string{"ver"},
}

var versionListCmd = &cobra.Command{
	Use:   "list <project-id>",
	Short: "List versions for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid project ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.ListVersionsAvailableInAProject(getContext(), projectID)
		if err != nil {
			return err
		}
		var result external.VersionCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var versionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new version",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := external.VersionWriteModel{
			Name: ptr(versionCreateName),
		}
		if versionCreateDescription != "" {
			fmt_ := external.FormattableFormat("markdown")
			body.Description = &external.Formattable{Format: &fmt_, Raw: ptr(versionCreateDescription)}
		}
		if versionCreateStartDate != "" {
			body.StartDate = parseDate(versionCreateStartDate)
		}
		if versionCreateEndDate != "" {
			body.EndDate = parseDate(versionCreateEndDate)
		}
		// Set the defining project link
		body.UnderscoreLinks = &struct {
			DefiningProject *external.Link `json:"definingProject,omitempty"`
		}{
			DefiningProject: &external.Link{Href: ptr(fmt.Sprintf("/api/v3/projects/%d", versionCreateProjectID))},
		}

		api := getClient().APIClient()
		resp, err := api.CreateVersion(getContext(), body)
		if err != nil {
			return err
		}
		var result external.VersionReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.AddCommand(versionListCmd)
	versionCmd.AddCommand(versionCreateCmd)

	// Create flags
	versionCreateCmd.Flags().IntVarP(&versionCreateProjectID, "project", "p", 0, "Project ID (required)")
	versionCreateCmd.Flags().StringVarP(&versionCreateName, "name", "n", "", "Version name (required)")
	versionCreateCmd.Flags().StringVarP(&versionCreateDescription, "description", "d", "", "Description")
	versionCreateCmd.Flags().StringVar(&versionCreateStartDate, "start", "", "Start date (YYYY-MM-DD)")
	versionCreateCmd.Flags().StringVar(&versionCreateEndDate, "end", "", "End date (YYYY-MM-DD)")
	_ = versionCreateCmd.MarkFlagRequired("project")
	_ = versionCreateCmd.MarkFlagRequired("name")
}
