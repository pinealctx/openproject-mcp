package cmd

import (
	"fmt"
	"strconv"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	"github.com/spf13/cobra"
)

var (
	notificationListPageSize int
	notificationListUnread   bool
)

// notificationCmd represents the notification command.
var notificationCmd = &cobra.Command{
	Use:   "notification",
	Short: "Manage notifications",
	Long: `Manage OpenProject notifications.

Notifications keep you informed about activity in OpenProject, including:
  - Mentions in work packages or comments
  - Assignments to work packages
  - Status changes on watched items
  - Updates to work packages you're watching

Available subcommands:
  list      List notifications (optionally filter to unread only)
  read      Mark a specific notification as read
  read-all  Mark all notifications as read

Examples:
  # List all notifications
  openproject-mcp notification list

  # List only unread notifications
  openproject-mcp notification list -u

  # Mark a specific notification as read
  openproject-mcp notification read 123

  # Mark all notifications as read
  openproject-mcp notification read-all

  # Output notifications as JSON
  openproject-mcp notification list -o json`,
	Aliases: []string{"notify"},
}

var notificationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		params := &openproject.ListNotificationsParams{}
		if notificationListPageSize > 0 {
			params.PageSize = ptr(notificationListPageSize)
		}
		if notificationListUnread {
			params.Filters = ptr(`[{"readIAN":{"operator":"=","values":["f"]}}]`)
		}

		resp, err := api.ListNotifications(getContext(), params)
		if err != nil {
			return err
		}
		var result openproject.NotificationCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var notificationReadCmd = &cobra.Command{
	Use:   "read <id>",
	Short: "Mark notification as read",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid notification ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.ReadNotification(getContext(), id)
		if err != nil {
			return err
		}
		if err := openproject.ReadResponse(resp, nil); err != nil {
			return err
		}
		fmt.Println("Notification marked as read")
		return nil
	},
}

var notificationReadAllCmd = &cobra.Command{
	Use:   "read-all",
	Short: "Mark all notifications as read",
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		// ReadNotifications marks all matching notifications as read
		// With empty filters it marks all as read
		_, err := api.ReadNotifications(getContext(), nil)
		if err != nil {
			return err
		}
		fmt.Println("All notifications marked as read")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(notificationCmd)
	notificationCmd.AddCommand(notificationListCmd)
	notificationCmd.AddCommand(notificationReadCmd)
	notificationCmd.AddCommand(notificationReadAllCmd)

	// List flags
	notificationListCmd.Flags().IntVarP(&notificationListPageSize, "page-size", "s", 20, "Number of results per page")
	notificationListCmd.Flags().BoolVarP(&notificationListUnread, "unread", "u", false, "Show only unread notifications")
}
