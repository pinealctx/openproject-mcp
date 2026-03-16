// OpenProject MCP Server - Main Entry Point
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pinealctx/openproject-mcp/internal/config"
	"github.com/pinealctx/openproject-mcp/pkg/server"
)

// Version is set at build time via -ldflags "-X main.Version=x.y.z".
var Version = "dev"

func main() {
	transport := flag.String("transport", "", "Transport type: stdio, sse, or http")
	port := flag.Int("port", 0, "Port for SSE/HTTP transport")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("openproject-mcp version %s\n", Version)
		os.Exit(0)
	}

	cfg := config.Load()

	if *transport != "" {
		cfg.Transport = *transport
	}
	if *port > 0 {
		cfg.Port = *port
	}

	setupLogging(cfg)

	if err := cfg.Validate(); err != nil {
		slog.Error("Configuration error", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting OpenProject MCP Server",
		"version", Version,
		"transport", cfg.Transport,
		"openproject_url", cfg.OpenProjectURL,
	)

	srv, err := server.New(cfg, Version)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal")
		cancel()
	}()

	if err := srv.Run(ctx); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped")
}

// setupLogging configures the logger based on the configuration.
func setupLogging(cfg *config.Config) {
	// In stdio mode stdout is used for the MCP protocol; discard all logs to
	// prevent any output from interfering with the client.
	if cfg.Transport == "stdio" {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		return
	}

	level, err := config.ParseLogLevel(cfg.LogLevel)
	if err != nil {
		level = 0
	}

	opts := &slog.HandlerOptions{Level: slog.Level(level)}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
}
