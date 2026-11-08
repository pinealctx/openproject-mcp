package openproject

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	external "github.com/pinealctx/openproject"
)

// The generated SDK omits pagination fields for projects and memberships.
// Add the documented collection parameters without replacing its API client.
func paginationEditor(offset, pageSize int) external.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		if offset < 0 || pageSize < 0 {
			return fmt.Errorf("offset and pageSize must not be negative")
		}
		query := req.URL.Query()
		if offset > 0 {
			query.Set("offset", strconv.Itoa(offset))
		}
		if pageSize > 0 {
			query.Set("pageSize", strconv.Itoa(pageSize))
		}
		req.URL.RawQuery = query.Encode()
		return nil
	}
}
