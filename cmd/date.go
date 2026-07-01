package cmd

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// parseDate parses a YYYY-MM-DD string into an OpenAPI date pointer.
func parseDate(s string) *openapi_types.Date {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &openapi_types.Date{Time: t}
}
