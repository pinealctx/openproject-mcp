// Package tools provides MCP tools for OpenProject.
package tools

import (
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

// Registry holds all tool registrations.
type Registry struct {
	client *openproject.Client
}

// schemaProps is a convenience alias for building property maps in newSchema.
type schemaProps = map[string]*jsonschema.Schema

// ---------- Schema helpers ----------

// noSchema is used for tools that take no parameters.
var noSchema = &jsonschema.Schema{Type: "object"}

// schemaStr returns a string property schema with the given description.
func schemaStr(desc string) *jsonschema.Schema {
	return &jsonschema.Schema{Type: "string", Description: desc}
}

// schemaInt returns an integer property schema with the given description.
func schemaInt(desc string) *jsonschema.Schema {
	return &jsonschema.Schema{Type: "integer", Description: desc}
}

// schemaBool returns a boolean property schema with the given description.
func schemaBool(desc string) *jsonschema.Schema {
	return &jsonschema.Schema{Type: "boolean", Description: desc}
}

// schemaEnum returns a string property schema restricted to the given values.
func schemaEnum(desc string, values ...string) *jsonschema.Schema {
	enums := make([]any, len(values))
	for i, v := range values {
		enums[i] = v
	}
	return &jsonschema.Schema{Type: "string", Description: desc, Enum: enums}
}

// newSchema builds an object schema with the supplied property map and optional
// required field names.
func newSchema(props map[string]*jsonschema.Schema, required ...string) *jsonschema.Schema {
	s := &jsonschema.Schema{
		Type:       "object",
		Properties: props,
	}
	if len(required) > 0 {
		s.Required = required
	}
	return s
}

// ---------- Tool registration ----------

// addTool registers a tool with the given JSON schema.
func addTool(server *mcp.Server, name, description string, schema *jsonschema.Schema, handler mcp.ToolHandler) {
	server.AddTool(&mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: schema,
	}, handler)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// NewRegistry creates a new tool registry.
func NewRegistry(client *openproject.Client) *Registry {
	return &Registry{
		client: client,
	}
}

// RegisterAllTools registers all OpenProject tools with the MCP server.
func (r *Registry) RegisterAllTools(server *mcp.Server) {
	// Connection / utility tools
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

	// Search tools
	r.registerSearchTools(server)

	// Board / Grid tools
	r.registerBoardTools(server)

	// Notification tools
	r.registerNotificationTools(server)
}
