package openproject

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUpdateWorkPackageFetchesLockVersionBeforePatch(t *testing.T) {
	var seen []string
	var patchBody map[string]any

	client, closeServer := newWorkPackageTestClient(t, 7, func(r *http.Request, body map[string]any) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPatch {
			patchBody = body
		}
	})
	defer closeServer()

	progress := 65
	assigneeID := 9
	accountableID := 10
	wp, err := client.UpdateWorkPackage(context.Background(), WorkPackageUpdateInput{
		ID:             42,
		Subject:        "Updated subject",
		Description:    "Updated body",
		StatusID:       "1",
		PriorityID:     "8",
		AssigneeID:     &assigneeID,
		AccountableID:  &accountableID,
		StartDate:      "2026-07-01",
		DueDate:        "2026-07-15",
		EstimatedTime:  "PT4H",
		PercentageDone: &progress,
	})
	if err != nil {
		t.Fatalf("UpdateWorkPackage returned error: %v", err)
	}
	if wp.Subject != "Updated subject" {
		t.Fatalf("unexpected subject: %q", wp.Subject)
	}

	assertSequence(t, seen, []string{
		"GET /api/v3/work_packages/42",
		"PATCH /api/v3/work_packages/42",
	})
	assertNumber(t, patchBody["lockVersion"], 7)
	assertEqual(t, patchBody["subject"], "Updated subject")
	assertEqual(t, nestedValue(t, patchBody, "description", "format"), "markdown")
	assertEqual(t, nestedValue(t, patchBody, "description", "raw"), "Updated body")
	assertEqual(t, patchBody["startDate"], "2026-07-01")
	assertEqual(t, patchBody["dueDate"], "2026-07-15")
	assertEqual(t, patchBody["estimatedTime"], "PT4H")
	assertNumber(t, patchBody["percentageDone"], 65)
	assertEqual(t, nestedValue(t, patchBody, "_links", "status", "href"), "/api/v3/statuses/1")
	assertEqual(t, nestedValue(t, patchBody, "_links", "priority", "href"), "/api/v3/priorities/8")
	assertEqual(t, nestedValue(t, patchBody, "_links", "assignee", "href"), "/api/v3/users/9")
	assertEqual(t, nestedValue(t, patchBody, "_links", "responsible", "href"), "/api/v3/users/10")
}

func TestCreateWorkPackageSetsAssigneeAndAccountable(t *testing.T) {
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/projects/3/work_packages" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode create body: %v", err)
		}
		writeJSON(t, w, map[string]any{"_type": "WorkPackage", "id": 42, "subject": "New task"})
	}))
	defer server.Close()

	assigneeID := 9
	accountableID := 10
	client := NewClientDirect(server.URL, "token", time.Second)
	if _, err := client.CreateWorkPackage(context.Background(), WorkPackageCreateInput{
		ProjectID:     3,
		Subject:       "New task",
		AssigneeID:    &assigneeID,
		AccountableID: &accountableID,
	}); err != nil {
		t.Fatalf("CreateWorkPackage returned error: %v", err)
	}

	assertEqual(t, nestedValue(t, requestBody, "_links", "assignee", "href"), "/api/v3/users/9")
	assertEqual(t, nestedValue(t, requestBody, "_links", "responsible", "href"), "/api/v3/users/10")
}

func TestUpdateWorkPackageClearsAssigneeAndAccountable(t *testing.T) {
	var patchBody map[string]any
	client, closeServer := newWorkPackageTestClient(t, 7, func(r *http.Request, body map[string]any) {
		if r.Method == http.MethodPatch {
			patchBody = body
		}
	})
	defer closeServer()

	if _, err := client.UpdateWorkPackage(context.Background(), WorkPackageUpdateInput{
		ID:               42,
		ClearAssignee:    true,
		ClearAccountable: true,
	}); err != nil {
		t.Fatalf("UpdateWorkPackage returned error: %v", err)
	}

	assertNilLinkHref(t, patchBody, "assignee")
	assertNilLinkHref(t, patchBody, "responsible")
}

