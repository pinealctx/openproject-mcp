// Package config provides configuration management for the OpenProject MCP server.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the OpenProject MCP server.
type Config struct {
	// OpenProject connection settings
	OpenProjectURL string
	APIKey         string

	// Proxy settings
	ProxyURL string

	// Logging settings
	LogLevel string

	// Server settings
	Transport string // "stdio", "sse", or "http"
	Port      int    // Port for SSE/HTTP transport

	// HTTP client settings
	Timeout time.Duration

	// Tool mode settings
	ToolMode      string // "default", "full", "custom"
	EnabledTools  string // comma-separated tool names for custom mode
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		OpenProjectURL: getEnv("OPENPROJECT_URL", ""),
		APIKey:         getEnv("OPENPROJECT_API_KEY", ""),
		ProxyURL:       getEnv("OPENPROJECT_PROXY", getEnv("HTTPS_PROXY", getEnv("HTTP_PROXY", ""))),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		Transport:      getEnv("TRANSPORT", "stdio"),
		Port:           getEnvInt("PORT", 8080),
		Timeout:        30 * time.Second,
		ToolMode:       getEnv("TOOL_MODE", "default"),
		EnabledTools:   getEnv("ENABLED_TOOLS", ""),
	}
}

// Validate checks if the configuration is valid.
// For stdio transport, OpenProject credentials are required at startup.
// For http/sse transport, credentials are optional — clients may supply them
// per-request via X-OpenProject-URL and X-OpenProject-API-Key headers.
func (c *Config) Validate() error {
	// Validate transport first — credential requirements depend on it.
	if c.Transport != "stdio" && c.Transport != "sse" && c.Transport != "http" {
		return fmt.Errorf("invalid TRANSPORT: must be 'stdio', 'sse', or 'http', got '%s'", c.Transport)
	}

	// stdio is single-tenant; credentials must be embedded at startup.
	if c.Transport == "stdio" {
		if c.OpenProjectURL == "" {
			return errors.New("OPENPROJECT_URL is required for stdio transport")
		}
		if c.APIKey == "" {
			return errors.New("OPENPROJECT_API_KEY is required for stdio transport")
		}
	}

	// Validate URL format when provided (applies to all transport modes).
	if c.OpenProjectURL != "" {
		parsedURL, err := url.Parse(c.OpenProjectURL)
		if err != nil {
			return fmt.Errorf("invalid OPENPROJECT_URL: %w", err)
		}
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return errors.New("OPENPROJECT_URL must start with http:// or https://")
		}
	}

	// Validate proxy URL if provided
	if c.ProxyURL != "" {
		proxyURL, err := url.Parse(c.ProxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
		if proxyURL.Scheme != "http" && proxyURL.Scheme != "https" && proxyURL.Scheme != "socks5" {
			return errors.New("proxy URL must use http, https, or socks5 scheme")
		}
	}

	if c.ToolMode != "default" && c.ToolMode != "full" && c.ToolMode != "custom" {
		return fmt.Errorf("invalid TOOL_MODE: must be 'default', 'full', or 'custom', got '%s'", c.ToolMode)
	}
	if c.ToolMode == "custom" && c.EnabledTools == "" {
		return errors.New("ENABLED_TOOLS is required when TOOL_MODE is 'custom'")
	}

	return nil
}

// IsConfigured reports whether server-level OpenProject credentials are set.
// When false in http/sse mode, clients must supply credentials via request headers.
func (c *Config) IsConfigured() bool {
	return c.OpenProjectURL != "" && c.APIKey != ""
}

// GetProxyURL returns the proxy URL if configured.
func (c *Config) GetProxyURL() (*url.URL, error) {
	if c.ProxyURL == "" {
		return nil, nil
	}
	return url.Parse(c.ProxyURL)
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// ParseLogLevel converts a string log level to slog.Level.
func ParseLogLevel(level string) (int, error) {
	switch strings.ToLower(level) {
	case "debug":
		return -4, nil // slog.LevelDebug
	case "info":
		return 0, nil // slog.LevelInfo
	case "warn", "warning":
		return 4, nil // slog.LevelWarn
	case "error":
		return 8, nil // slog.LevelError
	default:
		return 0, fmt.Errorf("unknown log level: %s", level)
	}
}
