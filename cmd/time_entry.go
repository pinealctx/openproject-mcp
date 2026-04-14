package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
	"github.com/spf13/cobra"
)

var (
	timeEntryListProjectID     int
	timeEntryListUserID        int
	timeEntryListWorkPackageID int
	timeEntryListPageSize      int
	timeEntryListSortBy        string

	timeEntryCreateProjectID     int
	timeEntryCreateWorkPackageID int
	timeEntryCreateActivityID    int
	timeEntryCreateUserID        int
	timeEntryCreateHours         string
	timeEntryCreateComment       string
	timeEntryCreateSpentOn       string

	timeEntryUpdateHours      string
	timeEntryUpdateComment    string
	timeEntryUpdateSpentOn    string
	timeEntryUpdateActivityID int
)

// timeEntryCmd represents the time-entry command.
var timeEntryCmd = &cobra.Command{
	Use:   "time-entry",
	Short: "Manage time entries (work logs)",
	Long: `Manage time entries for tracking work hours.

Time entries record the amount of time spent on projects or specific work packages.
They are essential for project tracking, billing, and resource management.

Available subcommands:
  list        List time entries (filterable by project, user, work package)
  create      Log a new time entry
  update      Modify an existing time entry
  delete      Remove a time entry
  activities  List available activity types (e.g., Development, Design, Meeting)

Examples:
  # List all time entries
  openproject-mcp time-entry list

  # List time entries for a specific project
  openproject-mcp time-entry list -p 42

  # List time entries for a specific user
  openproject-mcp time-entry list -u 5

  # List time entries for a specific work package
  openproject-mcp time-entry list -w 123

  # Log 4 hours of work today
  openproject-mcp time-entry create -H 4 -c "Implemented feature X"

  # Log 8.5 hours on a specific date
  openproject-mcp time-entry create -H 8.5 -d 2024-01-15 -c "Sprint work"

  # Log time for a specific work package
  openproject-mcp time-entry create -H 2 -w 123 -c "Bug fix"

  # List available activity types
  openproject-mcp time-entry activities

  # Update a time entry
  openproject-mcp time-entry update 456 -H 6 -c "Updated description"`,
	Aliases: []string{"time", "te"},
}

var timeEntryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List time entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		params := &openproject.ListTimeEntriesParams{}
		if timeEntryListPageSize > 0 {
			params.PageSize = ptr(timeEntryListPageSize)
		}
		if timeEntryListSortBy != "" {
			params.SortBy = ptr(normalizeSortBy(timeEntryListSortBy))
		}

		// Build filters
		var filters []string
		if timeEntryListProjectID > 0 {
			filters = append(filters, fmt.Sprintf(`{"project_id":{"operator":"=","values":["%d"]}}`, timeEntryListProjectID))
		}
		if timeEntryListUserID > 0 {
			filters = append(filters, fmt.Sprintf(`{"user_id":{"operator":"=","values":["%d"]}}`, timeEntryListUserID))
		}
		if timeEntryListWorkPackageID > 0 {
			filters = append(filters, fmt.Sprintf(`{"entity_id":{"operator":"=","values":["%d"]}}`, timeEntryListWorkPackageID))
		}
		if len(filters) > 0 {
			filterStr := "[" + joinStrings(filters, ",") + "]"
			params.Filters = ptr(filterStr)
		}

		resp, err := api.ListTimeEntries(getContext(), params)
		if err != nil {
			return err
		}
		var result openproject.TimeEntryCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var timeEntryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Log time entry",
	RunE: func(cmd *cobra.Command, args []string) error {
		spentOn := timeEntryCreateSpentOn
		if spentOn == "" {
			spentOn = time.Now().Format("2006-01-02")
		}
		body := external.TimeEntryModel{
			Hours: ptr(timeEntryCreateHours),
		}
		if timeEntryCreateComment != "" {
			fmt_ := external.FormattableFormat("markdown")
			body.Comment = &external.Formattable{Format: &fmt_, Raw: ptr(timeEntryCreateComment)}
		}

		// Set links for project, work package, activity, user
		if timeEntryCreateProjectID > 0 {
			body.UnderscoreLinks.Project = external.Link{Href: ptr(fmt.Sprintf("/api/v3/projects/%d", timeEntryCreateProjectID))}
		}
		if timeEntryCreateWorkPackageID > 0 {
			body.UnderscoreLinks.Entity = external.Link{Href: ptr(fmt.Sprintf("/api/v3/work_packages/%d", timeEntryCreateWorkPackageID))}
		}
		if timeEntryCreateActivityID > 0 {
			body.UnderscoreLinks.Activity = external.Link{Href: ptr(fmt.Sprintf("/api/v3/time_entry_activities/%d", timeEntryCreateActivityID))}
		}
		if timeEntryCreateUserID > 0 {
			body.UnderscoreLinks.User = external.Link{Href: ptr(fmt.Sprintf("/api/v3/users/%d", timeEntryCreateUserID))}
		}

		// Set spentOn via raw POST since TimeEntryModel.SpentOn is *openapi_types.Date
		// We'll use the raw client to create the time entry
		rawBody := map[string]interface{}{
			"hours": timeEntryCreateHours,
			"_links": buildTimeEntryLinks(),
		}
		if timeEntryCreateComment != "" {
			rawBody["comment"] = map[string]interface{}{
				"format": "markdown",
				"raw":    timeEntryCreateComment,
			}
		}
		rawBody["spentOn"] = spentOn

		var result openproject.TimeEntryModel
		if err := getClient().Post(getContext(), "/time_entries", rawBody, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var timeEntryUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update time entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid time entry ID: %s", args[0])
		}

		// Build update body using raw PATCH
		body := map[string]interface{}{}
		if timeEntryUpdateHours != "" {
			body["hours"] = timeEntryUpdateHours
		}
		if timeEntryUpdateComment != "" {
			body["comment"] = map[string]interface{}{
				"format": "markdown",
				"raw":    timeEntryUpdateComment,
			}
		}
		if timeEntryUpdateSpentOn != "" {
			body["spentOn"] = timeEntryUpdateSpentOn
		}
		if timeEntryUpdateActivityID > 0 {
			body["_links"] = map[string]interface{}{
				"activity": map[string]interface{}{
					"href": fmt.Sprintf("/api/v3/time_entry_activities/%d", timeEntryUpdateActivityID),
				},
			}
		}

		var result openproject.TimeEntryModel
		if err := getClient().Patch(getContext(), fmt.Sprintf("/time_entries/%d", id), body, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var timeEntryDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete time entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid time entry ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.DeleteTimeEntry(getContext(), id)
		if err != nil {
			return err
		}
		if err := openproject.ReadResponse(resp, nil); err != nil {
			return err
		}
		fmt.Println("Time entry deleted successfully")
		return nil
	},
}

