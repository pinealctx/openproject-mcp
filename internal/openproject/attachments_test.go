package openproject

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestListAndGetWorkPackageAttachments(t *testing.T) {
	content := []byte("release report")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/work_packages/42/attachments":
			writeJSON(t, w, attachmentCollectionFixture(42, attachmentFixture(7, 42, "report.pdf", content, "/api/v3/attachments/7/content")))
		case "/api/v3/attachments/7":
			writeJSON(t, w, attachmentFixture(7, 42, "report.pdf", content, "/api/v3/attachments/7/content"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	list, err := client.ListWorkPackageAttachments(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListWorkPackageAttachments returned error: %v", err)
	}
	if list.Total != 1 || len(list.Attachments) != 1 {
		t.Fatalf("unexpected attachment list: %#v", list)
	}
	if list.Attachments[0].FileName != "report.pdf" || list.Attachments[0].WorkPackageID != 42 {
		t.Fatalf("unexpected attachment: %#v", list.Attachments[0])
	}

	attachment, err := client.GetAttachment(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetAttachment returned error: %v", err)
	}
	encoded, err := json.Marshal(attachment)
	if err != nil {
		t.Fatalf("marshal attachment: %v", err)
	}
	if strings.Contains(string(encoded), "downloadLocation") || strings.Contains(string(encoded), "/content") {
		t.Fatalf("safe attachment output exposed download URL: %s", encoded)
	}
}

func TestUploadWorkPackageAttachmentStreamsMultipart(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "report.txt")
	if err := os.WriteFile(filePath, []byte("attachment body"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/work_packages/42/attachments" {
			http.NotFound(w, r)
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok || username != "apikey" || password != "token" {
			t.Fatalf("missing upload authentication: %q %q", username, password)
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("open multipart request: %v", err)
		}
		metadataPart, err := reader.NextPart()
		if err != nil {
			t.Fatalf("read metadata part: %v", err)
		}
		if metadataPart.FormName() != "metadata" || metadataPart.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected metadata part: %#v", metadataPart.Header)
		}
		var metadata map[string]any
		if err := json.NewDecoder(metadataPart).Decode(&metadata); err != nil {
			t.Fatalf("decode metadata: %v", err)
		}
		if metadata["fileName"] != "team-report.txt" {
			t.Fatalf("unexpected fileName: %#v", metadata["fileName"])
		}
		if nestedValue(t, metadata, "description", "raw") != "Team release report" {
			t.Fatalf("unexpected description: %#v", metadata["description"])
		}

		filePart, err := reader.NextPart()
		if err != nil {
			t.Fatalf("read file part: %v", err)
		}
		if filePart.FormName() != "file" || filePart.FileName() != "team-report.txt" {
			t.Fatalf("unexpected file part: form=%q file=%q", filePart.FormName(), filePart.FileName())
		}
		if !strings.HasPrefix(filePart.Header.Get("Content-Type"), "text/plain") {
			t.Fatalf("unexpected content type: %s", filePart.Header.Get("Content-Type"))
		}
		body, err := io.ReadAll(filePart)
		if err != nil {
			t.Fatalf("read file part: %v", err)
		}
		if string(body) != "attachment body" {
			t.Fatalf("unexpected file body: %q", body)
		}
		if _, err := reader.NextPart(); err != io.EOF {
			t.Fatalf("expected exactly two multipart parts, got %v", err)
		}
		writeJSON(t, w, attachmentFixture(9, 42, "team-report.txt", body, "/api/v3/attachments/9/content"))
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	attachment, err := client.UploadWorkPackageAttachment(context.Background(), AttachmentUploadInput{
		WorkPackageID: 42,
		FilePath:      filePath,
		FileName:      "team-report.txt",
		Description:   "Team release report",
	})
	if err != nil {
		t.Fatalf("UploadWorkPackageAttachment returned error: %v", err)
	}
	if attachment.ID != 9 || attachment.WorkPackageID != 42 {
		t.Fatalf("unexpected uploaded attachment: %#v", attachment)
	}
}

func TestDownloadAttachmentAuthenticatesSameOriginAndVerifiesDigest(t *testing.T) {
	content := []byte("verified attachment")
	var contentRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/attachments/7":
			writeJSON(t, w, attachmentFixture(7, 42, "report.txt", content, "/api/v3/attachments/7/content"))
		case "/api/v3/attachments/7/content":
			username, password, ok := r.BasicAuth()
			if !ok || username != "apikey" || password != "token" {
				t.Fatalf("same-origin download did not receive authentication")
			}
			contentRequests.Add(1)
			_, _ = w.Write(content)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "downloaded.txt")
	client := NewClientDirect(server.URL, "token", time.Second)
	result, err := client.DownloadAttachment(context.Background(), AttachmentDownloadInput{AttachmentID: 7, Destination: destination})
	if err != nil {
		t.Fatalf("DownloadAttachment returned error: %v", err)
	}
	if contentRequests.Load() != 1 || !result.DigestVerified || result.BytesWritten != int64(len(content)) {
		t.Fatalf("unexpected download result: %#v", result)
	}
	assertFileContents(t, destination, content)
}

func TestDownloadAttachmentDoesNotLeakAuthCrossOriginOrRedirect(t *testing.T) {
	content := []byte("external attachment")
	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("cross-origin request leaked Authorization header")
		}
		_, _ = w.Write(content)
	}))
	defer externalServer.Close()

	metadataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/attachments/7":
			writeJSON(t, w, attachmentFixture(7, 42, "external.txt", content, "/api/v3/attachments/7/content"))
		case "/api/v3/attachments/7/content":
			if r.Header.Get("Authorization") == "" {
				t.Fatalf("same-origin redirect request was missing Authorization")
			}
			http.Redirect(w, r, externalServer.URL+"/download", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer metadataServer.Close()

	client := NewClientDirect(metadataServer.URL, "token", time.Second)
	destination := filepath.Join(t.TempDir(), "external.txt")
	if _, err := client.DownloadAttachment(context.Background(), AttachmentDownloadInput{AttachmentID: 7, Destination: destination}); err != nil {
		t.Fatalf("redirected DownloadAttachment returned error: %v", err)
	}
	assertFileContents(t, destination, content)

	directServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, attachmentFixture(8, 42, "direct.txt", content, externalServer.URL+"/direct"))
	}))
	defer directServer.Close()
	client = NewClientDirect(directServer.URL, "token", time.Second)
	if _, err := client.DownloadAttachment(context.Background(), AttachmentDownloadInput{
		AttachmentID: 8,
		Destination:  filepath.Join(t.TempDir(), "direct.txt"),
	}); err != nil {
		t.Fatalf("cross-origin DownloadAttachment returned error: %v", err)
	}
}

