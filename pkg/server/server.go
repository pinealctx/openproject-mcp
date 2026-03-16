// Package server provides the MCP server implementation for OpenProject.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/config"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	"github.com/pinealctx/openproject-mcp/internal/tools"
)

// clientContextKey is the context key for the per-request OpenProject client.
type clientContextKey struct{}

// Server represents the MCP server for OpenProject.
type Server struct {
	config      *config.Config
	client      *openproject.Client // default client; nil when http/sse without pre-configured credentials
	logger      *slog.Logger
	version     string
	clientCache sync.Map // caches openproject.Client instances keyed by "url\x00apiKey"
}

// New creates a new MCP server.
func New(cfg *config.Config, version string) (*Server, error) {
	var client *openproject.Client
	if cfg.IsConfigured() {
		client = openproject.NewClient(cfg)
	}
	return &Server{
		config:  cfg,
		client:  client,
		logger:  slog.Default(),
		version: version,
	}, nil
}

// cachedClient returns a cached Client for the given credentials, creating one if needed.
func (s *Server) cachedClient(baseURL, apiKey string) *openproject.Client {
	key := baseURL + "\x00" + apiKey
	if v, ok := s.clientCache.Load(key); ok {
		return v.(*openproject.Client)
	}
	c := openproject.NewClientDirect(baseURL, apiKey, 30*time.Second)
	s.clientCache.Store(key, c)
	return c
}

// withClientMiddleware is an HTTP middleware that resolves the OpenProject client
// for each request. Resolution order:
//  1. X-OpenProject-URL + X-OpenProject-API-Key headers (per-request, multi-tenant)
//  2. Server-level default client (pre-configured via env vars)
//  3. Neither available → HTTP 401
func (s *Server) withClientMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var client *openproject.Client
		opURL := req.Header.Get("X-OpenProject-URL")
		apiKey := req.Header.Get("X-OpenProject-API-Key")
		if opURL != "" && apiKey != "" {
			client = s.cachedClient(opURL, apiKey)
		} else if s.client != nil {
			client = s.client
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"OpenProject credentials required: provide X-OpenProject-URL and X-OpenProject-API-Key headers"}`))
			return
		}
		ctx := context.WithValue(req.Context(), clientContextKey{}, client)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// Run starts the MCP server with the configured transport.
func (s *Server) Run(ctx context.Context) error {
	switch s.config.Transport {
	case "stdio":
		return s.runStdio(ctx)
	case "sse":
		return s.runSSE(ctx)
	case "http":
		return s.runHTTP(ctx)
	default:
		return fmt.Errorf("unknown transport: %s", s.config.Transport)
	}
}

// createMCPServer creates and configures the MCP server with all tools for the given client.
func (s *Server) createMCPServer(client *openproject.Client) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "openproject-mcp",
		Version: s.version,
	}, nil)

	registry := tools.NewRegistry(client)
	registry.RegisterAllTools(server)

	return server
}

// runStdio runs the server with stdio transport.
func (s *Server) runStdio(ctx context.Context) error {
	s.logger.Info("Starting MCP server with stdio transport")

	// For stdio, s.client is guaranteed non-nil (validated at startup).
	server := s.createMCPServer(s.client)

	// Run with stdio transport
	transport := &mcp.StdioTransport{}
	session, err := server.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Close session
	return session.Close()
}

// runSSE runs the server with SSE transport.
// Each SSE connection gets its own openproject.Client resolved via withClientMiddleware.
func (s *Server) runSSE(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.logger.Info("Starting MCP server with SSE transport", "address", addr,
		"default_configured", s.client != nil)

	sseHandler := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
		client := req.Context().Value(clientContextKey{}).(*openproject.Client)
		return s.createMCPServer(client)
	})

	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.withClientMiddleware(sseHandler),
	}

	// Handle shutdown
	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down SSE server")
		_ = httpServer.Shutdown(context.Background())
	}()

	return httpServer.ListenAndServe()
}

// runHTTP runs the server with streamable HTTP transport.
// Each request gets its own openproject.Client resolved via withClientMiddleware.
func (s *Server) runHTTP(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.logger.Info("Starting MCP server with HTTP transport", "address", addr,
		"default_configured", s.client != nil)

	httpHandler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		client := req.Context().Value(clientContextKey{}).(*openproject.Client)
		return s.createMCPServer(client)
	}, nil)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.withClientMiddleware(httpHandler),
	}

	// Handle shutdown
	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down HTTP server")
		_ = httpServer.Shutdown(context.Background())
	}()

	return httpServer.ListenAndServe()
}

// TestConnection tests the connection to OpenProject using the default server client.
func (s *Server) TestConnection(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("no default OpenProject client configured")
	}
	_, err := s.client.TestConnection(ctx)
	return err
}
