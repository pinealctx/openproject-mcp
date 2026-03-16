package cmd

import (
	"github.com/spf13/cobra"
)

var (
	searchType  string
	searchLimit int
)

// searchCmd represents the search command.
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search OpenProject",
	Long: `Search across OpenProject resources.

Provides full-text search across projects, work packages, and users.
Results are ranked by relevance and can be filtered by resource type.

Examples:
  # Search for all items containing "bug"
  openproject-mcp search "bug"

  # Search only in projects
  openproject-mcp search "website" -t project

  # Search only in work packages
  openproject-mcp search "authentication" -t work_package

  # Search only in users
  openproject-mcp search "john" -t user

  # Limit results to 5 items
  openproject-mcp search "sprint" -l 5

  # Output as JSON for further processing
  openproject-mcp search "urgent" -o json`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		results, err := getClient().Search(getContext(), query, searchType, searchLimit)
		if err != nil {
			return err
		}
		return output(results)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", "Resource type to search (project, work_package, user)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results")
}
