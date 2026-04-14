// Package tools provides MCP tools for OpenProject.
package tools

import (
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

// ToolGroup defines a named group of tools with a default-on flag.
type ToolGroup struct {
	IsDefault bool
	Tools     []string
}

// toolGroups defines all tool groups and their member tool names.
var toolGroups = map[string]ToolGroup{
	// Default mode groups
	"connection": {IsDefault: true, Tools: []string{"test_connection", "check_permissions", "get_current_user", "get_api_info"}},
	"project":   {IsDefault: true, Tools: []string{"list_projects", "get_project", "create_project", "update_project", "delete_project"}},
	"work_package": {IsDefault: true, Tools: []string{
		"list_work_packages", "get_work_package", "create_work_package",
		"update_work_package", "delete_work_package",
		"list_types", "list_statuses", "list_priorities", "list_available_assignees",
	}},
	"user":    {IsDefault: true, Tools: []string{"list_users", "get_user"}},
	"version": {IsDefault: true, Tools: []string{"list_versions", "create_version", "update_version", "delete_version"}},
	"search":  {IsDefault: true, Tools: []string{"search"}},
	// Full-only mode groups
	"membership": {IsDefault: false, Tools: []string{
		"list_memberships", "get_membership", "create_membership",
		"update_membership", "delete_membership", "list_project_members",
		"list_roles", "get_role",
	}},
	"relation": {IsDefault: false, Tools: []string{
		"set_work_package_parent", "remove_work_package_parent",
		"list_work_package_children",
		"create_work_package_relation", "list_work_package_relations",
		"get_work_package_relation", "update_work_package_relation",
		"delete_work_package_relation",
	}},
	"notification": {IsDefault: false, Tools: []string{
		"list_notifications", "mark_notification_read", "mark_all_notifications_read",
	}},
	"comment": {IsDefault: false, Tools: []string{
		"list_work_package_activities", "create_work_package_comment",
	}},
	"watcher": {IsDefault: false, Tools: []string{
		"list_work_package_watchers", "add_work_package_watcher", "remove_work_package_watcher",
	}},
	"group": {IsDefault: false, Tools: []string{
		"list_groups", "get_group", "create_group", "update_group", "delete_group",
	}},
	"document": {IsDefault: false, Tools: []string{
		"list_documents", "get_document", "update_document",
	}},
	"query": {IsDefault: false, Tools: []string{
		"list_queries", "get_query",
	}},
	"placeholder": {IsDefault: false, Tools: []string{
		"list_placeholder_users", "get_placeholder_user", "create_placeholder_user",
	}},
	"configuration": {IsDefault: false, Tools: []string{
		"view_configuration",
	}},
}

// Registry holds all tool registrations.
type Registry struct {
	client        *openproject.Client
	toolMode      string // "default", "full", "custom"
	enabledTools  map[string]bool
}

// schemaProps is a convenience alias for building property maps in newSchema.
type schemaProps = map[string]*jsonschema.Schema

// ---------- Schema helpers ----------

var noSchema = &jsonschema.Schema{Type: "object"}

func schemaStr(desc string) *jsonschema.Schema {
	return &jsonschema.Schema{Type: "string", Description: desc}
}

func schemaInt(desc string) *jsonschema.Schema {
	return &jsonschema.Schema{Type: "integer", Description: desc}
}

func schemaBool(desc string) *jsonschema.Schema {
	return &jsonschema.Schema{Type: "boolean", Description: desc}
}

func schemaEnum(desc string, values ...string) *jsonschema.Schema {
	enums := make([]any, len(values))
	for i, v := range values {
		enums[i] = v
	}
	return &jsonschema.Schema{Type: "string", Description: desc, Enum: enums}
}

func newSchema(props map[string]*jsonschema.Schema, required ...string) *jsonschema.Schema {
	s := &jsonschema.Schema{Type: "object", Properties: props}
	if len(required) > 0 {
		s.Required = required
	}
	return s
}

func addTool(server *mcp.Server, name, description string, schema *jsonschema.Schema, handler mcp.ToolHandler) {
	server.AddTool(&mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: schema,
	}, handler)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}



// NewRegistry creates a new tool registry.
func NewRegistry(client *openproject.Client) *Registry {
	return &Registry{
		client:   client,
		toolMode: "full",
	}
}

