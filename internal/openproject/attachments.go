package openproject

import (
	"context"
	"crypto/md5"  // #nosec G501 -- OpenProject may publish MD5 as an integrity checksum.
	"crypto/sha1" // #nosec G505 -- OpenProject may publish SHA-1 as an integrity checksum.
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	external "github.com/pinealctx/openproject"
)

// Attachment is the safe, user-facing representation of an OpenProject attachment.
// Download URLs stay internal because they may be short-lived signed URLs.
type Attachment struct {
	ID            int              `json:"id"`
	WorkPackageID int              `json:"work_package_id"`
	FileName      string           `json:"file_name"`
	FileSize      *int64           `json:"file_size,omitempty"`
	ContentType   string           `json:"content_type"`
	Status        string           `json:"status"`
	Description   string           `json:"description,omitempty"`
	Author        string           `json:"author,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	Digest        AttachmentDigest `json:"digest"`
}

type AttachmentDigest struct {
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
}

type AttachmentCollection struct {
	WorkPackageID int          `json:"work_package_id"`
	Total         int          `json:"total"`
	Attachments   []Attachment `json:"attachments"`
}

type AttachmentUploadInput struct {
	WorkPackageID int
	FilePath      string
	FileName      string
	Description   string
	ContentType   string
}

type AttachmentDownloadInput struct {
	AttachmentID int
	Destination  string
	Overwrite    bool
}

type AttachmentDownloadAllInput struct {
	WorkPackageID int
	Directory     string
	Overwrite     bool
}

type AttachmentDownloadResult struct {
	AttachmentID   int    `json:"attachment_id"`
	WorkPackageID  int    `json:"work_package_id"`
	FileName       string `json:"file_name"`
	Path           string `json:"path"`
	ContentType    string `json:"content_type"`
	BytesWritten   int64  `json:"bytes_written"`
	Digest         string `json:"digest,omitempty"`
	DigestVerified bool   `json:"digest_verified"`
	Warning        string `json:"warning,omitempty"`
}

type AttachmentDownloadFailure struct {
	AttachmentID int    `json:"attachment_id"`
	FileName     string `json:"file_name"`
	Error        string `json:"error"`
}

type AttachmentDownloadBatchResult struct {
	WorkPackageID int                         `json:"work_package_id"`
	Directory     string                      `json:"directory"`
	Downloaded    []AttachmentDownloadResult  `json:"downloaded"`
	Failed        []AttachmentDownloadFailure `json:"failed"`
}

type AttachmentDeleteResult struct {
	Attachment Attachment `json:"attachment"`
	Deleted    bool       `json:"deleted"`
}

type AttachmentBatchError struct {
	Failed int
}

func (e *AttachmentBatchError) Error() string {
	return fmt.Sprintf("failed to download %d attachment(s)", e.Failed)
}

type attachmentResource struct {
	Attachment
	downloadHref string
}

func (c *Client) ListWorkPackageAttachments(ctx context.Context, workPackageID int) (*AttachmentCollection, error) {
	var model external.AttachmentsModel
	resp, err := c.apiClient.ListWorkPackageAttachments(ctx, workPackageID)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if err := DecodeResponse(resp, err, &model); err != nil {
		return nil, err
	}

	result := &AttachmentCollection{
		WorkPackageID: workPackageID,
		Total:         model.Total,
		Attachments:   []Attachment{},
	}
	if model.UnderscoreEmbedded.Elements == nil {
		return result, nil
	}
	for _, item := range *model.UnderscoreEmbedded.Elements {
		resource, err := attachmentFromModel(item)
		if err != nil {
			return nil, err
		}
		if resource.WorkPackageID == 0 {
			resource.WorkPackageID = workPackageID
		}
		result.Attachments = append(result.Attachments, resource.Attachment)
	}
	result.Total = len(result.Attachments)
	return result, nil
}

func (c *Client) GetAttachment(ctx context.Context, attachmentID int) (*Attachment, error) {
	resource, err := c.getAttachmentResource(ctx, attachmentID)
	if err != nil {
		return nil, err
	}
	return &resource.Attachment, nil
}

func (c *Client) UploadWorkPackageAttachment(ctx context.Context, input AttachmentUploadInput) (*Attachment, error) {
	file, err := os.Open(input.FilePath)
	if err != nil {
		return nil, fmt.Errorf("open attachment file: %w", err)
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect attachment file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("attachment path is not a regular file: %s", input.FilePath)
	}

	fileName := input.FileName
	if fileName == "" {
		fileName = filepath.Base(input.FilePath)
	}
	fileName = safeAttachmentName(fileName, 0)
	contentType, err := detectAttachmentContentType(file, fileName, input.ContentType)
	if err != nil {
		return nil, err
	}

	reader, contentTypeHeader, waitForWriter := multipartAttachmentBody(file, fileName, input.Description, contentType)
	transferCtx, cancel := c.transferContext(ctx)
	defer cancel()
	requestURL := fmt.Sprintf("%s/api/v3/work_packages/%d/attachments", c.baseURL, input.WorkPackageID)
	req, err := http.NewRequestWithContext(transferCtx, http.MethodPost, requestURL, reader)
	if err != nil {
		_ = reader.CloseWithError(err)
		_ = waitForWriter()
		return nil, fmt.Errorf("create attachment upload request: %w", err)
	}
	req.SetBasicAuth("apikey", c.apiKey)
	req.Header.Set("Content-Type", contentTypeHeader)

	baseURL, parseErr := url.Parse(c.baseURL)
	if parseErr != nil {
		_ = reader.CloseWithError(parseErr)
		_ = waitForWriter()
		return nil, fmt.Errorf("parse OpenProject URL: %w", parseErr)
	}
	resp, requestErr := c.transferHTTPClient(baseURL).Do(req)
	if requestErr != nil {
		_ = reader.CloseWithError(requestErr)
		_ = waitForWriter()
		return nil, fmt.Errorf("upload attachment: %w", requestErr)
	}
	writerErr := waitForWriter()
	if writerErr != nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return nil, fmt.Errorf("stream attachment upload: %w", writerErr)
	}

	var model external.AttachmentModel
	if err := ReadResponse(resp, &model); err != nil {
		return nil, err
	}
	resource, err := attachmentFromModel(model)
	if err != nil {
		return nil, err
	}
	if resource.WorkPackageID == 0 {
		resource.WorkPackageID = input.WorkPackageID
	}
	return &resource.Attachment, nil
}

func (c *Client) DownloadAttachment(ctx context.Context, input AttachmentDownloadInput) (*AttachmentDownloadResult, error) {
	resource, err := c.getAttachmentResource(ctx, input.AttachmentID)
	if err != nil {
		return nil, err
	}
	destination := input.Destination
	if destination == "" {
		destination = safeAttachmentName(resource.FileName, resource.ID)
	}
	return c.downloadAttachmentResource(ctx, resource, destination, input.Overwrite)
}

func (c *Client) DownloadWorkPackageAttachments(ctx context.Context, input AttachmentDownloadAllInput) (*AttachmentDownloadBatchResult, error) {
	resources, err := c.listWorkPackageAttachmentResources(ctx, input.WorkPackageID)
	if err != nil {
		return nil, err
	}
	directory := input.Directory
	if directory == "" {
		directory = fmt.Sprintf("openproject-%d-attachments", input.WorkPackageID)
	}
	absoluteDirectory, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve attachment directory: %w", err)
	}

	destinations, err := planAttachmentDestinations(absoluteDirectory, resources, input.Overwrite)
	if err != nil {
		return nil, err
	}
	if len(resources) > 0 {
		if err := os.MkdirAll(absoluteDirectory, 0o750); err != nil {
			return nil, fmt.Errorf("create attachment directory: %w", err)
		}
	}

	result := &AttachmentDownloadBatchResult{
		WorkPackageID: input.WorkPackageID,
		Directory:     absoluteDirectory,
		Downloaded:    []AttachmentDownloadResult{},
		Failed:        []AttachmentDownloadFailure{},
	}
	for index := range resources {
		downloaded, downloadErr := c.downloadAttachmentResource(ctx, resources[index], destinations[index], input.Overwrite)
		if downloadErr != nil {
			result.Failed = append(result.Failed, AttachmentDownloadFailure{
				AttachmentID: resources[index].ID,
				FileName:     resources[index].FileName,
				Error:        downloadErr.Error(),
			})
			continue
		}
		result.Downloaded = append(result.Downloaded, *downloaded)
	}
	if len(result.Failed) > 0 {
		return result, &AttachmentBatchError{Failed: len(result.Failed)}
	}
	return result, nil
}

func (c *Client) DeleteAttachment(ctx context.Context, attachmentID int) (*AttachmentDeleteResult, error) {
	resource, err := c.getAttachmentResource(ctx, attachmentID)
	if err != nil {
		return nil, err
	}
	resp, err := c.apiClient.DeleteAttachment(ctx, attachmentID)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if err := DecodeResponse(resp, err, nil); err != nil {
		return nil, err
	}
	return &AttachmentDeleteResult{Attachment: resource.Attachment, Deleted: true}, nil
}

func (c *Client) listWorkPackageAttachmentResources(ctx context.Context, workPackageID int) ([]attachmentResource, error) {
	var model external.AttachmentsModel
	resp, err := c.apiClient.ListWorkPackageAttachments(ctx, workPackageID)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if err := DecodeResponse(resp, err, &model); err != nil {
		return nil, err
	}
	if model.UnderscoreEmbedded.Elements == nil {
		return []attachmentResource{}, nil
	}
	resources := make([]attachmentResource, 0, len(*model.UnderscoreEmbedded.Elements))
	for _, item := range *model.UnderscoreEmbedded.Elements {
		resource, err := attachmentFromModel(item)
		if err != nil {
			return nil, err
		}
		if resource.WorkPackageID == 0 {
			resource.WorkPackageID = workPackageID
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (c *Client) getAttachmentResource(ctx context.Context, attachmentID int) (attachmentResource, error) {
	var model external.AttachmentModel
	resp, err := c.apiClient.ViewAttachment(ctx, attachmentID)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if err := DecodeResponse(resp, err, &model); err != nil {
		return attachmentResource{}, err
	}
	resource, err := attachmentFromModel(model)
	if err != nil {
		return attachmentResource{}, err
	}
	if resource.WorkPackageID == 0 {
		return attachmentResource{}, fmt.Errorf("attachment #%d does not belong to a work package", attachmentID)
	}
	return resource, nil
}

func attachmentFromModel(model external.AttachmentModel) (attachmentResource, error) {
	resource := attachmentResource{
		Attachment: Attachment{
			ID:          derefIntValue(model.Id),
			FileName:    model.FileName,
			ContentType: model.ContentType,
			Status:      string(model.Status),
			CreatedAt:   model.CreatedAt,
			Digest: AttachmentDigest{
				Algorithm: model.Digest.Algorithm,
				Hash:      model.Digest.Hash,
			},
		},
	}
	if model.FileSize != nil {
		size := int64(*model.FileSize)
		resource.FileSize = &size
	}
	if model.Description.Raw != nil {
		resource.Description = *model.Description.Raw
	}
	if model.UnderscoreLinks != nil {
		resource.Author = derefString(model.UnderscoreLinks.Author.Title)
		resource.WorkPackageID = workPackageIDFromHref(derefString(model.UnderscoreLinks.Container.Href))
		resource.downloadHref = derefString(model.UnderscoreLinks.DownloadLocation.Href)
	}
	if resource.ID <= 0 {
		return attachmentResource{}, errors.New("attachment response is missing an ID")
	}
	return resource, nil
}

func (c *Client) downloadAttachmentResource(ctx context.Context, resource attachmentResource, destination string, overwrite bool) (*AttachmentDownloadResult, error) {
	if resource.downloadHref == "" {
		return nil, fmt.Errorf("attachment #%d does not provide a download location", resource.ID)
	}
	absoluteDestination, err := filepath.Abs(destination)
	if err != nil {
		return nil, fmt.Errorf("resolve attachment destination: %w", err)
	}
	if err := ensureDestinationAvailable(absoluteDestination, overwrite); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absoluteDestination), 0o750); err != nil {
		return nil, fmt.Errorf("create attachment destination directory: %w", err)
	}

	downloadURL, baseURL, err := resolveAttachmentDownloadURL(c.baseURL, resource.downloadHref)
	if err != nil {
		return nil, err
	}
	transferCtx, cancel := c.transferContext(ctx)
	defer cancel()
	req, err := http.NewRequestWithContext(transferCtx, http.MethodGet, downloadURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create attachment download request: %w", err)
	}
	if sameOrigin(baseURL, downloadURL) {
		req.SetBasicAuth("apikey", c.apiKey)
	}

	resp, err := c.transferHTTPClient(baseURL).Do(req)
	if err != nil {
		return nil, fmt.Errorf("download attachment: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, ReadResponse(resp, nil)
	}
	defer func() { _ = resp.Body.Close() }()

	tempFile, err := os.CreateTemp(filepath.Dir(absoluteDestination), ".openproject-mcp-*.part")
	if err != nil {
		return nil, fmt.Errorf("create attachment temporary file: %w", err)
	}
	tempPath := tempFile.Name()
	committed := false
	defer func() {
		_ = tempFile.Close()
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()

	hasher, digestSupported := attachmentDigestHasher(resource.Digest.Algorithm)
	var writer io.Writer = tempFile
	if digestSupported && resource.Digest.Hash != "" {
		writer = io.MultiWriter(tempFile, hasher)
	}
	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("write attachment: %w", err)
	}
	if resource.FileSize != nil && written != *resource.FileSize {
		return nil, fmt.Errorf("attachment size mismatch: expected %d bytes, wrote %d", *resource.FileSize, written)
	}
	digestVerified := false
	warning := ""
	if resource.Digest.Hash != "" {
		if digestSupported {
			actual := hex.EncodeToString(hasher.Sum(nil))
			if !strings.EqualFold(actual, resource.Digest.Hash) {
				return nil, fmt.Errorf("attachment digest mismatch: expected %s:%s, got %s", resource.Digest.Algorithm, resource.Digest.Hash, actual)
			}
			digestVerified = true
		} else {
			warning = fmt.Sprintf("digest algorithm %q is not supported; file size was verified", resource.Digest.Algorithm)
		}
	}
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("close attachment temporary file: %w", err)
	}
	if err := commitDownloadedFile(tempPath, absoluteDestination, overwrite); err != nil {
		return nil, err
	}
	committed = true

	return &AttachmentDownloadResult{
		AttachmentID:   resource.ID,
		WorkPackageID:  resource.WorkPackageID,
		FileName:       resource.FileName,
		Path:           absoluteDestination,
		ContentType:    resource.ContentType,
		BytesWritten:   written,
		Digest:         formatAttachmentDigest(resource.Digest),
		DigestVerified: digestVerified,
		Warning:        warning,
	}, nil
}

func multipartAttachmentBody(file *os.File, fileName, description, contentType string) (*io.PipeReader, string, func() error) {
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	done := make(chan error, 1)
	go func() {
		metadataHeader := make(textproto.MIMEHeader)
		metadataHeader.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "metadata"}))
		metadataHeader.Set("Content-Type", "application/json")
		metadataPart, err := multipartWriter.CreatePart(metadataHeader)
		if err == nil {
			metadata := map[string]any{"fileName": fileName}
			if description != "" {
				metadata["description"] = map[string]string{"format": "plain", "raw": description}
			}
			err = json.NewEncoder(metadataPart).Encode(metadata)
		}
		if err == nil {
			fileHeader := make(textproto.MIMEHeader)
			fileHeader.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "file", "filename": fileName}))
			fileHeader.Set("Content-Type", contentType)
			var filePart io.Writer
			filePart, err = multipartWriter.CreatePart(fileHeader)
			if err == nil {
				_, err = io.Copy(filePart, file)
			}
		}
		if closeErr := multipartWriter.Close(); err == nil {
			err = closeErr
		}
		_ = writer.CloseWithError(err)
		done <- err
	}()
	return reader, multipartWriter.FormDataContentType(), func() error { return <-done }
}

func detectAttachmentContentType(file *os.File, fileName, override string) (string, error) {
	if override != "" {
		if _, _, err := mime.ParseMediaType(override); err != nil {
			return "", fmt.Errorf("invalid attachment content type: %w", err)
		}
		return override, nil
	}
	if inferred := mime.TypeByExtension(filepath.Ext(fileName)); inferred != "" {
		return inferred, nil
	}
	buffer := make([]byte, 512)
	read, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("inspect attachment content type: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("rewind attachment file: %w", err)
	}
	return http.DetectContentType(buffer[:read]), nil
}

func (c *Client) transferContext(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := c.transferTimeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return context.WithTimeout(ctx, timeout)
}

func (c *Client) transferHTTPClient(baseURL *url.URL) *http.Client {
	client := *c.httpClient
	client.Timeout = c.transferTimeout
	if client.Timeout <= 0 {
		client.Timeout = 5 * time.Minute
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if err := validateDownloadURL(req.URL); err != nil {
			return err
		}
		if baseURL == nil || !sameOrigin(baseURL, req.URL) {
			req.Header.Del("Authorization")
		}
		return nil
	}
	return &client
}

func resolveAttachmentDownloadURL(base, href string) (*url.URL, *url.URL, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, nil, fmt.Errorf("parse OpenProject URL: %w", err)
	}
	reference, err := url.Parse(href)
	if err != nil {
		return nil, nil, fmt.Errorf("parse attachment download URL: %w", err)
	}
	resolved := baseURL.ResolveReference(reference)
	if err := validateDownloadURL(resolved); err != nil {
		return nil, nil, err
	}
	return resolved, baseURL, nil
}

func validateDownloadURL(value *url.URL) error {
	if value.Scheme != "http" && value.Scheme != "https" {
		return fmt.Errorf("unsupported attachment download URL scheme: %s", value.Scheme)
	}
	if value.Host == "" {
		return errors.New("attachment download URL is missing a host")
	}
	if value.User != nil {
		return errors.New("attachment download URL must not contain user information")
	}
	return nil
}

func sameOrigin(left, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) &&
		strings.EqualFold(left.Hostname(), right.Hostname()) &&
		effectivePort(left) == effectivePort(right)
}

func effectivePort(value *url.URL) string {
	if value.Port() != "" {
		return value.Port()
	}
	if strings.EqualFold(value.Scheme, "https") {
		return "443"
	}
	return "80"
}

func workPackageIDFromHref(href string) int {
	parsed, err := url.Parse(href)
	if err != nil {
		return 0
	}
	const marker = "/api/v3/work_packages/"
	index := strings.LastIndex(parsed.Path, marker)
	if index < 0 {
		return 0
	}
	value := strings.Trim(parsed.Path[index+len(marker):], "/")
	if value == "" || strings.Contains(value, "/") {
		return 0
	}
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func safeAttachmentName(value string, attachmentID int) string {
	value = strings.ReplaceAll(value, "\\", "/")
	value = path.Base(value)
	var builder strings.Builder
	for _, character := range value {
		if character < 32 || strings.ContainsRune(`<>:"/\\|?*`, character) {
			builder.WriteRune('_')
			continue
		}
		builder.WriteRune(character)
	}
	value = strings.TrimRight(strings.TrimSpace(builder.String()), ". ")
	if value == "" || value == "." || value == ".." {
		return fallbackAttachmentName(attachmentID)
	}
	stem := strings.ToUpper(strings.TrimSuffix(value, filepath.Ext(value)))
	if isWindowsReservedName(stem) {
		value = "_" + value
	}
	return value
}

