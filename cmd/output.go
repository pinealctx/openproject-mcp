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

var outputWriter io.Writer = os.Stdout

func output(data interface{}) error {
	if flagOutput == "json" {
		return outputJSON(data)
	}
	return outputText(data)
}

func outputJSON(data interface{}) error {
	enc := json.NewEncoder(outputWriter)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func outputText(data interface{}) error {
	switch v := data.(type) {
	case *openproject.ProjectCollectionModel:
		return outputProjectList(v)
	case *openproject.ProjectModel:
		return outputProject(v)
	case *openproject.WorkPackagesModel:
		return outputWorkPackageList(v)
	case *openproject.WorkPackageModel:
		return outputWorkPackage(v)
	case *openproject.UserCollectionModel:
		return outputUserList(v)
	case *openproject.UserModel:
		return outputUser(v)
	case *openproject.MembershipCollectionModel:
		return outputMembershipList(v)
	case *openproject.MembershipReadModel:
		return outputMembership(v)
	case *openproject.TimeEntryCollectionModel:
		return outputTimeEntryList(v)
	case *openproject.TimeEntryModel:
		return outputTimeEntry(v)
	case *openproject.VersionCollectionModel:
		return outputVersionList(v)
	case *openproject.VersionReadModel:
		return outputVersion(v)
	case *openproject.GridCollectionModel:
		return outputGridList(v)
	case *openproject.GridReadModel:
		return outputGrid(v)
	case *openproject.NotificationCollectionModel:
		return outputNotificationList(v)
	case *openproject.NotificationModel:
		return outputNotification(v)
	case *openproject.RelationCollectionModel:
		return outputRelationList(v)
	case *openproject.RelationReadModel:
		return outputRelation(v)
	case *openproject.StatusCollectionModel:
		return outputStatusList(v)
	case *openproject.StatusModel:
		return outputStatus(v)
	case *openproject.TypesByWorkspaceModel:
		return outputTypeList(v)
	case *openproject.TypeModel:
		return outputType(v)
	case *openproject.PriorityCollectionModel:
		return outputPriorityList(v)
	case *openproject.PriorityModel:
		return outputPriority(v)
	default:
		return outputJSON(data)
	}
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(outputWriter, 0, 0, 2, ' ', 0)
}

func dStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func dInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func dBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// --- Project output ---

func outputProjectList(list *openproject.ProjectCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tIDENTIFIER\tNAME\tACTIVE\tPUBLIC")
	for _, p := range list.UnderscoreEmbedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%s\t%t\t%t\n", dInt(p.Id), dStr(p.Identifier), dStr(p.Name), dBool(p.Active), dBool(p.Public))
	}
	return w.Flush()
}

