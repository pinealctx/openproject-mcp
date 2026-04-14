package cmd

import (
	"github.com/spf13/cobra"
)

// connectionCmd represents the connection command.
var connectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "Connection-related commands",
}

var connectionTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to OpenProject",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := getClient().TestConnection(getContext())
		if err != nil {
			return err
		}
		email := ""
		if user.Email != nil {
			email = *user.Email
		}
		cmd.Printf("Connection successful! Logged in as: %s (%s)", user.Name, email)
		return nil
	},
}

var connectionInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get API info",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := getClient().GetAPIRoot(getContext())
		if err != nil {
			return err
		}
		return output(info)
	},
}

func init() {
	rootCmd.AddCommand(connectionCmd)
	connectionCmd.AddCommand(connectionTestCmd)
	connectionCmd.AddCommand(connectionInfoCmd)
}
