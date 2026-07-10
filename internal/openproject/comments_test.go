package openproject

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestListWorkPackageActivities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v3/work_packages/42/activities" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, map[string]any{
			"_type": "Collection",
			"count": 1,
			"total": 1,
			"_embedded": map[string]any{
				"elements": []any{
					map[string]any{
						"_type":    "Activity",
						"id":       7,
						"comment":  map[string]any{"raw": "Investigation completed."},
						"internal": false,
						"_links": map[string]any{
							"user": map[string]any{"title": "Alice"},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	activities, err := client.ListWorkPackageActivities(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListWorkPackageActivities returned error: %v", err)
	}
	if activities.Total != 1 || len(activities.Embedded.Elements) != 1 {
		t.Fatalf("unexpected activities: %#v", activities)
	}
	activity := activities.Embedded.Elements[0]
	if testInt(activity.ID) != 7 {
		t.Fatalf("unexpected activity id: %d", testInt(activity.ID))
	}
	if activity.Comment == nil || activity.Comment.Raw == nil || *activity.Comment.Raw != "Investigation completed." {
		t.Fatalf("unexpected activity comment: %#v", activity.Comment)
	}
}

func TestCreateWorkPackageComment(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/work_packages/42/activities" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeJSON(t, w, map[string]any{
			"_type": "Activity",
			"id":    99,
		})
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	activity, err := client.CreateWorkPackageComment(context.Background(), WorkPackageCommentInput{
		WorkPackageID: 42,
		Raw:           "Please review the deployment notes.",
		Internal:      true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPackageComment returned error: %v", err)
	}
	if testInt(activity.Id) != 99 {
		t.Fatalf("unexpected activity id: %d", testInt(activity.Id))
	}
	if raw := nestedValue(t, body, "comment", "raw"); raw != "Please review the deployment notes." {
		t.Fatalf("unexpected raw comment: %#v", raw)
	}
	if internal, ok := body["internal"].(bool); !ok || !internal {
		t.Fatalf("expected internal=true in request body, got %#v", body["internal"])
	}
}

func TestCreateWorkPackageCommentOmitsInternalWhenFalse(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/work_packages/42/activities" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeJSON(t, w, map[string]any{
			"_type": "Activity",
			"id":    100,
		})
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	if _, err := client.CreateWorkPackageComment(context.Background(), WorkPackageCommentInput{
		WorkPackageID: 42,
		Raw:           "Public follow-up.",
	}); err != nil {
		t.Fatalf("CreateWorkPackageComment returned error: %v", err)
	}
	if _, ok := body["internal"]; ok {
		t.Fatalf("internal should be omitted when false, got %#v", body)
	}
	if raw := nestedValue(t, body, "comment", "raw"); !strings.Contains(raw.(string), "Public follow-up") {
		t.Fatalf("unexpected raw comment: %#v", raw)
	}
}

func testInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