func fallbackAttachmentName(attachmentID int) string {
	if attachmentID > 0 {
		return fmt.Sprintf("attachment-%d", attachmentID)
	}
	return "attachment"
}

func isWindowsReservedName(value string) bool {
	switch value {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	if len(value) == 4 && (strings.HasPrefix(value, "COM") || strings.HasPrefix(value, "LPT")) {
		return value[3] >= '1' && value[3] <= '9'
	}
	return false
}

func planAttachmentDestinations(directory string, resources []attachmentResource, overwrite bool) ([]string, error) {
	destinations := make([]string, len(resources))
	seen := map[string]struct{}{}
	for index, resource := range resources {
		name := safeAttachmentName(resource.FileName, resource.ID)
		key := strings.ToLower(name)
		if _, duplicate := seen[key]; duplicate {
			extension := filepath.Ext(name)
			name = strings.TrimSuffix(name, extension) + fmt.Sprintf("-%d", resource.ID) + extension
			key = strings.ToLower(name)
		}
		seen[key] = struct{}{}
		destination := filepath.Join(directory, name)
		if err := ensureDestinationAvailable(destination, overwrite); err != nil {
			return nil, err
		}
		destinations[index] = destination
	}
	return destinations, nil
}

func ensureDestinationAvailable(destination string, overwrite bool) error {
	info, err := os.Stat(destination)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect attachment destination: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("attachment destination is a directory: %s", destination)
	}
	if !overwrite {
		return fmt.Errorf("attachment destination already exists: %s (use --overwrite to replace it)", destination)
	}
	return nil
}

