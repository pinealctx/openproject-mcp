package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pinealctx/openproject-mcp/internal/openproject"
)

// outputWriter is the output destination.
var outputWriter io.Writer = os.Stdout

// output prints data in the configured format.
func output(data interface{}) error {
	if flagOutput == "json" {
		return outputJSON(data)
	}
	return outputText(data)
}

// outputJSON prints data as formatted JSON.
func outputJSON(data interface{}) error {
	enc := json.NewEncoder(outputWriter)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// outputText prints data in human-readable text format.
func outputText(data interface{}) error {
	switch v := data.(type) {
	case *openproject.ProjectList:
		return outputProjectList(v)
	case *openproject.Project:
		return outputProject(v)
	case *openproject.WorkPackageList:
		return outputWorkPackageList(v)
	case *openproject.WorkPackage:
		return outputWorkPackage(v)
	case *openproject.UserList:
		return outputUserList(v)
	case *openproject.User:
		return outputUser(v)
	case *openproject.MembershipList:
		return outputMembershipList(v)
	case *openproject.Membership:
		return outputMembership(v)
	case *openproject.TimeEntryList:
		return outputTimeEntryList(v)
	case *openproject.TimeEntry:
		return outputTimeEntry(v)
	case *openproject.TimeEntryActivityList:
		return outputTimeEntryActivityList(v)
	case *openproject.VersionList:
		return outputVersionList(v)
	case *openproject.Version:
		return outputVersion(v)
	case *openproject.GridList:
		return outputGridList(v)
	case *openproject.Grid:
		return outputGrid(v)
	case *openproject.NotificationList:
		return outputNotificationList(v)
	case *openproject.Notification:
		return outputNotification(v)
	case *openproject.RelationList:
		return outputRelationList(v)
	case *openproject.Relation:
		return outputRelation(v)
	case *openproject.StatusList:
		return outputStatusList(v)
	case *openproject.Status:
		return outputStatus(v)
	case *openproject.TypeList:
		return outputTypeList(v)
	case *openproject.Type:
		return outputType(v)
	case *openproject.PriorityList:
		return outputPriorityList(v)
	case *openproject.Priority:
		return outputPriority(v)
	case *openproject.RoleList:
		return outputRoleList(v)
	case *openproject.Role:
		return outputRole(v)
	case *openproject.SearchResults:
		return outputSearchResults(v)
	default:
		// Fallback to JSON for unknown types
		return outputJSON(data)
	}
}

// newTabWriter creates a new tabwriter for aligned output.
func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(outputWriter, 0, 0, 2, ' ', 0)
}

// --- Project output ---

func outputProjectList(list *openproject.ProjectList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tIDENTIFIER\tNAME\tACTIVE\tPUBLIC")
	for _, p := range list.Embedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%s\t%t\t%t\n", p.ID, p.Identifier, p.Name, p.Active, p.Public)
	}
	return w.Flush()
}