func TestWorkPackagePeopleValidationRunsBeforeNetwork(t *testing.T) {
	client := NewClientDirect("https://project.example", "token", time.Second)
	client.apiClient.Client = failingDoer{}

	invalidID := 0
	tests := []struct {
		name  string
		input WorkPackageUpdateInput
		want  string
	}{
		{name: "invalid assignee", input: WorkPackageUpdateInput{ID: 42, AssigneeID: &invalidID}, want: "assignee user ID must be greater than zero"},
		{name: "conflicting assignee", input: WorkPackageUpdateInput{ID: 42, AssigneeID: ptr(9), ClearAssignee: true}, want: "assignee cannot be set and cleared"},
		{name: "invalid accountable", input: WorkPackageUpdateInput{ID: 42, AccountableID: &invalidID}, want: "accountable user ID must be greater than zero"},
		{name: "conflicting accountable", input: WorkPackageUpdateInput{ID: 42, AccountableID: ptr(10), ClearAccountable: true}, want: "accountable cannot be set and cleared"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.UpdateWorkPackage(context.Background(), test.input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("UpdateWorkPackage error = %v, want %q", err, test.want)
			}
		})
	}

	_, err := client.CreateWorkPackage(context.Background(), WorkPackageCreateInput{
		ProjectID:     3,
		Subject:       "New task",
		AccountableID: &invalidID,
	})
	if err == nil || !strings.Contains(err.Error(), "accountable user ID must be greater than zero") {
		t.Fatalf("CreateWorkPackage error = %v", err)
	}
}

func TestSetAndRemoveWorkPackageParentUseCurrentLockVersion(t *testing.T) {
	tests := []struct {
		name       string
		call       func(context.Context, *Client) error
		assertBody func(*testing.T, map[string]any)
	}{
		{
			name: "set parent",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.SetWorkPackageParent(ctx, 42, 99)
				return err
			},
			assertBody: func(t *testing.T, body map[string]any) {
				assertEqual(t, nestedValue(t, body, "_links", "parent", "href"), "/api/v3/work_packages/99")
			},
		},
		{
			name: "remove parent",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.RemoveWorkPackageParent(ctx, 42)
				return err
			},
			assertBody: func(t *testing.T, body map[string]any) {
				parent, ok := nestedValue(t, body, "_links", "parent").(map[string]any)
				if !ok {
					t.Fatalf("expected parent link map, got %#v", nestedValue(t, body, "_links", "parent"))
				}
				if _, ok := parent["href"]; !ok {
					t.Fatalf("expected href key in parent link: %#v", parent)
				}
				if parent["href"] != nil {
					t.Fatalf("expected nil parent href, got %#v", parent["href"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seen []string
			var patchBody map[string]any
			client, closeServer := newWorkPackageTestClient(t, 11, func(r *http.Request, body map[string]any) {
				seen = append(seen, r.Method+" "+r.URL.Path)
				if r.Method == http.MethodPatch {
					patchBody = body
				}
			})
			defer closeServer()

			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
			assertSequence(t, seen, []string{
				"GET /api/v3/work_packages/42",
				"PATCH /api/v3/work_packages/42",
			})
			assertNumber(t, patchBody["lockVersion"], 11)
			tt.assertBody(t, patchBody)
		})
	}
}

func TestListWorkPackageChildrenUsesParentFilter(t *testing.T) {
	var filters string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v3/work_packages" {
			http.NotFound(w, r)
			return
		}
		filters = r.URL.Query().Get("filters")
		writeJSON(t, w, map[string]any{
			"_type": "Collection",
			"count": 0,
			"total": 0,
			"_embedded": map[string]any{
				"elements": []any{},
			},
		})
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	if _, err := client.ListWorkPackageChildren(context.Background(), 42); err != nil {
		t.Fatalf("ListWorkPackageChildren returned error: %v", err)
	}
	if !strings.Contains(filters, `"parent"`) || !strings.Contains(filters, `"42"`) {
		t.Fatalf("expected parent filter for work package 42, got %q", filters)
	}
}

