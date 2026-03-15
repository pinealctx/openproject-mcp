// Package tools provides MCP tools for OpenProject.
package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

// Registry holds all tool registrations.
type Registry struct {
	client *openproject.Client
}

// NewRegistry creates a new tool registry.
func NewRegistry(client *openproject.Client) *Registry {
	return &Registry{
		client: client,
	}
}

// RegisterAllTools registers all OpenProject tools with the MCP server.
func (r *Registry) RegisterAllTools(server *mcp.Server) {
	// Connection tools
	r.registerConnectionTools(server)

	// Project tools
	r.registerProjectTools(server)

	// Work package tools
	r.registerWorkPackageTools(server)

	// User tools
	r.registerUserTools(server)

	// Membership tools
	r.registerMembershipTools(server)

	// Time entry tools
	r.registerTimeEntryTools(server)

	// Version tools
	r.registerVersionTools(server)

	// Relation tools
	r.registerRelationTools(server)
}
