package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	projectapi "github.com/pinealctx/openproject-mcp/internal/openproject"
)

func TestPositiveID(t *testing.T) {
	if id, err := positiveID("42", "attachment"); err != nil || id != 42 {
		t.Fatalf("positiveID returned %d, %v", id, err)
	}
	for _, value := range []string{"0", "-1", "invalid"} {
		if _, err := positiveID(value, "attachment"); err == nil {
			t.Fatalf("positiveID(%q) should fail", value)
		}
	}
}

func TestAttachmentOutputDoesNotExposeDownloadURL(t *testing.T) {
	var buffer bytes.Buffer
	previousWriter := outputWriter
	previousFormat := flagOutput
	outputWriter = &buffer
	flagOutput = "json"
	t.Cleanup(func() {
		outputWriter = previousWriter
		flagOutput = previousFormat
	})

	attachment := &projectapi.Attachment{ID: 7, WorkPackageID: 42, FileName: "report.pdf"}
	if err := output(attachment); err != nil {
		t.Fatalf("output attachment: %v", err)
	}
	if strings.Contains(buffer.String(), "downloadLocation") || strings.Contains(buffer.String(), "download_url") {
		t.Fatalf("attachment output exposed a download URL: %s", buffer.String())
	}
}

func TestAttachmentDeleteRequiresYes(t *testing.T) {
	var deleted atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleted.Store(true)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet || r.URL.Path != "/api/v3/attachments/7" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          7,
			"fileName":    "report.pdf",
			"fileSize":    10,
			"contentType": "application/pdf",
			"status":      "uploaded",
			"description": map[string]any{"format": "plain", "raw": "Report"},
			"digest":      map[string]any{"algorithm": "sha256", "hash": "abc"},
			"createdAt":   "2026-07-17T08:00:00Z",
			"_links": map[string]any{
				"self":             map[string]any{"href": "/api/v3/attachments/7"},
				"container":        map[string]any{"href": "/api/v3/work_packages/42"},
				"author":           map[string]any{"href": "/api/v3/users/1", "title": "Alice"},
				"downloadLocation": map[string]any{"href": "/api/v3/attachments/7/content"},
			},
		})
	}))
	defer server.Close()

	previousClient := client
	previousContext := ctx
	previousWriter := outputWriter
	previousFormat := flagOutput
	previousYes := attachmentDeleteYes
	var buffer bytes.Buffer
	client = projectapi.NewClientDirect(server.URL, "token", time.Second)
	ctx = context.Background()
	outputWriter = &buffer
	flagOutput = "text"
	attachmentDeleteYes = false
	t.Cleanup(func() {
		client = previousClient
		ctx = previousContext
		outputWriter = previousWriter
		flagOutput = previousFormat
		attachmentDeleteYes = previousYes
	})

	err := attachmentDeleteCmd.RunE(attachmentDeleteCmd, []string{"7"})
	if err == nil || !strings.Contains(err.Error(), "requires --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
	if deleted.Load() {
		t.Fatal("delete request was sent without --yes")
	}
	if !strings.Contains(buffer.String(), "Attachment ID: 7") || !strings.Contains(buffer.String(), "Work package ID: 42") {
		t.Fatalf("delete preview did not identify the target: %s", buffer.String())
	}
}
