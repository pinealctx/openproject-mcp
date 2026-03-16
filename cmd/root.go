// Package cmd provides the CLI commands for openproject-mcp.
package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/pinealctx/openproject-mcp/internal/config"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via -ldflags.
	Version = "dev"

	// Global flags
	flagURL     string
	flagAPIKey  string
	flagOutput  string
	flagVerbose bool

	// Global client and context
	client *openproject.Client
	ctx    context.Context
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "openproject-mcp",
	Short: "OpenProject CLI and MCP Server",
	Long: `OpenProject MCP provides both a CLI for direct interaction with OpenProject
and an MCP server for AI assistants.

This tool can be used in two modes:

1. CLI Mode: Direct command-line interaction with OpenProject API
   - Manage projects, work packages, users, time entries, etc.
   - Search and query OpenProject data
   - Scriptable for automation and CI/CD integration

2. MCP Server Mode: Model Context Protocol server for AI assistants
   - Exposes 48+ tools for AI assistants like Claude
   - Supports stdio, SSE, and HTTP transports
   - Enables AI-powered project management workflows

Authentication:
  Set OPENPROJECT_URL and OPENPROJECT_API_KEY environment variables,
  or use --url and --api-key flags.

Examples:
  # Start MCP server (default)
  openproject-mcp

  # Start MCP server with SSE transport
  openproject-mcp mcp -t sse -p 3000

  # List all projects
  openproject-mcp project list

  # Get project details
  openproject-mcp project get 42

  # Create a new work package
  openproject-mcp wp create -p 42 -s "Implement feature X"

  # List work packages in a project
  openproject-mcp wp list -p 42

  # Search for items
  openproject-mcp search "bug report"

  # Output in JSON format
  openproject-mcp project list -o json

When run without a subcommand, it starts the MCP server in stdio mode.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip client initialization for MCP command (it handles its own setup)
		if cmd.Name() == "mcp" || (cmd.Parent() != nil && cmd.Parent().Name() == "mcp") {
			return nil
		}

		// Load config from environment
		cfg := config.Load()

		// Override with flags if provided
		if flagURL != "" {
			cfg.OpenProjectURL = flagURL
		}
		if flagAPIKey != "" {
			cfg.APIKey = flagAPIKey
		}

		// Validate credentials
		if cfg.OpenProjectURL == "" {
			return fmt.Errorf("OpenProject URL is required: set OPENPROJECT_URL env or use --url flag")
		}
		if cfg.APIKey == "" {
			return fmt.Errorf("API key is required: set OPENPROJECT_API_KEY env or use --api-key flag")
		}

		// Setup logging
		setupLogging(cfg, cmd)

		// Create client
		client = openproject.NewClient(cfg)
		ctx = context.Background()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: start MCP server
		runMCPServer(cmd)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "OpenProject URL (or OPENPROJECT_URL env)")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key (or OPENPROJECT_API_KEY env)")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "text", "Output format: text, json")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose output")

	// Set version
	rootCmd.Version = Version
}

// setupLogging configures logging based on the configuration.
func setupLogging(cfg *config.Config, _ *cobra.Command) {
	// In CLI mode, suppress logs unless verbose
	if !flagVerbose {
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

// getClient returns the global client, creating one if needed.
func getClient() *openproject.Client {
	if client == nil {
		cfg := config.Load()
		if flagURL != "" {
			cfg.OpenProjectURL = flagURL
		}
		if flagAPIKey != "" {
			cfg.APIKey = flagAPIKey
		}
		cfg.Timeout = 30 * time.Second
		client = openproject.NewClient(cfg)
	}
	return client
}

// getContext returns the global context.
func getContext() context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}