func outputProject(p *openproject.Project) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", p.ID)
	fmt.Fprintf(outputWriter, "Identifier: %s\n", p.Identifier)
	fmt.Fprintf(outputWriter, "Name: %s\n", p.Name)
	fmt.Fprintf(outputWriter, "Description: %s\n", p.Description.String())
	fmt.Fprintf(outputWriter, "Active: %t\n", p.Active)
	fmt.Fprintf(outputWriter, "Public: %t\n", p.Public)
	if p.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if p.UpdatedAt != nil {
		fmt.Fprintf(outputWriter, "Updated: %s\n", p.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Work Package output ---

func outputWorkPackageList(list *openproject.WorkPackageList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tSUBJECT\tTYPE\tSTATUS\tASSIGNEE\tDUE")
	for _, wp := range list.Embedded.Elements {
		assignee := extractAssignee(wp.Links)
		due := wp.DueDate
		if due == "" {
			due = "-"
		}
		typeName := extractTypeName(wp.Links)
		statusName := extractStatusName(wp.Links)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", wp.ID, truncate(wp.Subject, 40), typeName, statusName, assignee, due)
	}
	return w.Flush()
}

func outputWorkPackage(wp *openproject.WorkPackage) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", wp.ID)
	fmt.Fprintf(outputWriter, "Subject: %s\n", wp.Subject)
	fmt.Fprintf(outputWriter, "Description: %s\n", wp.Description.String())
	fmt.Fprintf(outputWriter, "Type: %s\n", extractTypeName(wp.Links))
	fmt.Fprintf(outputWriter, "Status: %s\n", extractStatusName(wp.Links))
	fmt.Fprintf(outputWriter, "Priority: %s\n", extractPriorityName(wp.Links))
	fmt.Fprintf(outputWriter, "Assignee: %s\n", extractAssignee(wp.Links))
	fmt.Fprintf(outputWriter, "Progress: %d%%\n", wp.PercentageDone)
	if wp.EstimatedTime != "" {
		fmt.Fprintf(outputWriter, "Estimated: %s\n", wp.EstimatedTime)
	}
	if wp.StartDate != "" {
		fmt.Fprintf(outputWriter, "Start: %s\n", wp.StartDate)
	}
	if wp.DueDate != "" {
		fmt.Fprintf(outputWriter, "Due: %s\n", wp.DueDate)
	}
	if wp.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", wp.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if wp.UpdatedAt != nil {
		fmt.Fprintf(outputWriter, "Updated: %s\n", wp.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- User output ---

func outputUserList(list *openproject.UserList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tLOGIN\tNAME\tEMAIL\tSTATUS")
	for _, u := range list.Embedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", u.ID, u.Login, u.Name, u.Email, u.Status)
	}
	return w.Flush()
}

func outputUser(u *openproject.User) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", u.ID)
	fmt.Fprintf(outputWriter, "Login: %s\n", u.Login)
	fmt.Fprintf(outputWriter, "Name: %s\n", u.Name)
	fmt.Fprintf(outputWriter, "Email: %s\n", u.Email)
	fmt.Fprintf(outputWriter, "Admin: %t\n", u.Admin)
	fmt.Fprintf(outputWriter, "Status: %s\n", u.Status)
	fmt.Fprintf(outputWriter, "Language: %s\n", u.Language)
	if u.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", u.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Membership output ---

func outputMembershipList(list *openproject.MembershipList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tPROJECT\tUSER\tROLES")
	for _, m := range list.Embedded.Elements {
		project := extractProjectName(m.Links)
		user := extractPrincipalName(m.Links)
		roles := extractRoleNames(m.Links)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", m.ID, project, user, roles)
	}
	return w.Flush()
}

func outputMembership(m *openproject.Membership) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", m.ID)
	fmt.Fprintf(outputWriter, "Project: %s\n", extractProjectName(m.Links))
	fmt.Fprintf(outputWriter, "User: %s\n", extractPrincipalName(m.Links))
	fmt.Fprintf(outputWriter, "Roles: %s\n", extractRoleNames(m.Links))
	if m.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Time Entry output ---

func outputTimeEntryList(list *openproject.TimeEntryList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tDATE\tHOURS\tUSER\tPROJECT\tCOMMENT")
	for _, t := range list.Embedded.Elements {
		user := extractTimeEntryUser(t.Links)
		project := extractTimeEntryProject(t.Links)
		comment := truncate(t.Comment, 30)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", t.ID, t.SpentOn, t.Hours, user, project, comment)
	}
	return w.Flush()
}

func outputTimeEntry(t *openproject.TimeEntry) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", t.ID)
	fmt.Fprintf(outputWriter, "Date: %s\n", t.SpentOn)
	fmt.Fprintf(outputWriter, "Hours: %s\n", t.Hours)
	fmt.Fprintf(outputWriter, "User: %s\n", extractTimeEntryUser(t.Links))
	fmt.Fprintf(outputWriter, "Project: %s\n", extractTimeEntryProject(t.Links))
	fmt.Fprintf(outputWriter, "Work Package: %s\n", extractTimeEntryWorkPackage(t.Links))
	fmt.Fprintf(outputWriter, "Activity: %s\n", extractTimeEntryActivity(t.Links))
	fmt.Fprintf(outputWriter, "Comment: %s\n", t.Comment)
	if t.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func outputTimeEntryActivityList(list *openproject.TimeEntryActivityList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tDEFAULT\tACTIVE")
	for _, a := range list.Embedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%t\t%t\n", a.ID, a.Name, a.IsDefault, a.Active)
	}
	return w.Flush()
}

// --- Version output ---

func outputVersionList(list *openproject.VersionList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tSTATUS\tSTART\tEND")
	for _, v := range list.Embedded.Elements {
		start := v.StartDate
		if start == "" {
			start = "-"
		}
		end := v.EndDate
		if end == "" {
			end = "-"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Status, start, end)
	}
	return w.Flush()
}

func outputVersion(v *openproject.Version) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", v.ID)
	fmt.Fprintf(outputWriter, "Name: %s\n", v.Name)
	fmt.Fprintf(outputWriter, "Description: %s\n", v.Description)
	fmt.Fprintf(outputWriter, "Status: %s\n", v.Status)
	if v.StartDate != "" {
		fmt.Fprintf(outputWriter, "Start: %s\n", v.StartDate)
	}
	if v.EndDate != "" {
		fmt.Fprintf(outputWriter, "End: %s\n", v.EndDate)
	}
	if v.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", v.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Grid/Board output ---

func outputGridList(list *openproject.GridList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tROWS\tCOLS\tWIDGETS")
	for _, g := range list.Embedded.Elements {
		widgets := 0
		if g.Embedded != nil {
			widgets = len(g.Embedded.Widgets)
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\n", g.ID, g.RowCount, g.ColumnCount, widgets)
	}
	return w.Flush()
}

func outputGrid(g *openproject.Grid) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", g.ID)
	fmt.Fprintf(outputWriter, "Rows: %d\n", g.RowCount)
	fmt.Fprintf(outputWriter, "Columns: %d\n", g.ColumnCount)
	if g.Embedded != nil {
		fmt.Fprintf(outputWriter, "Widgets:\n")
		for _, w := range g.Embedded.Widgets {
			fmt.Fprintf(outputWriter, "  - ID: %d, Type: %s, Position: (%d,%d) to (%d,%d)\n",
				w.ID, w.Identifier, w.StartRow, w.StartColumn, w.EndRow, w.EndColumn)
		}
	}
	return nil
}

// --- Notification output ---

func outputNotificationList(list *openproject.NotificationList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tREASON\tREAD\tCREATED")
	for _, n := range list.Embedded.Elements {
		read := "No"
		if n.ReadIAN != nil && *n.ReadIAN {
			read = "Yes"
		}
		created := "-"
		if n.CreatedAt != nil {
			created = n.CreatedAt.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", n.ID, n.Reason, read, created)
	}
	return w.Flush()
}

func outputNotification(n *openproject.Notification) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", n.ID)
	fmt.Fprintf(outputWriter, "Reason: %s\n", n.Reason)
	read := "No"
	if n.ReadIAN != nil && *n.ReadIAN {
		read = "Yes"
	}
	fmt.Fprintf(outputWriter, "Read: %s\n", read)
	if n.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", n.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if n.Links != nil {
		if n.Links.Actor != nil {
			fmt.Fprintf(outputWriter, "Actor: %s\n", n.Links.Actor.Title)
		}
		if n.Links.Project != nil {
			fmt.Fprintf(outputWriter, "Project: %s\n", n.Links.Project.Title)
		}
	}
	return nil
}