func outputProject(p *openproject.ProjectModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", dInt(p.Id))
	fmt.Fprintf(outputWriter, "Identifier: %s\n", dStr(p.Identifier))
	fmt.Fprintf(outputWriter, "Name: %s\n", dStr(p.Name))
	if p.Description != nil && p.Description.Raw != nil {
		fmt.Fprintf(outputWriter, "Description: %s\n", *p.Description.Raw)
	}
	fmt.Fprintf(outputWriter, "Active: %t\n", dBool(p.Active))
	fmt.Fprintf(outputWriter, "Public: %t\n", dBool(p.Public))
	if p.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if p.UpdatedAt != nil {
		fmt.Fprintf(outputWriter, "Updated: %s\n", p.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Work Package output ---

func outputWorkPackageList(list *openproject.WorkPackagesModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tSUBJECT\tTYPE\tSTATUS\tASSIGNEE")
	for _, wp := range list.UnderscoreEmbedded.Elements {
		assignee := "-"
		if wp.UnderscoreLinks.Assignee != nil {
			assignee = dStr(wp.UnderscoreLinks.Assignee.Title)
		}
		typeName := dStr(wp.UnderscoreLinks.Type.Title)
		statusName := dStr(wp.UnderscoreLinks.Status.Title)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", dInt(wp.Id), truncate(wp.Subject, 40), typeName, statusName, assignee)
	}
	return w.Flush()
}

func outputWorkPackage(wp *openproject.WorkPackageModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", dInt(wp.Id))
	fmt.Fprintf(outputWriter, "Subject: %s\n", wp.Subject)
	if wp.Description != nil && wp.Description.Raw != nil {
		fmt.Fprintf(outputWriter, "Description: %s\n", *wp.Description.Raw)
	}
	fmt.Fprintf(outputWriter, "Type: %s\n", dStr(wp.UnderscoreLinks.Type.Title))
	fmt.Fprintf(outputWriter, "Status: %s\n", dStr(wp.UnderscoreLinks.Status.Title))
	fmt.Fprintf(outputWriter, "Priority: %s\n", dStr(wp.UnderscoreLinks.Priority.Title))
	if wp.UnderscoreLinks.Assignee != nil {
		fmt.Fprintf(outputWriter, "Assignee: %s\n", dStr(wp.UnderscoreLinks.Assignee.Title))
	}
	if wp.PercentageDone != nil {
		fmt.Fprintf(outputWriter, "Progress: %d%%\n", *wp.PercentageDone)
	}
	if wp.EstimatedTime != nil {
		fmt.Fprintf(outputWriter, "Estimated: %s\n", *wp.EstimatedTime)
	}
	if wp.StartDate != nil {
		fmt.Fprintf(outputWriter, "Start: %s\n", wp.StartDate.String())
	}
	if wp.DueDate != nil {
		fmt.Fprintf(outputWriter, "Due: %s\n", wp.DueDate.String())
	}
	if wp.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", wp.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- User output ---

func outputUserList(list *openproject.UserCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tSTATUS")
	for _, u := range list.UnderscoreEmbedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", u.Id, u.Name, dStr(u.Email), dStr(u.Status))
	}
	return w.Flush()
}

func outputUser(u *openproject.UserModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", u.Id)
	fmt.Fprintf(outputWriter, "Name: %s\n", u.Name)
	if u.Login != nil {
		fmt.Fprintf(outputWriter, "Login: %s\n", *u.Login)
	}
	if u.Email != nil {
		fmt.Fprintf(outputWriter, "Email: %s\n", *u.Email)
	}
	if u.Admin != nil {
		fmt.Fprintf(outputWriter, "Admin: %t\n", *u.Admin)
	}
	if u.Status != nil {
		fmt.Fprintf(outputWriter, "Status: %s\n", *u.Status)
	}
	if u.Language != nil {
		fmt.Fprintf(outputWriter, "Language: %s\n", *u.Language)
	}
	if u.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", u.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Membership output ---

func outputMembershipList(list *openproject.MembershipCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tROLES")
	for _, m := range list.UnderscoreEmbedded.Elements {
		roles := "-"
		if len(m.UnderscoreLinks.Roles) > 0 {
			names := make([]string, len(m.UnderscoreLinks.Roles))
			for i, r := range m.UnderscoreLinks.Roles {
				names[i] = dStr(r.Title)
			}
			roles = strings.Join(names, ", ")
		}
		fmt.Fprintf(w, "%d\t%s\n", m.Id, roles)
	}
	return w.Flush()
}

func outputMembership(m *openproject.MembershipReadModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", m.Id)
	roles := "-"
	if len(m.UnderscoreLinks.Roles) > 0 {
		names := make([]string, len(m.UnderscoreLinks.Roles))
		for i, r := range m.UnderscoreLinks.Roles {
			names[i] = dStr(r.Title)
		}
		roles = strings.Join(names, ", ")
	}
	fmt.Fprintf(outputWriter, "Roles: %s\n", roles)
	fmt.Fprintf(outputWriter, "Created: %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))
	return nil
}

// --- Time Entry output ---

func outputTimeEntryList(list *openproject.TimeEntryCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tDATE\tHOURS\tCOMMENT")
	for _, t := range list.UnderscoreEmbedded.Elements {
		comment := ""
		if t.Comment != nil && t.Comment.Raw != nil {
			comment = truncate(*t.Comment.Raw, 30)
		}
		spentOn := "-"
		if t.SpentOn != nil {
			spentOn = t.SpentOn.String()
		}
		hours := "-"
		if t.Hours != nil {
			hours = *t.Hours
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", dInt(t.Id), spentOn, hours, comment)
	}
	return w.Flush()
}

func outputTimeEntry(t *openproject.TimeEntryModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", dInt(t.Id))
	if t.SpentOn != nil {
		fmt.Fprintf(outputWriter, "Date: %s\n", t.SpentOn.String())
	}
	if t.Hours != nil {
		fmt.Fprintf(outputWriter, "Hours: %s\n", *t.Hours)
	}
	if t.Comment != nil && t.Comment.Raw != nil {
		fmt.Fprintf(outputWriter, "Comment: %s\n", *t.Comment.Raw)
	}
	if t.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Version output ---

func outputVersionList(list *openproject.VersionCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tSTATUS\tSTART\tEND")
	for _, v := range list.UnderscoreEmbedded.Elements {
		start := "-"
		if v.StartDate != nil {
			start = v.StartDate.String()
		}
		end := "-"
		if v.EndDate != nil {
			end = v.EndDate.String()
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", v.Id, v.Name, v.Status, start, end)
	}
	return w.Flush()
}

func outputVersion(v *openproject.VersionReadModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", v.Id)
	fmt.Fprintf(outputWriter, "Name: %s\n", v.Name)
	fmt.Fprintf(outputWriter, "Status: %s\n", v.Status)
	if v.StartDate != nil {
		fmt.Fprintf(outputWriter, "Start: %s\n", v.StartDate.String())
	}
	if v.EndDate != nil {
		fmt.Fprintf(outputWriter, "End: %s\n", v.EndDate.String())
	}
	fmt.Fprintf(outputWriter, "Created: %s\n", v.CreatedAt.Format("2006-01-02 15:04:05"))
	return nil
}

// --- Grid/Board output ---

func outputGridList(list *openproject.GridCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tROWS\tCOLS\tWIDGETS")
	for _, g := range list.UnderscoreEmbedded.Elements {
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\n", g.Id, g.RowCount, g.ColumnCount, len(g.Widgets))
	}
	return w.Flush()
}

func outputGrid(g *openproject.GridReadModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", g.Id)
	fmt.Fprintf(outputWriter, "Rows: %d\n", g.RowCount)
	fmt.Fprintf(outputWriter, "Columns: %d\n", g.ColumnCount)
	if len(g.Widgets) > 0 {
		fmt.Fprintf(outputWriter, "Widgets:\n")
		for _, w := range g.Widgets {
			fmt.Fprintf(outputWriter, "  - ID: %d, Type: %s, Position: (%d,%d) to (%d,%d)\n",
				dInt(w.Id), w.Identifier, w.StartRow, w.StartColumn, w.EndRow, w.EndColumn)
		}
	}
	return nil
}

// --- Notification output ---

func outputNotificationList(list *openproject.NotificationCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tREASON\tREAD\tCREATED")
	for _, n := range list.UnderscoreEmbedded.Elements {
		read := "No"
		if n.ReadIAN != nil && *n.ReadIAN {
			read = "Yes"
		}
		created := "-"
		if n.CreatedAt != nil {
			created = n.CreatedAt.Format("2006-01-02 15:04")
		}
		reason := "-"
		if n.Reason != nil {
			reason = string(*n.Reason)
		}
		id := 0
		if n.Id != nil {
			id = *n.Id
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", id, reason, read, created)
	}
	return w.Flush()
}

func outputNotification(n *openproject.NotificationModel) error {
	if n.Id != nil {
		fmt.Fprintf(outputWriter, "ID: %d\n", *n.Id)
	}
	if n.Reason != nil {
		fmt.Fprintf(outputWriter, "Reason: %s\n", *n.Reason)
	}
	read := "No"
	if n.ReadIAN != nil && *n.ReadIAN {
		read = "Yes"
	}
	fmt.Fprintf(outputWriter, "Read: %s\n", read)
	if n.CreatedAt != nil {
		fmt.Fprintf(outputWriter, "Created: %s\n", n.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Relation output ---

func outputRelationList(list *openproject.RelationCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tTYPE\tDELAY")
	for _, r := range list.UnderscoreEmbedded.Elements {
		delay := "-"
		if r.Lag != nil && *r.Lag > 0 {
			delay = fmt.Sprintf("%d days", *r.Lag)
		}
		rType := "-"
		if r.Type != nil {
			rType = string(*r.Type)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\n", dInt(r.Id), rType, delay)
	}
	return w.Flush()
}

func outputRelation(r *openproject.RelationReadModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", dInt(r.Id))
	if r.Type != nil {
		fmt.Fprintf(outputWriter, "Type: %s\n", *r.Type)
	}
	if r.Description != nil {
		fmt.Fprintf(outputWriter, "Description: %s\n", *r.Description)
	}
	if r.Lag != nil && *r.Lag > 0 {
		fmt.Fprintf(outputWriter, "Delay: %d days\n", *r.Lag)
	}
	return nil
}

// --- Status output ---

func outputStatusList(list *openproject.StatusCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tDEFAULT\tCLOSED\tREADONLY")
	for _, s := range list.UnderscoreEmbedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%t\t%t\t%t\n", dInt(s.Id), dStr(s.Name), dBool(s.IsDefault), dBool(s.IsClosed), dBool(s.IsReadonly))
	}
	return w.Flush()
}

func outputStatus(s *openproject.StatusModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", dInt(s.Id))
	fmt.Fprintf(outputWriter, "Name: %s\n", dStr(s.Name))
	if s.Position != nil {
		fmt.Fprintf(outputWriter, "Position: %d\n", *s.Position)
	}
	fmt.Fprintf(outputWriter, "Default: %t\n", dBool(s.IsDefault))
	fmt.Fprintf(outputWriter, "Closed: %t\n", dBool(s.IsClosed))
	fmt.Fprintf(outputWriter, "Readonly: %t\n", dBool(s.IsReadonly))
	if s.Color != nil {
		fmt.Fprintf(outputWriter, "Color: %s\n", *s.Color)
	}
	return nil
}

// --- Type output ---

func outputTypeList(list *openproject.TypesByWorkspaceModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tDEFAULT\tMILESTONE")
	if list.UnderscoreEmbedded.Elements != nil {
		for _, t := range *list.UnderscoreEmbedded.Elements {
			fmt.Fprintf(w, "%d\t%s\t%t\t%t\n", dInt(t.Id), dStr(t.Name), dBool(t.IsDefault), dBool(t.IsMilestone))
		}
	}
	return w.Flush()
}

func outputType(t *openproject.TypeModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", dInt(t.Id))
	fmt.Fprintf(outputWriter, "Name: %s\n", dStr(t.Name))
	if t.Position != nil {
		fmt.Fprintf(outputWriter, "Position: %d\n", *t.Position)
	}
	fmt.Fprintf(outputWriter, "Default: %t\n", dBool(t.IsDefault))
	fmt.Fprintf(outputWriter, "Milestone: %t\n", dBool(t.IsMilestone))
	if t.Color != nil {
		fmt.Fprintf(outputWriter, "Color: %s\n", *t.Color)
	}
	return nil
}

// --- Priority output ---

func outputPriorityList(list *openproject.PriorityCollectionModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tDEFAULT")
	for _, p := range list.UnderscoreEmbedded.Elements {
		fmt.Fprintf(w, "%d\t%s\t%t\n", dInt(p.Id), dStr(p.Name), dBool(p.IsDefault))
	}
	return w.Flush()
}

func outputPriority(p *openproject.PriorityModel) error {
	fmt.Fprintf(outputWriter, "ID: %d\n", dInt(p.Id))
	fmt.Fprintf(outputWriter, "Name: %s\n", dStr(p.Name))
	if p.Position != nil {
		fmt.Fprintf(outputWriter, "Position: %d\n", *p.Position)
	}
	fmt.Fprintf(outputWriter, "Default: %t\n", dBool(p.IsDefault))
	return nil
}

// --- Role output ---

func outputRoleList(roles []openproject.RoleModel) error {
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME")
	for _, r := range roles {
		fmt.Fprintf(w, "%d\t%s\n", dInt(r.Id), r.Name)
	}
	return w.Flush()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
