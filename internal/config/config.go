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
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.OpenProjectURL == "" {
		return errors.New("OPENPROJECT_URL is required")
	}
	if c.APIKey == "" {
		return errors.New("OPENPROJECT_API_KEY is required")
	}

	// Validate URL format
	parsedURL, err := url.Parse(c.OpenProjectURL)
	if err != nil {
		return fmt.Errorf("invalid OPENPROJECT_URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("OPENPROJECT_URL must start with http:// or https://")
	}

	// Validate transport
	if c.Transport != "stdio" && c.Transport != "sse" && c.Transport != "http" {
		return fmt.Errorf("invalid TRANSPORT: must be 'stdio', 'sse', or 'http', got '%s'", c.Transport)
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

	return nil
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
