package tools

import "testing"

func TestDefaultToolModeIncludesWorkPackageComments(t *testing.T) {
	registry := NewRegistryWithMode(nil, "default", "")
	available := map[string]bool{}
	for _, tool := range registry.ListAvailableTools() {
		available[tool] = true
	}

	for _, tool := range []string{"list_work_package_activities", "create_work_package_comment"} {
		if !available[tool] {
			t.Fatalf("default tool mode should include %s", tool)
		}
	}
}

func TestDefaultToolModeIncludesReadOnlyAttachments(t *testing.T) {
	registry := NewRegistryWithMode(nil, "default", "")
	available := map[string]bool{}
	for _, tool := range registry.ListAvailableTools() {
		available[tool] = true
	}

	for _, tool := range []string{"list_work_package_attachments", "get_attachment"} {
		if !available[tool] {
			t.Fatalf("default tool mode should include %s", tool)
		}
	}
	for _, tool := range []string{"upload_attachment", "download_attachment", "delete_attachment"} {
		if available[tool] {
			t.Fatalf("MCP tool mode must not expose %s", tool)
		}
	}
}
