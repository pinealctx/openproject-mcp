package tools

import (
	"strings"
	"testing"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

func TestFormatSearchResultsIncludesTypesAndPartialWarnings(t *testing.T) {
	text := formatSearchResults(&openproject.SearchResults{
		Query: "release",
		Count: 1,
		Total: 3,
		Results: []openproject.SearchResult{{
			Type:   openproject.SearchTypeWorkPackage,
			ID:     42,
			Title:  "Prepare release",
			Status: "In progress",
		}},
		Warnings: []openproject.SearchWarning{{
			ResourceType: openproject.SearchTypeUser,
			Message:      "search user: forbidden",
		}},
	})

	for _, expected := range []string{
		`Found 1 of 3 matching resources for "release"`,
		"[work_package] #42 Prepare release",
		"In progress",
		"Warning: search user: forbidden",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("formatted search results missing %q:\n%s", expected, text)
		}
	}
}
