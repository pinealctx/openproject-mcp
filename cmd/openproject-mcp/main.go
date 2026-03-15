// OpenProject MCP Server - Main Entry Point
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pinealctx/openproject-mcp/internal/config"
	"github.com/pinealctx/openproject-mcp/pkg/server"
)

// Version is set at build time.
var Version = "dev"

func main() {
	// Parse command line flags
	transport := flag.String("transport", "", "Transport type: stdio, sse, or http")
	port := flag.Int("port", 0, "Port for SSE/HTTP transport")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("openproject-mcp version %s\n", Version)
		os.Exit(0)
	}

	// Load configuration
	cfg := config.Load()

	// Override with command line flags if provided
	if *transport != "" {
		cfg.Transport = *transport
	}
	if *port > 0 {
		cfg.Port = *port
	}

	// Setup logging
	setupLogging(cfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		slog.Error("Configuration error", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting OpenProject MCP Server",
		"version", Version,
		"transport", cfg.Transport,
		"openproject_url", cfg.OpenProjectURL,
	)

	// Create server
	srv, err := server.New(cfg, Version)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal")
		cancel()
	}()

	// Run server
	if err := srv.Run(ctx); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped")
}

// setupLogging configures the logger based on the configuration.
func setupLogging(cfg *config.Config) {
	level, err := config.ParseLogLevel(cfg.LogLevel)
	if err != nil {
		level = 0 // default to info
	}

	opts := &slog.HandlerOptions{
		Level: slog.Level(level),
	}

	// Use JSON handler for stdio transport, text for others
	var handler slog.Handler
	if cfg.Transport == "stdio" {
		// For stdio, log to stderr to avoid interfering with MCP protocol
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}