func TestDownloadAttachmentRejectsUnsafeURLAndDigestMismatch(t *testing.T) {
	content := []byte("tampered-content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/attachments/7":
			fixture := attachmentFixture(7, 42, "unsafe.txt", content, "file:///tmp/unsafe")
			writeJSON(t, w, fixture)
		case "/api/v3/attachments/8":
			fixture := attachmentFixture(8, 42, "mismatch.txt", []byte("expected"), "/content/8")
			writeJSON(t, w, fixture)
		case "/api/v3/attachments/9":
			fixture := attachmentFixture(9, 42, "digest.txt", []byte("expected"), "/content/9")
			writeJSON(t, w, fixture)
		case "/content/8":
			_, _ = w.Write(content)
		case "/content/9":
			_, _ = w.Write([]byte("modified"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	if _, err := client.DownloadAttachment(context.Background(), AttachmentDownloadInput{
		AttachmentID: 7,
		Destination:  filepath.Join(t.TempDir(), "unsafe.txt"),
	}); err == nil || !strings.Contains(err.Error(), "unsupported attachment download URL scheme") {
		t.Fatalf("expected unsafe URL error, got %v", err)
	}

	directory := t.TempDir()
	destination := filepath.Join(directory, "mismatch.txt")
	if _, err := client.DownloadAttachment(context.Background(), AttachmentDownloadInput{
		AttachmentID: 8,
		Destination:  destination,
	}); err == nil || !strings.Contains(err.Error(), "attachment size mismatch") {
		t.Fatalf("expected verification error, got %v", err)
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("failed download left destination behind: %v", err)
	}
	parts, err := filepath.Glob(filepath.Join(directory, ".openproject-mcp-*.part"))
	if err != nil || len(parts) != 0 {
		t.Fatalf("failed download left temporary files: %#v, %v", parts, err)
	}

	digestDestination := filepath.Join(directory, "digest.txt")
	if _, err := client.DownloadAttachment(context.Background(), AttachmentDownloadInput{
		AttachmentID: 9,
		Destination:  digestDestination,
	}); err == nil || !strings.Contains(err.Error(), "attachment digest mismatch") {
		t.Fatalf("expected digest mismatch, got %v", err)
	}
	if _, err := os.Stat(digestDestination); !os.IsNotExist(err) {
		t.Fatalf("digest mismatch left destination behind: %v", err)
	}
}

func TestDownloadAllDisambiguatesNamesAndReportsPartialFailure(t *testing.T) {
	first := []byte("first")
	second := []byte("second")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/work_packages/42/attachments":
			writeJSON(t, w, attachmentCollectionFixture(42,
				attachmentFixture(1, 42, "report.txt", first, "/content/1"),
				attachmentFixture(2, 42, "report.txt", second, "/content/2"),
			))
		case "/content/1":
			_, _ = w.Write(first)
		case "/content/2":
			http.Error(w, "download failed", http.StatusBadGateway)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	directory := filepath.Join(t.TempDir(), "attachments")
	client := NewClientDirect(server.URL, "token", time.Second)
	result, err := client.DownloadWorkPackageAttachments(context.Background(), AttachmentDownloadAllInput{
		WorkPackageID: 42,
		Directory:     directory,
	})
	if err == nil || !strings.Contains(err.Error(), "failed to download 1 attachment") {
		t.Fatalf("expected partial failure, got %v", err)
	}
	if len(result.Downloaded) != 1 || len(result.Failed) != 1 {
		t.Fatalf("unexpected batch result: %#v", result)
	}
	assertFileContents(t, filepath.Join(directory, "report.txt"), first)
	if result.Failed[0].AttachmentID != 2 || result.Failed[0].FileName != "report.txt" {
		t.Fatalf("unexpected failed attachment: %#v", result.Failed[0])
	}
}

func TestDeleteAttachmentReturnsDeletedMetadata(t *testing.T) {
	content := []byte("delete me")
	var deleted atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/attachments/7":
			writeJSON(t, w, attachmentFixture(7, 42, "obsolete.txt", content, "/content/7"))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v3/attachments/7":
			deleted.Store(true)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	result, err := client.DeleteAttachment(context.Background(), 7)
	if err != nil {
		t.Fatalf("DeleteAttachment returned error: %v", err)
	}
	if !deleted.Load() || !result.Deleted || result.Attachment.FileName != "obsolete.txt" {
		t.Fatalf("unexpected delete result: %#v", result)
	}
}

func TestSafeAttachmentName(t *testing.T) {
	tests := map[string]string{
		"../../secret.txt":   "secret.txt",
		`..\\..\\secret.txt`: "secret.txt",
		"CON.txt":            "_CON.txt",
		"bad<name>.txt":      "bad_name_.txt",
		"...":                "attachment-9",
	}
	for input, expected := range tests {
		if actual := safeAttachmentName(input, 9); actual != expected {
			t.Errorf("safeAttachmentName(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestCommitDownloadedFileOverwritesWithRollbackBackup(t *testing.T) {
	directory := t.TempDir()
	destination := filepath.Join(directory, "report.txt")
	tempPath := filepath.Join(directory, "download.part")
	if err := os.WriteFile(destination, []byte("old"), 0o600); err != nil {
		t.Fatalf("write existing file: %v", err)
	}
	if err := os.WriteFile(tempPath, []byte("new"), 0o600); err != nil {
		t.Fatalf("write temporary file: %v", err)
	}
	if err := commitDownloadedFile(tempPath, destination, true); err != nil {
		t.Fatalf("commitDownloadedFile returned error: %v", err)
	}
	assertFileContents(t, destination, []byte("new"))
	if _, err := os.Stat(destination + ".openproject-mcp.bak"); !os.IsNotExist(err) {
		t.Fatalf("overwrite left backup behind: %v", err)
	}
}

func attachmentCollectionFixture(workPackageID int, attachments ...map[string]any) map[string]any {
	return map[string]any{
		"_type":     "Collection",
		"count":     len(attachments),
		"total":     len(attachments),
		"_links":    map[string]any{"self": map[string]any{"href": "/api/v3/work_packages/" + strconv.Itoa(workPackageID) + "/attachments"}},
		"_embedded": map[string]any{"elements": attachments},
	}
}

func attachmentFixture(id, workPackageID int, fileName string, content []byte, downloadHref string) map[string]any {
	digest := sha256.Sum256(content)
	return map[string]any{
		"_type":       "Attachment",
		"id":          id,
		"fileName":    fileName,
		"fileSize":    len(content),
		"contentType": "text/plain",
		"status":      "uploaded",
		"description": map[string]any{"format": "plain", "raw": "Fixture attachment"},
		"digest":      map[string]any{"algorithm": "sha256", "hash": hex.EncodeToString(digest[:])},
		"createdAt":   "2026-07-17T08:00:00Z",
		"_links": map[string]any{
			"self":             map[string]any{"href": "/api/v3/attachments/" + strconv.Itoa(id)},
			"container":        map[string]any{"href": "/api/v3/work_packages/" + strconv.Itoa(workPackageID)},
			"author":           map[string]any{"href": "/api/v3/users/1", "title": "Alice"},
			"downloadLocation": map[string]any{"href": downloadHref},
		},
	}
}

func assertFileContents(t *testing.T, filePath string, expected []byte) {
	t.Helper()
	// #nosec G304 -- callers pass paths created inside test-owned temporary directories.
	actual, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(actual) != string(expected) {
		t.Fatalf("unexpected file contents: %q", actual)
	}
}
