package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	openapi_types "github.com/oapi-codegen/runtime/types"
	external "github.com/pinealctx/openproject"
)

// errorResult creates an error tool result.
func errorResult(msg string, args ...any) *mcp.CallToolResult {
	text := msg
	if len(args) > 0 {
		text = fmt.Sprintf(msg, args...)
	}
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

// textResult creates a successful text tool result.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

// parseArgs parses arguments from the request into target struct.
func parseArgs(args any, target any) error {
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// derefStr returns the dereferenced string or empty.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// strPtr returns a pointer to the string, or nil if empty.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(i int) *int {
	return &i
}

// normalizeSortBy converts "name:asc" or "name:asc,updatedAt:desc" format
// to OpenProject API JSON array format: [["name","asc"]]
func normalizeSortBy(s string) string {
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "[") {
		return s // already in JSON format
	}
	parts := strings.Split(s, ",")
	var pairs []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		segments := strings.SplitN(p, ":", 2)
		field := segments[0]
		dir := "asc"
		if len(segments) > 1 {
			dir = segments[1]
		}
		pairs = append(pairs, fmt.Sprintf(`["%s","%s"]`, field, dir))
	}
	return "[" + strings.Join(pairs, ",") + "]"
}

// parseDatePtr parses a YYYY-MM-DD string into an openapi_types.Date pointer.
func parseDatePtr(s string) *openapi_types.Date {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	d := openapi_types.Date{Time: t}
	return &d
}

// formatUser formats a UserModel as markdown.
func formatUser(u *external.UserModel) string {
	result := fmt.Sprintf("# %s\n\n", u.Name)
	result += fmt.Sprintf("- **ID:** %d\n", u.Id)
	if u.Login != nil {
		result += fmt.Sprintf("- **Login:** %s\n", *u.Login)
	}
	if u.Email != nil {
		result += fmt.Sprintf("- **Email:** %s\n", *u.Email)
	}
	if u.FirstName != nil {
		result += fmt.Sprintf("- **First Name:** %s\n", *u.FirstName)
	}
	if u.LastName != nil {
		result += fmt.Sprintf("- **Last Name:** %s\n", *u.LastName)
	}
	if u.Admin != nil {
		result += fmt.Sprintf("- **Admin:** %v\n", *u.Admin)
	}
	if u.Status != nil {
		result += fmt.Sprintf("- **Status:** %s\n", *u.Status)
	}
	if u.Language != nil {
		result += fmt.Sprintf("- **Language:** %s\n", *u.Language)
	}
	return result
}

// formatProject formats a ProjectModel as markdown.
func formatProject(p *external.ProjectModel) string {
	result := fmt.Sprintf("# %s\n\n", derefStr(p.Name))
	result += fmt.Sprintf("- **ID:** %d\n", derefInt(p.Id))
	result += fmt.Sprintf("- **Identifier:** %s\n", derefStr(p.Identifier))
	result += fmt.Sprintf("- **Active:** %v\n", derefBool(p.Active))
	result += fmt.Sprintf("- **Public:** %v\n", derefBool(p.Public))
	if p.Description != nil && p.Description.Raw != nil && *p.Description.Raw != "" {
		result += fmt.Sprintf("\n## Description\n%s\n", *p.Description.Raw)
	}
	return result
}
