// Package server provides the MCP server implementation for OpenProject.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/config"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	"github.com/pinealctx/openproject-mcp/internal/tools"
)

// Server represents the MCP server for OpenProject.
type Server struct {
	config  *config.Config
	client  *openproject.Client
	logger  *slog.Logger
	version string
}

// New creates a new MCP server.
func New(cfg *config.Config, version string) (*Server, error) {
	// Create OpenProject client
	client := openproject.NewClient(cfg)

	return &Server{
		config:  cfg,
		client:  client,
		logger:  slog.Default(),
		version: version,
	}, nil
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

// createMCPServer creates and configures the MCP server with all tools.
func (s *Server) createMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "openproject-mcp",
		Version: s.version,
	}, nil)

	// Register all tools
	registry := tools.NewRegistry(s.client)
	registry.RegisterAllTools(server)

	return server
}

// runStdio runs the server with stdio transport.
func (s *Server) runStdio(ctx context.Context) error {
	s.logger.Info("Starting MCP server with stdio transport")

	server := s.createMCPServer()

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
func (s *Server) runSSE(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.logger.Info("Starting MCP server with SSE transport", "address", addr)

	// Create SSE handler
	handler := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
		return s.createMCPServer()
	})

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
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
func (s *Server) runHTTP(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.logger.Info("Starting MCP server with HTTP transport", "address", addr)

	// Create streamable HTTP handler
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return s.createMCPServer()
	}, nil)

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Handle shutdown
	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down HTTP server")
		_ = httpServer.Shutdown(context.Background())
	}()

	return httpServer.ListenAndServe()
}

// TestConnection tests the connection to OpenProject.
func (s *Server) TestConnection(ctx context.Context) error {
	_, err := s.client.TestConnection(ctx)
	return err
}
