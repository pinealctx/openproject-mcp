package openproject

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSearchAggregatesAccessibleResourcesAndWarnsOnForbiddenType(t *testing.T) {
	var requested []string
	var requestedMu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestedMu.Lock()
		requested = append(requested, request.URL.Path)
		requestedMu.Unlock()

		switch request.URL.Path {
		case "/api/v3/work_packages":
			assertSearchFilter(t, request, "search", "**", "pre-dev")
			writeJSON(t, response, map[string]any{
				"_type": "WorkPackageCollection",
				"count": 1,
				"total": 2,
				"_embedded": map[string]any{"elements": []any{map[string]any{
					"_type":   "WorkPackage",
					"id":      42,
					"subject": "Prepare pre-dev release",
					"_links":  map[string]any{"status": map[string]any{"title": "In progress"}},
				}}},
			})
		case "/api/v3/projects":
			assertSearchFilter(t, request, "name_and_identifier", "~", "pre-dev")
			writeJSON(t, response, map[string]any{
				"_type": "ProjectCollection",
				"count": 1,
				"total": 1,
				"_embedded": map[string]any{"elements": []any{map[string]any{
					"_type":      "Project",
					"id":         7,
					"name":       "Pre-Dev Operations",
					"identifier": "pre-dev-operations",
				}}},
			})
		case "/api/v3/users":
			response.WriteHeader(http.StatusForbidden)
			writeJSON(t, response, map[string]any{
				"_type":           "Error",
				"errorIdentifier": "urn:openproject-org:api:v3:errors:Unauthenticated",
				"message":         "Insufficient permissions.",
			})
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	result, err := client.Search(context.Background(), SearchInput{Query: " pre-dev ", Limit: 5})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if result.Query != "pre-dev" || result.Count != 2 || result.Total != 3 {
		t.Fatalf("Search returned unexpected summary: %#v", result)
	}
	if len(result.Results) != 2 || result.Results[0].Type != SearchTypeWorkPackage || result.Results[1].Type != SearchTypeProject {
		t.Fatalf("Search returned unexpected results: %#v", result.Results)
	}
	if len(result.Warnings) != 1 || result.Warnings[0].ResourceType != SearchTypeUser || !strings.Contains(result.Warnings[0].Message, "403 Forbidden") {
		t.Fatalf("Search returned unexpected warnings: %#v", result.Warnings)
	}
	if strings.Join(requested, ",") != "/api/v3/work_packages,/api/v3/projects,/api/v3/users" {
		t.Fatalf("Search requested unexpected endpoints: %v", requested)
	}
}

func TestSearchWorkPackagesEscapesFilterAndHonorsLimit(t *testing.T) {
	query := `pre "dev" \\ rollout`
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v3/work_packages" {
			http.NotFound(response, request)
			return
		}
		assertSearchFilter(t, request, "search", "**", query)
		if got := request.URL.Query().Get("pageSize"); got != "2" {
			t.Errorf("pageSize = %q, want 2", got)
		}
		writeJSON(t, response, map[string]any{
			"_type":     "WorkPackageCollection",
			"count":     0,
			"total":     0,
			"_embedded": map[string]any{"elements": []any{}},
		})
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	result, err := client.Search(context.Background(), SearchInput{
		Query: query,
		Type:  SearchTypeWorkPackage,
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if result.Count != 0 || result.Total != 0 || result.ResourceType != SearchTypeWorkPackage {
		t.Fatalf("Search returned unexpected result: %#v", result)
	}
}

func TestSearchExplicitForbiddenTypeReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v3/users" {
			t.Errorf("requested path = %q, want /api/v3/users", request.URL.Path)
		}
		response.WriteHeader(http.StatusForbidden)
		writeJSON(t, response, map[string]any{
			"_type":           "Error",
			"errorIdentifier": "urn:openproject-org:api:v3:errors:Unauthenticated",
			"message":         "Insufficient permissions.",
		})
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	_, err := client.Search(context.Background(), SearchInput{Query: "team", Type: SearchTypeUser})
	if err == nil || !strings.Contains(err.Error(), "search user") || !strings.Contains(err.Error(), "403 Forbidden") {
		t.Fatalf("Search error = %v, want typed 403 error", err)
	}
}

func TestSearchValidatesInputBeforeNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		t.Fatalf("unexpected network request to %s", request.URL.Path)
	}))
	defer server.Close()
	client := NewClientDirect(server.URL, "token", time.Second)

	tests := []struct {
		name  string
		input SearchInput
		want  string
	}{
		{name: "blank query", input: SearchInput{Query: "  "}, want: "query is required"},
		{name: "unknown type", input: SearchInput{Query: "release", Type: "ticket"}, want: "search type must be"},
		{name: "negative limit", input: SearchInput{Query: "release", Limit: -1}, want: "limit must be"},
		{name: "excessive limit", input: SearchInput{Query: "release", Limit: MaxSearchLimit + 1}, want: "limit must be"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.Search(context.Background(), test.input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Search error = %v, want %q", err, test.want)
			}
		})
	}
}

func assertSearchFilter(t *testing.T, request *http.Request, name, operator, value string) {
	t.Helper()
	var filters []map[string]struct {
		Operator string   `json:"operator"`
		Values   []string `json:"values"`
	}
	if err := json.Unmarshal([]byte(request.URL.Query().Get("filters")), &filters); err != nil {
		t.Fatalf("decode filters: %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("filters = %#v, want one filter", filters)
	}
	filter, ok := filters[0][name]
	if !ok || filter.Operator != operator || len(filter.Values) != 1 || filter.Values[0] != value {
		t.Fatalf("filter = %#v, want %s %s %q", filters, name, operator, value)
	}
}
