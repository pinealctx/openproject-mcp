package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

var (
	timeEntryListProjectID    int
	timeEntryListUserID       int
	timeEntryListWorkPackageID int
	timeEntryListPageSize     int
	timeEntryListSortBy       string

	timeEntryCreateProjectID    int
	timeEntryCreateWorkPackageID int
	timeEntryCreateActivityID   int
	timeEntryCreateUserID       int
	timeEntryCreateHours        string
	timeEntryCreateComment      string
	timeEntryCreateSpentOn      string

	timeEntryUpdateHours      string
	timeEntryUpdateComment    string
	timeEntryUpdateSpentOn    string
	timeEntryUpdateActivityID int
)

// timeEntryCmd represents the time-entry command.
var timeEntryCmd = &cobra.Command{
	Use:     "time-entry",
	Short:   "Manage time entries (work logs)",
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
		opts := &openproject.ListTimeEntriesOptions{
			PageSize: timeEntryListPageSize,
			SortBy:   timeEntryListSortBy,
		}
		if timeEntryListProjectID > 0 || timeEntryListUserID > 0 || timeEntryListWorkPackageID > 0 {
			var filters []openproject.TimeEntryFilter
			if timeEntryListProjectID > 0 {
				filters = append(filters, openproject.TimeEntryFilter{
					Name: "project", Operator: "=", Values: []string{strconv.Itoa(timeEntryListProjectID)},
				})
			}
			if timeEntryListUserID > 0 {
				filters = append(filters, openproject.TimeEntryFilter{
					Name: "user", Operator: "=", Values: []string{strconv.Itoa(timeEntryListUserID)},
				})
			}
			if timeEntryListWorkPackageID > 0 {
				filters = append(filters, openproject.TimeEntryFilter{
					Name: "workPackage", Operator: "=", Values: []string{strconv.Itoa(timeEntryListWorkPackageID)},
				})
			}
			opts.Filters = filters
		}
		list, err := getClient().ListTimeEntries(getContext(), opts)
		if err != nil {
			return err
		}
		return output(list)
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
		opts := &openproject.CreateTimeEntryOptions{
			Hours:       timeEntryCreateHours,
			Comment:     timeEntryCreateComment,
			SpentOn:     spentOn,
			ProjectID:   timeEntryCreateProjectID,
			WorkPackage: timeEntryCreateWorkPackageID,
			ActivityID:  timeEntryCreateActivityID,
			UserID:      timeEntryCreateUserID,
		}
		entry, err := getClient().CreateTimeEntry(getContext(), opts)
		if err != nil {
			return err
		}
		return output(entry)
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
		opts := &openproject.UpdateTimeEntryOptions{
			Hours:      timeEntryUpdateHours,
			Comment:    timeEntryUpdateComment,
			SpentOn:    timeEntryUpdateSpentOn,
			ActivityID: timeEntryUpdateActivityID,
		}
		entry, err := getClient().UpdateTimeEntry(getContext(), id, opts)
		if err != nil {
			return err
		}
		return output(entry)
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
		if err := getClient().DeleteTimeEntry(getContext(), id); err != nil {
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
		list, err := getClient().ListTimeEntryActivities(getContext())
		if err != nil {
			return err
		}
		return output(list)
	},
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
	timeEntryCreateCmd.MarkFlagRequired("hours")

	// Update flags
	timeEntryUpdateCmd.Flags().StringVarP(&timeEntryUpdateHours, "hours", "H", "", "Hours spent")
	timeEntryUpdateCmd.Flags().StringVarP(&timeEntryUpdateComment, "comment", "c", "", "Comment")
	timeEntryUpdateCmd.Flags().StringVarP(&timeEntryUpdateSpentOn, "date", "d", "", "Date (YYYY-MM-DD)")
	timeEntryUpdateCmd.Flags().IntVarP(&timeEntryUpdateActivityID, "activity", "a", 0, "Activity ID")
}