func TestUpdateWorkPackageRequiresFetchedLockVersion(t *testing.T) {
	patchCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(t, w, map[string]any{
				"_type":   "WorkPackage",
				"id":      42,
				"subject": "Existing subject",
			})
		case http.MethodPatch:
			patchCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	_, err := client.UpdateWorkPackage(context.Background(), WorkPackageUpdateInput{
		ID:      42,
		Subject: "Updated subject",
	})
	if err == nil || !strings.Contains(err.Error(), "missing lockVersion") {
		t.Fatalf("expected missing lockVersion error, got %v", err)
	}
	if patchCalled {
		t.Fatal("PATCH should not be called without a fetched lockVersion")
	}
}

func TestRawPatchTransportErrorDoesNotPanic(t *testing.T) {
	client := NewClientDirect("https://project.example", "token", time.Second)
	client.apiClient.Client = failingDoer{}

	err := client.Patch(context.Background(), "/work_packages/42", map[string]any{"lockVersion": 1}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "network down") {
		t.Fatalf("expected transport error, got %v", err)
	}
}

func TestDecodeResponseHandlesNilResponse(t *testing.T) {
	err := DecodeResponse(nil, errors.New("network down"), nil)
	if err == nil || !strings.Contains(err.Error(), "network down") {
		t.Fatalf("expected original error, got %v", err)
	}

	err = DecodeResponse(nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "response is nil") {
		t.Fatalf("expected nil response error, got %v", err)
	}
}

type failingDoer struct{}

func (failingDoer) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("network down")
}

func newWorkPackageTestClient(t *testing.T, lockVersion int, record func(*http.Request, map[string]any)) (*Client, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/work_packages/42" {
			http.NotFound(w, r)
			return
		}

		var body map[string]any
		switch r.Method {
		case http.MethodGet:
			record(r, nil)
			writeJSON(t, w, map[string]any{
				"_type":       "WorkPackage",
				"id":          42,
				"lockVersion": lockVersion,
				"subject":     "Existing subject",
			})
		case http.MethodPatch:
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode patch body: %v", err)
			}
			record(r, body)
			writeJSON(t, w, map[string]any{
				"_type":       "WorkPackage",
				"id":          42,
				"lockVersion": lockVersion + 1,
				"subject":     "Updated subject",
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	return NewClientDirect(server.URL, "token", time.Second), server.Close
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

func assertSequence(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected sequence length: got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected sequence: got %#v want %#v", got, want)
		}
	}
}

func nestedValue(t *testing.T, root map[string]any, path ...string) any {
	t.Helper()
	var current any = root
	for _, key := range path {
		obj, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("expected map while resolving %v, got %#v", path, current)
		}
		current, ok = obj[key]
		if !ok {
			t.Fatalf("missing key %q while resolving %v in %#v", key, path, obj)
		}
	}
	return current
}

func assertEqual(t *testing.T, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func assertNumber(t *testing.T, got any, want float64) {
	t.Helper()
	number, ok := got.(float64)
	if !ok {
		t.Fatalf("got %#v (%T), want JSON number %v", got, got, want)
	}
	if number != want {
		t.Fatalf("got %v, want %v", number, want)
	}
}

func assertNilLinkHref(t *testing.T, body map[string]any, linkName string) {
	t.Helper()
	link, ok := nestedValue(t, body, "_links", linkName).(map[string]any)
	if !ok {
		t.Fatalf("expected %s link map, got %#v", linkName, nestedValue(t, body, "_links", linkName))
	}
	if href, ok := link["href"]; !ok || href != nil {
		t.Fatalf("expected nil %s href, got %#v", linkName, link)
	}
}