// --- Relation output ---

func outputRelationList(list *openproject.RelationList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tTYPE\tFROM\tTO\tDELAY")
	for _, r := range list.Embedded.Elements {
		from := extractRelationFrom(r.Links)
		to := extractRelationTo(r.Links)
		delay := "-"
		if r.Delay > 0 {
			delay = fmt.Sprintf("%d days", r.Delay)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", r.ID, r.Type, from, to, delay)
	}
	return w.Flush()
}

func outputRelation(r *openproject.Relation) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", r.ID)
	fmt.Fprintf(outputWriter, "Type: %s\n", r.Type)
	fmt.Fprintf(outputWriter, "From: %s\n", extractRelationFrom(r.Links))
	fmt.Fprintf(outputWriter, "To: %s\n", extractRelationTo(r.Links))
	if r.Description != "" {
		fmt.Fprintf(outputWriter, "Description: %s\n", r.Description)
	}
	if r.Delay > 0 {
		fmt.Fprintf(outputWriter, "Delay: %d days\n", r.Delay)
	}
	return nil
}

// --- Status output ---

func outputStatusList(list *openproject.StatusList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tDEFAULT\tCLOSED\tREADONLY")
	for _, s := range list.Embedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%t\t%t\t%t\n", s.ID, s.Name, s.IsDefault, s.IsClosed, s.IsReadonly)
	}
	return w.Flush()
}

func outputStatus(s *openproject.Status) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", s.ID)
	fmt.Fprintf(outputWriter, "Name: %s\n", s.Name)
	fmt.Fprintf(outputWriter, "Position: %d\n", s.Position)
	fmt.Fprintf(outputWriter, "Default: %t\n", s.IsDefault)
	fmt.Fprintf(outputWriter, "Closed: %t\n", s.IsClosed)
	fmt.Fprintf(outputWriter, "Readonly: %t\n", s.IsReadonly)
	if s.Color != "" {
		fmt.Fprintf(outputWriter, "Color: %s\n", s.Color)
	}
	return nil
}

// --- Type output ---

