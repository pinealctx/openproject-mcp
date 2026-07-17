package cmd

import (
	"fmt"
	"strconv"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	"github.com/spf13/cobra"
)

var (
	attachmentUploadName        string
	attachmentUploadDescription string
	attachmentUploadContentType string

	attachmentDownloadDestination  string
	attachmentDownloadDirectory    string
	attachmentDownloadOverwrite    bool
	attachmentDownloadAllOverwrite bool
	attachmentDeleteYes            bool
)

var attachmentCmd = &cobra.Command{
	Use:     "attachment",
	Aliases: []string{"attachments"},
	Short:   "Manage work package attachments",
	Long: `Manage files attached to OpenProject work packages.

The CLI can list metadata, upload files, download one or all attachments, and
delete an attachment. Downloaded files are verified before they replace their
final destination.

Examples:
  openproject-mcp attachment list 123
  openproject-mcp attachment get 456 -o json
  openproject-mcp attachment upload 123 ./report.pdf --description "Release report"
  openproject-mcp attachment download 456 --destination ./report.pdf
  openproject-mcp attachment download-all 123 --directory ./ticket-123-files
  openproject-mcp attachment delete 456 --yes`,
}

var attachmentListCmd = &cobra.Command{
	Use:   "list <work-package-id>",
	Short: "List attachments for a work package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workPackageID, err := positiveID(args[0], "work package")
		if err != nil {
			return err
		}
		result, err := getClient().ListWorkPackageAttachments(getContext(), workPackageID)
		if err != nil {
			return err
		}
		return output(result)
	},
}

var attachmentGetCmd = &cobra.Command{
	Use:   "get <attachment-id>",
	Short: "Get attachment metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		attachmentID, err := positiveID(args[0], "attachment")
		if err != nil {
			return err
		}
		result, err := getClient().GetAttachment(getContext(), attachmentID)
		if err != nil {
			return err
		}
		return output(result)
	},
}

var attachmentUploadCmd = &cobra.Command{
	Use:   "upload <work-package-id> <file>",
	Short: "Upload an attachment to a work package",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		workPackageID, err := positiveID(args[0], "work package")
		if err != nil {
			return err
		}
		result, err := getClient().UploadWorkPackageAttachment(getContext(), openproject.AttachmentUploadInput{
			WorkPackageID: workPackageID,
			FilePath:      args[1],
			FileName:      attachmentUploadName,
			Description:   attachmentUploadDescription,
			ContentType:   attachmentUploadContentType,
		})
		if err != nil {
			return err
		}
		return output(result)
	},
}

var attachmentDownloadCmd = &cobra.Command{
	Use:   "download <attachment-id>",
	Short: "Download one attachment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		attachmentID, err := positiveID(args[0], "attachment")
		if err != nil {
			return err
		}
		result, err := getClient().DownloadAttachment(getContext(), openproject.AttachmentDownloadInput{
			AttachmentID: attachmentID,
			Destination:  attachmentDownloadDestination,
			Overwrite:    attachmentDownloadOverwrite,
		})
		if err != nil {
			return err
		}
		return output(result)
	},
}

var attachmentDownloadAllCmd = &cobra.Command{
	Use:   "download-all <work-package-id>",
	Short: "Download all attachments for a work package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workPackageID, err := positiveID(args[0], "work package")
		if err != nil {
			return err
		}
		result, downloadErr := getClient().DownloadWorkPackageAttachments(getContext(), openproject.AttachmentDownloadAllInput{
			WorkPackageID: workPackageID,
			Directory:     attachmentDownloadDirectory,
			Overwrite:     attachmentDownloadAllOverwrite,
		})
		if result != nil {
			if err := output(result); err != nil {
				return err
			}
		}
		return downloadErr
	},
}

var attachmentDeleteCmd = &cobra.Command{
	Use:   "delete <attachment-id>",
	Short: "Delete an attachment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		attachmentID, err := positiveID(args[0], "attachment")
		if err != nil {
			return err
		}
		if !attachmentDeleteYes {
			attachment, err := getClient().GetAttachment(getContext(), attachmentID)
			if err != nil {
				return err
			}
			if err := output(attachment); err != nil {
				return err
			}
			return fmt.Errorf("attachment deletion requires --yes: attachment #%d (%s), work package #%d", attachment.ID, attachment.FileName, attachment.WorkPackageID)
		}
		result, err := getClient().DeleteAttachment(getContext(), attachmentID)
		if err != nil {
			return err
		}
		return output(result)
	},
}

func init() {
	rootCmd.AddCommand(attachmentCmd)
	attachmentCmd.AddCommand(attachmentListCmd)
	attachmentCmd.AddCommand(attachmentGetCmd)
	attachmentCmd.AddCommand(attachmentUploadCmd)
	attachmentCmd.AddCommand(attachmentDownloadCmd)
	attachmentCmd.AddCommand(attachmentDownloadAllCmd)
	attachmentCmd.AddCommand(attachmentDeleteCmd)

	attachmentUploadCmd.Flags().StringVar(&attachmentUploadName, "name", "", "Override the uploaded file name")
	attachmentUploadCmd.Flags().StringVarP(&attachmentUploadDescription, "description", "d", "", "Attachment description")
	attachmentUploadCmd.Flags().StringVar(&attachmentUploadContentType, "content-type", "", "Override the detected MIME type")

	attachmentDownloadCmd.Flags().StringVar(&attachmentDownloadDestination, "destination", "", "Exact destination file path (default: current directory and attachment file name)")
	attachmentDownloadCmd.Flags().BoolVar(&attachmentDownloadOverwrite, "overwrite", false, "Replace an existing destination file")

	attachmentDownloadAllCmd.Flags().StringVar(&attachmentDownloadDirectory, "directory", "", "Destination directory (default: ./openproject-<work-package-id>-attachments)")
	attachmentDownloadAllCmd.Flags().BoolVar(&attachmentDownloadAllOverwrite, "overwrite", false, "Replace existing destination files")

	attachmentDeleteCmd.Flags().BoolVarP(&attachmentDeleteYes, "yes", "y", false, "Confirm permanent attachment deletion")
}

func positiveID(value, kind string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s ID: %s", kind, value)
	}
	return id, nil
}
