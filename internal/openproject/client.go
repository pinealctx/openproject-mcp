// Package openproject provides a client adapter for the OpenProject API.
// It wraps the generated client from github.com/pinealctx/openproject.
package openproject

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pinealctx/openproject-mcp/internal/config"

	external "github.com/pinealctx/openproject"
)

// Client wraps the generated OpenProject API client with authentication
// and convenience methods.
type Client struct {
	apiClient *external.Client
	baseURL   string
	apiKey    string
	logger    *slog.Logger
}

// NewClient creates a new OpenProject API client from config.
func NewClient(cfg *config.Config) *Client {
	serverURL := strings.TrimSuffix(cfg.OpenProjectURL, "/")
	client, err := external.NewClient(serverURL,
		external.WithHTTPClient(newHTTPClient(cfg)),
		external.WithRequestEditorFn(basicAuthEditor(cfg.APIKey)),
	)
	if err != nil {
		// Should not happen with valid URL
		return nil
	}
	return &Client{
		apiClient: client,
		baseURL:   serverURL,
		apiKey:    cfg.APIKey,
		logger:    slog.Default(),
	}
}

// NewClientDirect creates a client from raw credentials.
// Used by HTTP/SSE server for per-request client creation.
func NewClientDirect(baseURL, apiKey string, timeout time.Duration) *Client {
	serverURL := strings.TrimSuffix(baseURL, "/")
	cfg := &config.Config{Timeout: timeout}
	client, err := external.NewClient(serverURL,
		external.WithHTTPClient(newHTTPClient(cfg)),
		external.WithRequestEditorFn(basicAuthEditor(apiKey)),
	)
	if err != nil {
		return nil
	}
	return &Client{
		apiClient: client,
		baseURL:   serverURL,
		apiKey:    apiKey,
		logger:    slog.Default(),
	}
}

// SetLogger sets the client logger.
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logger
}

// APIClient returns the underlying generated API client for direct method calls.
func (c *Client) APIClient() *external.Client {
	return c.apiClient
}

// ReadResponse reads and unmarshals an HTTP response body into target.
// Returns an APIError for non-2xx status codes.
func ReadResponse(resp *http.Response, target any) error {
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		var apiErr external.ErrorResponse
		if json.Unmarshal(body, &apiErr) != nil {
			apiErr.Message = string(body)
		}
		return &APIError{
			StatusCode:    resp.StatusCode,
			Message:       apiErr.Message,
			ErrorID:       apiErr.ErrorIdentifier,
			ResponseBody:  string(body),
		}
	}
	if target != nil && len(body) > 0 {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return nil
}

// ReadResponseRaw reads the response body as bytes without unmarshaling.
func ReadResponseRaw(resp *http.Response) ([]byte, error) {
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}

// ReadResponseRawTo reads the response and unmarshals into target, handling non-2xx as error.
// This is similar to ReadResponse but works with raw map[string]interface{} types.
func ReadResponseRawTo(resp *http.Response, target any) error {
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		var apiErr external.ErrorResponse
		if json.Unmarshal(body, &apiErr) != nil {
			apiErr.Message = string(body)
		}
		return &APIError{
			StatusCode:   resp.StatusCode,
			Message:      apiErr.Message,
			ErrorID:      apiErr.ErrorIdentifier,
			ResponseBody: string(body),
		}
	}
	if target != nil && len(body) > 0 {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return nil
}

// Get performs a raw GET request to /api/v3/{path}.
// Prefer using c.APIClient() methods directly.
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// Post performs a raw POST request to /api/v3/{path}.
func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// Patch performs a raw PATCH request to /api/v3/{path}.
func (c *Client) Patch(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPatch, path, body, result)
}

// Delete performs a raw DELETE request to /api/v3/{path}.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// TestConnection tests the connection by fetching the current user.
func (c *Client) TestConnection(ctx context.Context) (*external.UserModel, error) {
	resp, err := c.apiClient.ListUsers(ctx, &external.ListUsersParams{
		Filters:  ptr(`[{"status":{"operator":"!","values":["*"]}}]`),
		PageSize: ptr(1),
	})
	if err != nil {
		return nil, err
	}
	// Just check we can reach the API; close body immediately.
	_ = resp.Body.Close()
	return c.GetCurrentUser(ctx)
}

// GetCurrentUser returns the current authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*external.UserModel, error) {
	var user external.UserModel
	if err := c.Get(ctx, "/users/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAPIRoot retrieves the API root document.
func (c *Client) GetAPIRoot(ctx context.Context) (*external.RootModel, error) {
	resp, err := c.apiClient.ViewRoot(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var root external.RootModel
	if err := ReadResponse(resp, &root); err != nil {
		return nil, err
	}
	return &root, nil
}

// doRequest is a fallback for paths not covered by the generated client.
func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	reqURL := c.baseURL + "/api/v3" + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth("apikey", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.apiClient.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return ReadResponse(resp, result)
}

// --- helpers ---

func basicAuthEditor(apiKey string) external.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.SetBasicAuth("apikey", apiKey)
		req.Header.Set("Content-Type", "application/json")
		return nil
	}
}

func newHTTPClient(cfg *config.Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	if cfg.ProxyURL != "" {
		if proxyURL, err := url.Parse(cfg.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}

func ptr[T any](v T) *T { return &v }

// APIError represents an error from the OpenProject API.
type APIError struct {
	StatusCode   int
	Message      string
	ErrorID      string
	ResponseBody string
}

func (e *APIError) Error() string {
	return e.Message
}
