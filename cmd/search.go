package cmd

import (
	"github.com/pinealctx/openproject-mcp/internal/openproject"
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

Searches projects, work packages, and users through their supported API filters.
Without --type, each resource type is queried independently and inaccessible
types are reported as warnings without hiding successful results.

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
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := getClient().Search(getContext(), openproject.SearchInput{
			Query: args[0],
			Type:  searchType,
			Limit: searchLimit,
		})
		if err != nil {
			return err
		}
		return output(result)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", "Resource type to search: project, work_package, or user")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", openproject.DefaultSearchLimit, "Maximum results per resource type")
}
