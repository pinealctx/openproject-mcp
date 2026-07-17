// OpenProject MCP Server - Main Entry Point
package main

import (
	"os"
	"runtime/debug"

	"github.com/pinealctx/openproject-mcp/cmd"
)

// Version is set at build time via -ldflags "-X main.Version=x.y.z".
var Version = "dev"

func main() {
	// Set version in cmd package
	cmd.Version = resolvedVersion(Version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func resolvedVersion(injected string) string {
	if injected != "" && injected != "dev" {
		return injected
	}
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	return versionFromBuildInfo(injected, buildInfo.Main.Version)
}

func versionFromBuildInfo(injected, moduleVersion string) string {
	if injected != "" && injected != "dev" {
		return injected
	}
	if moduleVersion != "" && moduleVersion != "(devel)" {
		return moduleVersion
	}
	return "dev"
}
