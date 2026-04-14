package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
)

func (r *Registry) registerConfigurationTools(server *mcp.Server) {
	addTool(server, "view_configuration",
		"View the OpenProject instance configuration (feature flags, settings, etc.)",
		noSchema,
		r.viewConfiguration)
}

func (r *Registry) viewConfiguration(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := r.client.APIClient().ViewConfiguration(ctx)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return errorResult("Failed to view configuration: %v", err), nil
	}
	var cfg external.ConfigurationModel
	if err := openproject.ReadResponse(resp, &cfg); err != nil {
		return errorResult("Failed to view configuration: %v", err), nil
	}

	result := "# OpenProject Configuration\n\n"
	result += fmt.Sprintf("- **Host Name:** %s\n", derefStr(cfg.HostName))
	result += fmt.Sprintf("- **Duration Format:** %s\n", derefStr(cfg.DurationFormat))
	if cfg.MaximumAttachmentFileSize != nil {
		result += fmt.Sprintf("- **Max Attachment Size:** %d bytes\n", *cfg.MaximumAttachmentFileSize)
	}
	if cfg.PerPageOptions != nil {
		result += fmt.Sprintf("- **Per Page Options:** %v\n", *cfg.PerPageOptions)
	}
	if cfg.ActiveFeatureFlags != nil {
		result += fmt.Sprintf("\n## Active Feature Flags (%d)\n\n", len(*cfg.ActiveFeatureFlags))
		for _, flag := range *cfg.ActiveFeatureFlags {
			result += fmt.Sprintf("- `%s`\n", flag)
		}
	}
	return textResult(result), nil
}
