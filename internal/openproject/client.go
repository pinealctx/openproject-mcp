package openproject

import (
	"bytes"
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
)

// Client is an OpenProject API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClient creates a new OpenProject API client.
func NewClient(cfg *config.Config) *Client {
	// Create HTTP transport with optional proxy
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	// Configure proxy if set
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	return &Client{
		baseURL:    strings.TrimSuffix(cfg.OpenProjectURL, "/"),
		apiKey:     cfg.APIKey,
		httpClient: httpClient,
		logger:     slog.Default(),
	}
}

// NewClientDirect creates a client directly from credentials without a Config struct.
// Used by the HTTP/SSE server to create per-request clients from headers.
func NewClientDirect(baseURL, apiKey string, timeout time.Duration) *Client {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
		logger: slog.Default(),
	}
}

// SetLogger sets the client logger.
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logger
}

// doRequest performs an HTTP request to the OpenProject API.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	fullURL := c.baseURL + "/api/v3" + path
	c.logger.Debug("making request", "method", method, "url", fullURL)

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use Basic Auth with apikey as username (OpenProject standard)
	// Format: Authorization: Basic base64(apikey:<token>)
	req.SetBasicAuth("apikey", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug("received response", "status", resp.StatusCode, "body", string(respBody))

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			apiErr.Message = string(respBody)
			apiErr.StatusCode = resp.StatusCode
		}
		return &apiErr
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPatch, path, body, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// TestConnection tests the connection to OpenProject.
func (c *Client) TestConnection(ctx context.Context) (*User, error) {
	var user User
	if err := c.Get(ctx, "/users/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetCurrentUser returns the current authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	return c.TestConnection(ctx)
}

// GetAPIRoot retrieves the API root document (version info, available links).
func (c *Client) GetAPIRoot(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.Get(ctx, "", &result); err != nil {
		return nil, err
	}
	return result, nil
}