// NewRegistryWithMode creates a registry with the specified tool mode.
func NewRegistryWithMode(client *openproject.Client, toolMode, enabledTools string) *Registry {
	r := &Registry{
		client:   client,
		toolMode: toolMode,
	}
	if toolMode == "custom" && enabledTools != "" {
		r.enabledTools = make(map[string]bool)
		for _, t := range strings.Split(enabledTools, ",") {
			name := strings.TrimSpace(t)
			if name != "" {
				r.enabledTools[name] = true
			}
		}
	}
	return r
}

// shouldRegisterGroup checks if a tool group should be registered.
func (r *Registry) shouldRegisterGroup(groupName string) bool {
	switch r.toolMode {
	case "full":
		return true
	case "default":
		g, ok := toolGroups[groupName]
		return ok && g.IsDefault
	case "custom":
		g, ok := toolGroups[groupName]
		if !ok {
			return false
		}
		for _, t := range g.Tools {
			if r.enabledTools[t] {
				return true
			}
		}
		return false
	default:
		return true
	}
}

// shouldRegisterTool checks if an individual tool should be registered.
func (r *Registry) shouldRegisterTool(toolName string) bool {
	switch r.toolMode {
	case "full":
		return true
	case "default":
		return true // group-level check is sufficient
	case "custom":
		return r.enabledTools[toolName]
	default:
		return true
	}
}

// ListAvailableTools returns all available tool names for the current mode.
func (r *Registry) ListAvailableTools() []string {
	var tools []string
	for name, group := range toolGroups {
		if r.shouldRegisterGroup(name) {
			for _, t := range group.Tools {
				if r.shouldRegisterTool(t) {
					tools = append(tools, t)
				}
			}
		}
	}
	return tools
}

// RegisterAllTools registers all OpenProject tools based on the configured mode.
func (r *Registry) RegisterAllTools(server *mcp.Server) {
	// Connection / utility tools
	if r.shouldRegisterGroup("connection") {
		r.registerConnectionTools(server)
	}
	if r.shouldRegisterGroup("project") {
		r.registerProjectTools(server)
	}
	if r.shouldRegisterGroup("work_package") {
		r.registerWorkPackageTools(server)
	}
	if r.shouldRegisterGroup("user") {
		r.registerUserTools(server)
	}
	if r.shouldRegisterGroup("membership") {
		r.registerMembershipTools(server)
	}
	if r.shouldRegisterGroup("version") {
		r.registerVersionTools(server)
	}
	if r.shouldRegisterGroup("relation") {
		r.registerRelationTools(server)
	}
	if r.shouldRegisterGroup("search") {
		r.registerSearchTools(server)
	}
	if r.shouldRegisterGroup("notification") {
		r.registerNotificationTools(server)
	}
	if r.shouldRegisterGroup("comment") {
		r.registerCommentTools(server)
	}
	if r.shouldRegisterGroup("watcher") {
		r.registerWatcherTools(server)
	}
	if r.shouldRegisterGroup("group") {
		r.registerGroupTools(server)
	}
	if r.shouldRegisterGroup("document") {
		r.registerDocumentTools(server)
	}
	if r.shouldRegisterGroup("query") {
		r.registerQueryTools(server)
	}
	if r.shouldRegisterGroup("placeholder") {
		r.registerPlaceholderTools(server)
	}
	if r.shouldRegisterGroup("configuration") {
		r.registerConfigurationTools(server)
	}
}

// ToolModeHelp returns a help string describing the tool mode configuration.
func ToolModeHelp() string {
	var b strings.Builder
	b.WriteString("Available tool groups:\n\n")
	b.WriteString("Default mode groups (always enabled in 'default' mode):\n")
	for name, g := range toolGroups {
		if g.IsDefault {
			fmt.Fprintf(&b, "  %s: %s\n", name, strings.Join(g.Tools, ", "))
		}
	}
	b.WriteString("\nFull-only mode groups (enabled in 'full' mode):\n")
	for name, g := range toolGroups {
		if !g.IsDefault {
			fmt.Fprintf(&b, "  %s: %s\n", name, strings.Join(g.Tools, ", "))
		}
	}
	return b.String()
}
