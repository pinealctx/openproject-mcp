// OpenProject MCP Server - Main Entry Point
package main

import (
	"os"

	"github.com/pinealctx/openproject-mcp/cmd"
)

// Version is set at build time via -ldflags "-X main.Version=x.y.z".
var Version = "dev"

func main() {
	// Set version in cmd package
	cmd.Version = Version
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