func commitDownloadedFile(tempPath, destination string, overwrite bool) error {
	if !overwrite {
		if err := ensureDestinationAvailable(destination, false); err != nil {
			return err
		}
		if err := os.Rename(tempPath, destination); err != nil {
			return fmt.Errorf("commit attachment file: %w", err)
		}
		return nil
	}

	if _, err := os.Stat(destination); errors.Is(err, os.ErrNotExist) {
		if err := os.Rename(tempPath, destination); err != nil {
			return fmt.Errorf("commit attachment file: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("inspect attachment destination: %w", err)
	}

	backup := destination + ".openproject-mcp.bak"
	if _, err := os.Stat(backup); err == nil {
		return fmt.Errorf("attachment backup path already exists: %s", backup)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect attachment backup path: %w", err)
	}
	if err := os.Rename(destination, backup); err != nil {
		return fmt.Errorf("backup existing attachment file: %w", err)
	}
	if err := os.Rename(tempPath, destination); err != nil {
		_ = os.Rename(backup, destination)
		return fmt.Errorf("commit attachment file: %w", err)
	}
	_ = os.Remove(backup)
	return nil
}

func attachmentDigestHasher(algorithm string) (hash.Hash, bool) {
	switch strings.ToLower(strings.ReplaceAll(algorithm, "-", "")) {
	case "md5":
		return md5.New(), true // #nosec G401 -- integrity compatibility with OpenProject.
	case "sha1":
		return sha1.New(), true // #nosec G401 -- integrity compatibility with OpenProject.
	case "sha256":
		return sha256.New(), true
	case "sha512":
		return sha512.New(), true
	default:
		return nil, false
	}
}

func formatAttachmentDigest(digest AttachmentDigest) string {
	if digest.Algorithm == "" || digest.Hash == "" {
		return ""
	}
	return digest.Algorithm + ":" + digest.Hash
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
