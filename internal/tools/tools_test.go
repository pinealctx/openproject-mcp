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
