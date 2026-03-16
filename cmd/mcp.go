package cmd

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/pinealctx/openproject-mcp/internal/config"
	"github.com/pinealctx/openproject-mcp/pkg/server"
)

var (
	mcpTransport string
	mcpPort      int
)

// mcpCmd represents the mcp command.
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server",
	Long: `Start the MCP (Model Context Protocol) server for AI assistants.

The MCP server exposes 48+ tools for AI assistants (like Claude) to interact
with OpenProject. This enables AI-powered project management workflows.

TRANSPORT MODES:

  stdio (default)
    Communication via stdin/stdout. Used by MCP clients that spawn the server
    as a subprocess. This is the most common mode for desktop applications.
    Compatible with: Claude Desktop, Cursor, Continue, Zed, etc.

  sse (Server-Sent Events)
    HTTP-based unidirectional server push. For web-based MCP clients.
    Requires --port flag to specify listening port.

  http (Streamable HTTP)
    Standard HTTP transport for web-based clients.
    Requires --port flag to specify listening port.

AUTHENTICATION:

  Required environment variables:
    OPENPROJECT_URL     Your OpenProject instance URL (e.g., https://openproject.example.com)
    OPENPROJECT_API_KEY Your API token (generate in OpenProject -> My Account -> API token)

  Alternatively, use global flags:
    --url       OpenProject URL
    --api-key   API key

STDIO MODE USAGE:

  The server reads JSON-RPC messages from stdin and writes responses to stdout.
  All logging is suppressed to avoid protocol interference.

  Direct invocation (for testing):
    export OPENPROJECT_URL="https://your-instance.openproject.com"
    export OPENPROJECT_API_KEY="your-api-token"
    openproject-mcp mcp

  With Claude Desktop, add to claude_desktop_config.json:
    {
      "mcpServers": {
        "openproject": {
          "command": "/path/to/openproject-mcp",
          "env": {
            "OPENPROJECT_URL": "https://your-instance.openproject.com",
            "OPENPROJECT_API_KEY": "your-api-token"
          }
        }
      }
    }

  With Cursor, add to .cursor/mcp.json:
    {
      "mcpServers": {
        "openproject": {
          "command": "/path/to/openproject-mcp",
          "env": {
            "OPENPROJECT_URL": "https://your-instance.openproject.com",
            "OPENPROJECT_API_KEY": "your-api-token"
          }
        }
      }
    }

EXAMPLES:

  # Start in stdio mode (default)
  openproject-mcp mcp

  # Explicit stdio mode
  openproject-mcp mcp -t stdio

  # SSE mode on port 3000
  openproject-mcp mcp -t sse -p 3000

  # HTTP mode on port 8080
  openproject-mcp mcp -t http -p 8080`,
	Run: func(cmd *cobra.Command, args []string) {
		runMCPServer(cmd)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVarP(&mcpTransport, "transport", "t", "", "Transport type: stdio, sse, or http")
	mcpCmd.Flags().IntVarP(&mcpPort, "port", "p", 0, "Port for SSE/HTTP transport")
}

// runMCPServer starts the MCP server with the configured transport.
func runMCPServer(cmd *cobra.Command) {
	cfg := config.Load()

	// Apply flags
	if mcpTransport != "" {
		cfg.Transport = mcpTransport
	} else if cmd.CalledAs() != "mcp" {
		// If called as root (no subcommand), default to stdio
		cfg.Transport = "stdio"
	}

	if mcpPort > 0 {
		cfg.Port = mcpPort
	}

	// Setup logging for MCP mode
	setupMCPLogging(cfg)

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

// setupMCPLogging configures logging for MCP server mode.
func setupMCPLogging(cfg *config.Config) {
	// In stdio mode stdout is used for the MCP protocol; discard all logs
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