var timeEntryActivitiesCmd = &cobra.Command{
	Use:   "activities",
	Short: "List time entry activities",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use raw GET since there's no generated method for listing activities
		var result interface{}
		if err := getClient().Get(getContext(), "/time_entries/activities", &result); err != nil {
			return err
		}
		return output(result)
	},
}

// buildTimeEntryLinks constructs the _links portion for time entry creation.
func buildTimeEntryLinks() map[string]interface{} {
	links := map[string]interface{}{}
	if timeEntryCreateProjectID > 0 {
		links["project"] = map[string]interface{}{
			"href": fmt.Sprintf("/api/v3/projects/%d", timeEntryCreateProjectID),
		}
	}
	if timeEntryCreateWorkPackageID > 0 {
		links["entity"] = map[string]interface{}{
			"href": fmt.Sprintf("/api/v3/work_packages/%d", timeEntryCreateWorkPackageID),
		}
	}
	if timeEntryCreateActivityID > 0 {
		links["activity"] = map[string]interface{}{
			"href": fmt.Sprintf("/api/v3/time_entry_activities/%d", timeEntryCreateActivityID),
		}
	}
	if timeEntryCreateUserID > 0 {
		links["user"] = map[string]interface{}{
			"href": fmt.Sprintf("/api/v3/users/%d", timeEntryCreateUserID),
		}
	}
	return links
}

// joinStrings joins string slice with separator.
func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func init() {
	rootCmd.AddCommand(timeEntryCmd)
	timeEntryCmd.AddCommand(timeEntryListCmd)
	timeEntryCmd.AddCommand(timeEntryCreateCmd)
	timeEntryCmd.AddCommand(timeEntryUpdateCmd)
	timeEntryCmd.AddCommand(timeEntryDeleteCmd)
	timeEntryCmd.AddCommand(timeEntryActivitiesCmd)

	// List flags
	timeEntryListCmd.Flags().IntVarP(&timeEntryListProjectID, "project", "p", 0, "Filter by project ID")
	timeEntryListCmd.Flags().IntVarP(&timeEntryListUserID, "user", "u", 0, "Filter by user ID")
	timeEntryListCmd.Flags().IntVarP(&timeEntryListWorkPackageID, "work-package", "w", 0, "Filter by work package ID")
	timeEntryListCmd.Flags().IntVarP(&timeEntryListPageSize, "page-size", "s", 20, "Number of results per page")
	timeEntryListCmd.Flags().StringVar(&timeEntryListSortBy, "sort-by", "spentOn:desc", "Sort criteria")

	// Create flags
	timeEntryCreateCmd.Flags().IntVarP(&timeEntryCreateProjectID, "project", "p", 0, "Project ID")
	timeEntryCreateCmd.Flags().IntVarP(&timeEntryCreateWorkPackageID, "work-package", "w", 0, "Work package ID")
	timeEntryCreateCmd.Flags().IntVarP(&timeEntryCreateActivityID, "activity", "a", 0, "Activity ID")
	timeEntryCreateCmd.Flags().IntVarP(&timeEntryCreateUserID, "user", "u", 0, "User ID (defaults to current user)")
	timeEntryCreateCmd.Flags().StringVarP(&timeEntryCreateHours, "hours", "H", "", "Hours spent (e.g., 8.5 or PT8H30M) (required)")
	timeEntryCreateCmd.Flags().StringVarP(&timeEntryCreateComment, "comment", "c", "", "Comment")
	timeEntryCreateCmd.Flags().StringVarP(&timeEntryCreateSpentOn, "date", "d", "", "Date (YYYY-MM-DD, defaults to today)")
	_ = timeEntryCreateCmd.MarkFlagRequired("hours")

	// Update flags
	timeEntryUpdateCmd.Flags().StringVarP(&timeEntryUpdateHours, "hours", "H", "", "Hours spent")
	timeEntryUpdateCmd.Flags().StringVarP(&timeEntryUpdateComment, "comment", "c", "", "Comment")
	timeEntryUpdateCmd.Flags().StringVarP(&timeEntryUpdateSpentOn, "date", "d", "", "Date (YYYY-MM-DD)")
	timeEntryUpdateCmd.Flags().IntVarP(&timeEntryUpdateActivityID, "activity", "a", 0, "Activity ID")
}
