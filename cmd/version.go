package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
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
		list, err := getClient().ListVersions(getContext(), projectID)
		if err != nil {
			return err
		}
		return output(list)
	},
}

var versionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new version",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := &openproject.CreateVersionOptions{
			ProjectID:   versionCreateProjectID,
			Name:        versionCreateName,
			Description: versionCreateDescription,
			StartDate:   versionCreateStartDate,
			EndDate:     versionCreateEndDate,
		}
		version, err := getClient().CreateVersion(getContext(), opts)
		if err != nil {
			return err
		}
		return output(version)
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
	versionCreateCmd.MarkFlagRequired("project")
	versionCreateCmd.MarkFlagRequired("name")
}
