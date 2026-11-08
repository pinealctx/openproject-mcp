package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

func TestProjectAndMembershipToolsForwardPagination(t *testing.T) {
	for _, kind := range []string{"projects", "memberships"} {
		t.Run(kind, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v3/"+kind || r.URL.Query().Get("offset") != "2" || r.URL.Query().Get("pageSize") != "3" {
					t.Errorf("unexpected request: %s", r.URL)
				}
				_, _ = io.WriteString(w, `{"total":8,"_embedded":{"elements":[]}}`)
			}))
			defer server.Close()
			registry := NewRegistry(openproject.NewClientDirect(server.URL, "test-token", time.Second))
			request := &mcp.CallToolRequest{Params: &mcp.CallToolParams{Arguments: json.RawMessage(`{"offset":2,"pageSize":3}`)}}
			call := registry.listProjects
			if kind == "memberships" {
				call = registry.listMemberships
			}
			result, err := call(context.Background(), request)
			if err != nil || result.IsError {
				t.Fatalf("call failed: %v, %+v", err, result)
			}
			if !strings.Contains(result.Content[0].(*mcp.TextContent).Text, "Showing 0 of 8") {
				t.Fatal("page incorrectly described as complete collection")
			}
		})
	}
}
