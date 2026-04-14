package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
	external "github.com/pinealctx/openproject"
	oapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/spf13/cobra"
)

var (
	wpListProjectID int
	wpListPageSize  int
	wpListSortBy    string
	wpListFilters   string

	wpCreateProjectID     int
	wpCreateSubject       string
	wpCreateDescription   string
	wpCreateType          string
	wpCreateStatus        string
	wpCreatePriority      string
	wpCreateAssignee      int
	wpCreateStartDate     string
	wpCreateDueDate       string
	wpCreateEstimatedTime string

	wpUpdateSubject       string
	wpUpdateDescription   string
	wpUpdateStatus        string
	wpUpdatePriority      string
	wpUpdateAssignee      int
	wpUpdateStartDate     string
	wpUpdateDueDate       string
	wpUpdateEstimatedTime string
	wpUpdateProgress      int

	wpRelationFromID int
	wpRelationToID   int
	wpRelationType   string
	wpRelationDesc   string
	wpRelationDelay  int
)

// workPackageCmd represents the work-package command.
var workPackageCmd = &cobra.Command{
	Use:   "work-package",
	Short: "Manage work packages (tasks, bugs, features, etc.)",
	Long: `Manage work packages in OpenProject.

Work packages are the core work items in OpenProject. They can represent:
  - Tasks: Regular work items
  - Bugs: Defects or issues
  - Features: New functionality
  - Milestones: Project milestones
  - And custom types defined in your OpenProject instance

Work packages can be organized hierarchically (parent/child relationships)
and linked together with various relation types.

Available subcommands:
  list          List work packages (optionally filtered by project)
  get           Get details of a specific work package
  create        Create a new work package
  update        Update work package properties
  delete        Delete a work package
  children      List child work packages
  set-parent    Set parent work package (for hierarchy)
  remove-parent Remove parent relationship
  relation      Manage relations between work packages

Examples:
  # List all work packages in a project
  openproject-mcp wp list -p 42

  # List work packages assigned to me
  openproject-mcp wp list --filters '[{"assignee":{"operator":"=","values":["me"]}}]'

  # Get work package details
  openproject-mcp wp get 123

  # Create a new task
  openproject-mcp wp create -p 42 -s "Implement user authentication" -d "Add login/logout functionality"

  # Create a bug report
  openproject-mcp wp create -p 42 -s "Fix login crash" --type "Bug" --priority "High"

  # Update work package status
  openproject-mcp wp update 123 --status "In Progress"

  # Update progress
  openproject-mcp wp update 123 --progress 75

  # Assign work package
  openproject-mcp wp update 123 --assignee 5

  # Set parent for hierarchy
  openproject-mcp wp set-parent 123 -p 100

  # Create a "blocks" relation
  openproject-mcp wp relation create --from 123 --to 456 --type blocks

  # Delete a work package
  openproject-mcp wp delete 123`,
	Aliases: []string{"wp"},
}

var wpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List work packages",
	Long: `List work packages visible to the authenticated user.

Can be filtered by project and other criteria using filters.
Use -o json for machine-readable output.

Examples:
  # List all work packages
  openproject-mcp wp list

  # List work packages in a specific project
  openproject-mcp wp list -p 42

  # List work packages sorted by due date
  openproject-mcp wp list --sort-by "dueDate:asc"

  # Filter by status using JSON filters
  openproject-mcp wp list --filters '[{"status":{"operator":"=","values":["In Progress"]}}]'

  # Output as JSON
  openproject-mcp wp list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		api := getClient().APIClient()
		var result openproject.WorkPackagesModel

		if wpListProjectID > 0 {
			params := &openproject.GetProjectWorkPackageCollectionParams{}
			if wpListPageSize > 0 {
				params.PageSize = ptr(wpListPageSize)
			}
			if wpListSortBy != "" {
				params.SortBy = ptr(normalizeSortBy(wpListSortBy))
			}
			if wpListFilters != "" {
				params.Filters = ptr(wpListFilters)
			}
			resp, err := api.GetProjectWorkPackageCollection(getContext(), wpListProjectID, params)
			if err != nil {
				return err
			}
			if err := openproject.ReadResponse(resp, &result); err != nil {
				return err
			}
		} else {
			params := &openproject.ListWorkPackagesParams{}
			if wpListPageSize > 0 {
				params.PageSize = ptr(wpListPageSize)
			}
			if wpListSortBy != "" {
				params.SortBy = ptr(normalizeSortBy(wpListSortBy))
			}
			if wpListFilters != "" {
				params.Filters = ptr(wpListFilters)
			}
			resp, err := api.ListWorkPackages(getContext(), params)
			if err != nil {
				return err
			}
			if err := openproject.ReadResponse(resp, &result); err != nil {
				return err
			}
		}
		return output(&result)
	},
}

var wpGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get work package details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid work package ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.ViewWorkPackage(getContext(), id, nil)
		if err != nil {
			return err
		}
		var result openproject.WorkPackageModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var wpCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new work package",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := external.WorkPackageModel{
			Subject: wpCreateSubject,
		}
		if wpCreateDescription != "" {
			f := external.FormattableFormat("markdown")
			body.Description = &external.Formattable{Format: &f, Raw: ptr(wpCreateDescription)}
		}
		if wpCreateStartDate != "" {
			body.StartDate = parseDate(wpCreateStartDate)
		}
		if wpCreateDueDate != "" {
			body.DueDate = parseDate(wpCreateDueDate)
		}
		if wpCreateEstimatedTime != "" {
			body.EstimatedTime = ptr(wpCreateEstimatedTime)
		}

		// Set links
		if wpCreateType != "" {
			body.UnderscoreLinks.Type = external.Link{Href: ptr("/api/v3/types/" + wpCreateType)}
		}
		if wpCreateStatus != "" {
			body.UnderscoreLinks.Status = external.Link{Href: ptr("/api/v3/statuses/" + wpCreateStatus)}
		}
		if wpCreatePriority != "" {
			body.UnderscoreLinks.Priority = external.Link{Href: ptr("/api/v3/priorities/" + wpCreatePriority)}
		}
		if wpCreateAssignee > 0 {
			body.UnderscoreLinks.Assignee = &external.Link{Href: ptr(fmt.Sprintf("/api/v3/users/%d", wpCreateAssignee))}
		}

		api := getClient().APIClient()
		resp, err := api.CreateProjectWorkPackage(getContext(), wpCreateProjectID, nil, body)
		if err != nil {
			return err
		}
		var result openproject.WorkPackageModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var wpUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a work package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid work package ID: %s", args[0])
		}

		// Use raw PATCH since WorkPackagePatchModel doesn't include all fields
		// (e.g., percentageDone is missing from the patch model)
		body := map[string]interface{}{}
		if wpUpdateSubject != "" {
			body["subject"] = wpUpdateSubject
		}
		if wpUpdateDescription != "" {
			body["description"] = map[string]interface{}{
				"format": "markdown",
				"raw":    wpUpdateDescription,
			}
		}
		if wpUpdateStartDate != "" {
			body["startDate"] = wpUpdateStartDate
		}
		if wpUpdateDueDate != "" {
			body["dueDate"] = wpUpdateDueDate
		}
		if wpUpdateEstimatedTime != "" {
			body["estimatedTime"] = wpUpdateEstimatedTime
		}
		if cmd.Flags().Changed("progress") {
			body["percentageDone"] = wpUpdateProgress
		}

		// Set links
		links := map[string]interface{}{}
		if wpUpdateStatus != "" {
			links["status"] = map[string]interface{}{
				"href": "/api/v3/statuses/" + wpUpdateStatus,
			}
		}
		if wpUpdatePriority != "" {
			links["priority"] = map[string]interface{}{
				"href": "/api/v3/priorities/" + wpUpdatePriority,
			}
		}
		if wpUpdateAssignee > 0 {
			links["assignee"] = map[string]interface{}{
				"href": fmt.Sprintf("/api/v3/users/%d", wpUpdateAssignee),
			}
		}
		if len(links) > 0 {
			body["_links"] = links
		}

		var result openproject.WorkPackageModel
		if err := getClient().Patch(getContext(), fmt.Sprintf("/work_packages/%d", id), body, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var wpDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a work package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid work package ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.DeleteWorkPackage(getContext(), id)
		if err != nil {
			return err
		}
		if err := openproject.ReadResponse(resp, nil); err != nil {
			return err
		}
		fmt.Println("Work package deleted successfully")
		return nil
	},
}

var wpChildrenCmd = &cobra.Command{
	Use:   "children <id>",
	Short: "List children of a work package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid work package ID: %s", args[0])
		}
		var result openproject.WorkPackagesModel
		if err := getClient().Get(getContext(), fmt.Sprintf("/work_packages/%d/children", id), &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var wpParentID int

var wpSetParentCmd = &cobra.Command{
	Use:   "set-parent <id>",
	Short: "Set parent of a work package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid work package ID: %s", args[0])
		}
		body := map[string]interface{}{
			"_links": map[string]interface{}{
				"parent": map[string]interface{}{
					"href": fmt.Sprintf("/api/v3/work_packages/%d", wpParentID),
				},
			},
		}
		var result openproject.WorkPackageModel
		if err := getClient().Patch(getContext(), fmt.Sprintf("/work_packages/%d", id), body, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var wpRemoveParentCmd = &cobra.Command{
	Use:   "remove-parent <id>",
	Short: "Remove parent from a work package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid work package ID: %s", args[0])
		}
		body := map[string]interface{}{
			"lockVersion": 0,
			"_links": map[string]interface{}{
				"parent": nil,
			},
		}
		var result openproject.WorkPackageModel
		if err := getClient().Patch(getContext(), fmt.Sprintf("/work_packages/%d", id), body, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

// Relation subcommands
var wpRelationCmd = &cobra.Command{
	Use:   "relation",
	Short: "Manage work package relations",
}

var wpRelationListCmd = &cobra.Command{
	Use:   "list <id>",
	Short: "List relations for a work package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid work package ID: %s", args[0])
		}
		api := getClient().APIClient()
		params := &openproject.ListRelationsParams{
			Filters: ptr(fmt.Sprintf(`[{"from":{"operator":"=","values":["%d"]}}]`, id)),
		}
		resp, err := api.ListRelations(getContext(), params)
		if err != nil {
			return err
		}
		var result openproject.RelationCollectionModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var wpRelationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a relation between work packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := external.RelationWriteModel{
			Type: external.RelationWriteModelType(wpRelationType),
		}
		body.UnderscoreLinks.To = &external.Link{Href: ptr(fmt.Sprintf("/api/v3/work_packages/%d", wpRelationToID))}
		if wpRelationDesc != "" {
			body.Description = ptr(wpRelationDesc)
		}
		if wpRelationDelay > 0 {
			body.Lag = ptr(wpRelationDelay)
		}

		// CreateRelation takes the FROM work package ID
		api := getClient().APIClient()
		resp, err := api.CreateRelation(getContext(), wpRelationFromID, body)
		if err != nil {
			return err
		}
		var result openproject.RelationReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var wpRelationGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get relation details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid relation ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.GetRelation(getContext(), id)
		if err != nil {
			return err
		}
		var result openproject.RelationReadModel
		if err := openproject.ReadResponse(resp, &result); err != nil {
			return err
		}
		return output(&result)
	},
}

var wpRelationDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a relation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid relation ID: %s", args[0])
		}
		api := getClient().APIClient()
		resp, err := api.DeleteRelation(getContext(), id)
		if err != nil {
			return err
		}
		if err := openproject.ReadResponse(resp, nil); err != nil {
			return err
		}
		fmt.Println("Relation deleted successfully")
		return nil
	},
}

// parseDate parses a "YYYY-MM-DD" string into an openapi_types.Date pointer.
// Returns nil if the string is empty.
func parseDate(s string) *oapi_types.Date {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	d := oapi_types.Date{Time: t}
	return &d
}

func init() {
	rootCmd.AddCommand(workPackageCmd)
	workPackageCmd.AddCommand(wpListCmd)
	workPackageCmd.AddCommand(wpGetCmd)
	workPackageCmd.AddCommand(wpCreateCmd)
	workPackageCmd.AddCommand(wpUpdateCmd)
	workPackageCmd.AddCommand(wpDeleteCmd)
	workPackageCmd.AddCommand(wpChildrenCmd)
	workPackageCmd.AddCommand(wpSetParentCmd)
	workPackageCmd.AddCommand(wpRemoveParentCmd)
	workPackageCmd.AddCommand(wpRelationCmd)

	wpRelationCmd.AddCommand(wpRelationListCmd)
	wpRelationCmd.AddCommand(wpRelationCreateCmd)
	wpRelationCmd.AddCommand(wpRelationGetCmd)
	wpRelationCmd.AddCommand(wpRelationDeleteCmd)

	// List flags
	wpListCmd.Flags().IntVarP(&wpListProjectID, "project", "p", 0, "Filter by project ID")
	wpListCmd.Flags().IntVarP(&wpListPageSize, "page-size", "s", 20, "Number of results per page")
	wpListCmd.Flags().StringVar(&wpListSortBy, "sort-by", "updatedAt:desc", "Sort criteria")
	wpListCmd.Flags().StringVar(&wpListFilters, "filters", "", "Raw JSON filter string")

	// Create flags
	wpCreateCmd.Flags().IntVarP(&wpCreateProjectID, "project", "p", 0, "Project ID (required)")
	wpCreateCmd.Flags().StringVarP(&wpCreateSubject, "subject", "s", "", "Work package subject (required)")
	wpCreateCmd.Flags().StringVarP(&wpCreateDescription, "description", "d", "", "Description")
	wpCreateCmd.Flags().StringVar(&wpCreateType, "type", "", "Type ID or name")
	wpCreateCmd.Flags().StringVar(&wpCreateStatus, "status", "", "Status ID")
	wpCreateCmd.Flags().StringVar(&wpCreatePriority, "priority", "", "Priority ID")
	wpCreateCmd.Flags().IntVar(&wpCreateAssignee, "assignee", 0, "Assignee user ID")
	wpCreateCmd.Flags().StringVar(&wpCreateStartDate, "start", "", "Start date (YYYY-MM-DD)")
	wpCreateCmd.Flags().StringVar(&wpCreateDueDate, "due", "", "Due date (YYYY-MM-DD)")
	wpCreateCmd.Flags().StringVar(&wpCreateEstimatedTime, "estimate", "", "Estimated time (e.g., PT4H)")
	_ = wpCreateCmd.MarkFlagRequired("project")
	_ = wpCreateCmd.MarkFlagRequired("subject")

	// Update flags
	wpUpdateCmd.Flags().StringVarP(&wpUpdateSubject, "subject", "s", "", "Work package subject")
	wpUpdateCmd.Flags().StringVarP(&wpUpdateDescription, "description", "d", "", "Description")
	wpUpdateCmd.Flags().StringVar(&wpUpdateStatus, "status", "", "Status ID")
	wpUpdateCmd.Flags().StringVar(&wpUpdatePriority, "priority", "", "Priority ID")
	wpUpdateCmd.Flags().IntVar(&wpUpdateAssignee, "assignee", 0, "Assignee user ID")
	wpUpdateCmd.Flags().StringVar(&wpUpdateStartDate, "start", "", "Start date (YYYY-MM-DD)")
	wpUpdateCmd.Flags().StringVar(&wpUpdateDueDate, "due", "", "Due date (YYYY-MM-DD)")
	wpUpdateCmd.Flags().StringVar(&wpUpdateEstimatedTime, "estimate", "", "Estimated time (e.g., PT4H)")
	wpUpdateCmd.Flags().IntVarP(&wpUpdateProgress, "progress", "r", 0, "Percentage done (0-100)")

	// Set parent flags
	wpSetParentCmd.Flags().IntVarP(&wpParentID, "parent", "p", 0, "Parent work package ID (required)")
	_ = wpSetParentCmd.MarkFlagRequired("parent")

	// Relation create flags
	wpRelationCreateCmd.Flags().IntVarP(&wpRelationFromID, "from", "f", 0, "From work package ID (required)")
	wpRelationCreateCmd.Flags().IntVarP(&wpRelationToID, "to", "t", 0, "To work package ID (required)")
	wpRelationCreateCmd.Flags().StringVarP(&wpRelationType, "type", "y", "relates", "Relation type (relates, follows, precedes, blocks, blocked_by, duplicates, duplicated)")
	wpRelationCreateCmd.Flags().StringVarP(&wpRelationDesc, "description", "d", "", "Description")
	wpRelationCreateCmd.Flags().IntVar(&wpRelationDelay, "delay", 0, "Delay in days")
	_ = wpRelationCreateCmd.MarkFlagRequired("from")
	_ = wpRelationCreateCmd.MarkFlagRequired("to")
}
