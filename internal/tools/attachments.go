package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	projectapi "github.com/pinealctx/openproject-mcp/internal/openproject"
)

type ListWorkPackageAttachmentsArgs struct {
	WorkPackageID int `json:"workPackageId"`
}

type GetAttachmentArgs struct {
	AttachmentID int `json:"attachmentId"`
}

func (r *Registry) registerAttachmentTools(server *mcp.Server) {
	addTool(server, "list_work_package_attachments",
		"List file attachment metadata for a work package without downloading file content",
		newSchema(schemaProps{
			"workPackageId": schemaInt("Work package ID"),
		}, "workPackageId"),
		r.listWorkPackageAttachments)

	addTool(server, "get_attachment",
		"Get metadata for a work package attachment without exposing its download URL",
		newSchema(schemaProps{
			"attachmentId": schemaInt("Attachment ID"),
		}, "attachmentId"),
		r.getAttachment)
}

func (r *Registry) listWorkPackageAttachments(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ListWorkPackageAttachmentsArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}
	list, err := r.client.ListWorkPackageAttachments(ctx, args.WorkPackageID)
	if err != nil {
		return errorResult("Failed to list attachments: %v", err), nil
	}
	if len(list.Attachments) == 0 {
		return textResult(fmt.Sprintf("No attachments found for work package #%d.", args.WorkPackageID)), nil
	}

	var result strings.Builder
	fmt.Fprintf(&result, "# Attachments for work package #%d\n\n", args.WorkPackageID)
	for index := range list.Attachments {
		result.WriteString(formatAttachmentMetadata(&list.Attachments[index]))
		result.WriteString("\n")
	}
	return textResult(result.String()), nil
}

func (r *Registry) getAttachment(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args GetAttachmentArgs
	if err := parseArgs(req.Params.Arguments, &args); err != nil {
		return errorResult("Invalid arguments: %v", err), nil
	}
	attachment, err := r.client.GetAttachment(ctx, args.AttachmentID)
	if err != nil {
		return errorResult("Failed to get attachment: %v", err), nil
	}
	return textResult("# Attachment\n\n" + formatAttachmentMetadata(attachment)), nil
}

func formatAttachmentMetadata(attachment *projectapi.Attachment) string {
	var result strings.Builder
	fmt.Fprintf(&result, "- **ID:** %d\n", attachment.ID)
	fmt.Fprintf(&result, "- **Work package ID:** %d\n", attachment.WorkPackageID)
	fmt.Fprintf(&result, "- **File name:** %s\n", attachment.FileName)
	if attachment.FileSize != nil {
		fmt.Fprintf(&result, "- **File size:** %d bytes\n", *attachment.FileSize)
	}
	fmt.Fprintf(&result, "- **Content type:** %s\n", attachment.ContentType)
	fmt.Fprintf(&result, "- **Status:** %s\n", attachment.Status)
	if attachment.Author != "" {
		fmt.Fprintf(&result, "- **Author:** %s\n", attachment.Author)
	}
	if attachment.Description != "" {
		fmt.Fprintf(&result, "- **Description:** %s\n", attachment.Description)
	}
	if !attachment.CreatedAt.IsZero() {
		fmt.Fprintf(&result, "- **Created:** %s\n", attachment.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if attachment.Digest.Algorithm != "" && attachment.Digest.Hash != "" {
		fmt.Fprintf(&result, "- **Digest:** %s:%s\n", attachment.Digest.Algorithm, attachment.Digest.Hash)
	}
	return result.String()
}
