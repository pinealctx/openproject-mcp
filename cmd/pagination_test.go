package cmd

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	"github.com/spf13/cobra"
)

func TestListCommandsForwardPagination(t *testing.T) {
	oldClient, oldContext, oldWriter, oldFormat := client, ctx, outputWriter, flagOutput
	t.Cleanup(func() { client, ctx, outputWriter, flagOutput = oldClient, oldContext, oldWriter, oldFormat })
	ctx, outputWriter, flagOutput = context.Background(), io.Discard, "json"
	for _, tc := range []struct {
		command *cobra.Command
		path    string
		project int
	}{
		{projectListCmd, "/api/v3/projects", 0},
		{membershipListCmd, "/api/v3/memberships", 0},
		{wpListCmd, "/api/v3/work_packages", 0},
		{wpListCmd, "/api/v3/projects/3/work_packages", 3},
		{userListCmd, "/api/v3/users", 0},
		{notificationListCmd, "/api/v3/notifications", 0},
	} {
		t.Run(tc.path, func(t *testing.T) {
			requests, expectedPage := 0, 1
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.URL.Path != tc.path || r.URL.Query().Get("offset") != strconv.Itoa(expectedPage) || r.URL.Query().Get("pageSize") != "2" {
					t.Errorf("unexpected list request: %s", r.URL)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"_type":"Collection","total":5,"count":0,"_embedded":{"elements":[]},"_links":{}}`)
			}))
			defer server.Close()
			client = openproject.NewClientDirect(server.URL, "test-token", time.Second)
			oldProject := wpListProjectID
			wpListProjectID = tc.project
			defer func() { wpListProjectID = oldProject }()
			for _, name := range []string{"offset", "page-size"} {
				flag := tc.command.Flags().Lookup(name)
				value, changed := flag.Value.String(), flag.Changed
				defer func() { _ = flag.Value.Set(value); flag.Changed = changed }()
			}
			for _, page := range []string{"1", "2"} {
				expectedPage, _ = strconv.Atoi(page)
				if err := tc.command.ParseFlags([]string{"--offset", page, "--page-size", "2"}); err != nil {
					t.Fatal(err)
				}
				if err := tc.command.RunE(tc.command, nil); err != nil {
					t.Fatal(err)
				}
			}
			if requests != 2 {
				t.Fatalf("got %d requests, want 2", requests)
			}
			for _, flags := range [][]string{{"--offset", "0"}, {"--offset", "1", "--page-size", "-1"}} {
				if err := tc.command.ParseFlags(flags); err != nil {
					t.Fatal(err)
				}
				if err := tc.command.RunE(tc.command, nil); err == nil {
					t.Fatal("invalid pagination accepted")
				}
			}
			if requests != 2 {
				t.Fatal("invalid pagination reached server")
			}
		})
	}
}
