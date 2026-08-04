package cmd

import (
	"bytes"
	"strings"
	"testing"

	projectapi "github.com/pinealctx/openproject-mcp/internal/openproject"
)

func TestSearchTextOutputIsConciseAndIncludesWarnings(t *testing.T) {
	results := &projectapi.SearchResults{
		Query: "release",
		Limit: 5,
		Count: 2,
		Total: 7,
		Results: []projectapi.SearchResult{
			{Type: projectapi.SearchTypeWorkPackage, ID: 42, Title: "Prepare release", Status: "In progress"},
			{Type: projectapi.SearchTypeProject, ID: 7, Title: "Release Operations", Identifier: "release-operations"},
		},
		Warnings: []projectapi.SearchWarning{{ResourceType: projectapi.SearchTypeUser, Message: "search user: forbidden"}},
	}

	var buffer bytes.Buffer
	previousWriter := outputWriter
	previousFormat := flagOutput
	outputWriter = &buffer
	flagOutput = "text"
	t.Cleanup(func() {
		outputWriter = previousWriter
		flagOutput = previousFormat
	})

	if err := output(results); err != nil {
		t.Fatalf("output search results: %v", err)
	}
	for _, expected := range []string{
		"Search: release",
		"Returned: 2",
		"work_package",
		"Prepare release",
		"In progress",
		"release-operations",
		"Warning: search user: forbidden",
	} {
		if !strings.Contains(buffer.String(), expected) {
			t.Fatalf("search output missing %q:\n%s", expected, buffer.String())
		}
	}
}