func outputTypeList(list *openproject.TypeList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tDEFAULT\tMILESTONE")
	for _, t := range list.Embedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%t\t%t\n", t.ID, t.Name, t.IsDefault, t.IsMilestone)
	}
	return w.Flush()
}

func outputType(t *openproject.Type) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", t.ID)
	fmt.Fprintf(outputWriter, "Name: %s\n", t.Name)
	fmt.Fprintf(outputWriter, "Position: %d\n", t.Position)
	fmt.Fprintf(outputWriter, "Default: %t\n", t.IsDefault)
	fmt.Fprintf(outputWriter, "Milestone: %t\n", t.IsMilestone)
	if t.Color != "" {
		fmt.Fprintf(outputWriter, "Color: %s\n", t.Color)
	}
	return nil
}

// --- Priority output ---

func outputPriorityList(list *openproject.PriorityList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tDEFAULT")
	for _, p := range list.Embedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%t\n", p.ID, p.Name, p.IsDefault)
	}
	return w.Flush()
}

func outputPriority(p *openproject.Priority) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", p.ID)
	fmt.Fprintf(outputWriter, "Name: %s\n", p.Name)
	fmt.Fprintf(outputWriter, "Position: %d\n", p.Position)
	fmt.Fprintf(outputWriter, "Default: %t\n", p.IsDefault)
	if p.Color != "" {
		fmt.Fprintf(outputWriter, "Color: %s\n", p.Color)
	}
	return nil
}

// --- Role output ---

func outputRoleList(list *openproject.RoleList) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME")
	for _, r := range list.Embedded.Elements {
		fmt.Fprintf(w, "%d\t%s\n", r.ID, r.Name)
	}
	return w.Flush()
}

func outputRole(r *openproject.Role) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", r.ID)
	fmt.Fprintf(outputWriter, "Name: %s\n", r.Name)
	if len(r.Permissions) > 0 {
		fmt.Fprintf(outputWriter, "Permissions: %s\n", strings.Join(r.Permissions, ", "))
	}
	return nil
}

// --- Search output ---

func outputSearchResults(results *openproject.SearchResults) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tTYPE\tTITLE")
	for _, r := range results.Embedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%s\n", r.ID, r.Type, r.Title)
	}
	return w.Flush()
}

// --- Helper functions ---

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func extractAssignee(links *openproject.WorkPackageLinks) string {
	if links == nil || links.Assignee == nil {
		return "-"
	}
	return links.Assignee.Title
}

func extractTypeName(links *openproject.WorkPackageLinks) string {
	if links == nil || links.Type == nil {
		return "-"
	}
	return links.Type.Title
}

func extractStatusName(links *openproject.WorkPackageLinks) string {
	if links == nil || links.Status == nil {
		return "-"
	}
	return links.Status.Title
}

func extractPriorityName(links *openproject.WorkPackageLinks) string {
	if links == nil || links.Priority == nil {
		return "-"
	}
	return links.Priority.Title
}

func extractProjectName(links *openproject.MembershipLinks) string {
	if links == nil || links.Project == nil {
		return "-"
	}
	return links.Project.Title
}

func extractPrincipalName(links *openproject.MembershipLinks) string {
	if links == nil || links.Principal == nil {
		return "-"
	}
	return links.Principal.Title
}

func extractRoleNames(links *openproject.MembershipLinks) string {
	if links == nil || len(links.Roles) == 0 {
		return "-"
	}
	names := make([]string, len(links.Roles))
	for i, r := range links.Roles {
		names[i] = r.Title
	}
	return strings.Join(names, ", ")
}

func extractTimeEntryUser(links *openproject.TimeEntryLinks) string {
	if links == nil || links.User == nil {
		return "-"
	}
	return links.User.Title
}

func extractTimeEntryProject(links *openproject.TimeEntryLinks) string {
	if links == nil || links.Project == nil {
		return "-"
	}
	return links.Project.Title
}

func extractTimeEntryWorkPackage(links *openproject.TimeEntryLinks) string {
	if links == nil || links.WorkPackage == nil {
		return "-"
	}
	return links.WorkPackage.Title
}

func extractTimeEntryActivity(links *openproject.TimeEntryLinks) string {
	if links == nil || links.Activity == nil {
		return "-"
	}
	return links.Activity.Title
}

func extractRelationFrom(links *openproject.RelationLinks) string {
	if links == nil || links.From == nil {
		return "-"
	}
	return links.From.Title
}

func extractRelationTo(links *openproject.RelationLinks) string {
	if links == nil || links.To == nil {
		return "-"
	}
	return links.To.Title
}
