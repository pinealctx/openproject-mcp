package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	external "github.com/pinealctx/openproject"
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
	case *external.ProjectCollectionModel:
		return outputProjectList(v)
	case *external.ProjectModel:
		return outputProject(v)
	case *external.WorkPackagesModel:
		return outputWorkPackageList(v)
	case *external.WorkPackageModel:
		return outputWorkPackage(v)
	case *external.UserCollectionModel:
		return outputUserList(v)
	case *external.UserModel:
		return outputUser(v)
	case *external.MembershipCollectionModel:
		return outputMembershipList(v)
	case *external.MembershipReadModel:
		return outputMembership(v)
	case *external.VersionCollectionModel:
		return outputVersionList(v)
	case *external.VersionReadModel:
		return outputVersion(v)
	case *external.NotificationCollectionModel:
		return outputNotificationList(v)
	case *external.NotificationModel:
		return outputNotification(v)
	case *external.RelationCollectionModel:
		return outputRelationList(v)
	case *external.RelationReadModel:
		return outputRelation(v)
	case *external.StatusCollectionModel:
		return outputStatusList(v)
	case *external.StatusModel:
		return outputStatus(v)
	case *external.TypesByWorkspaceModel:
		return outputTypeList(v)
	case *external.TypeModel:
		return outputType(v)
	case *external.PriorityCollectionModel:
		return outputPriorityList(v)
	case *external.PriorityModel:
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

func outputProjectList(list *external.ProjectCollectionModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tIDENTIFIER\tNAME\tACTIVE\tPUBLIC")
	for _, p := range list.UnderscoreEmbedded.Elements {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%t\t%t\n", dInt(p.Id), dStr(p.Identifier), dStr(p.Name), dBool(p.Active), dBool(p.Public))
	}
	return w.Flush()
}

func outputProject(p *external.ProjectModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", dInt(p.Id))
	_, _ = fmt.Fprintf(outputWriter, "Identifier: %s\n", dStr(p.Identifier))
	_, _ = fmt.Fprintf(outputWriter, "Name: %s\n", dStr(p.Name))
	if p.Description != nil && p.Description.Raw != nil {
		_, _ = fmt.Fprintf(outputWriter, "Description: %s\n", *p.Description.Raw)
	}
	_, _ = fmt.Fprintf(outputWriter, "Active: %t\n", dBool(p.Active))
	_, _ = fmt.Fprintf(outputWriter, "Public: %t\n", dBool(p.Public))
	if p.CreatedAt != nil {
		_, _ = fmt.Fprintf(outputWriter, "Created: %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if p.UpdatedAt != nil {
		_, _ = fmt.Fprintf(outputWriter, "Updated: %s\n", p.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Work Package output ---

func outputWorkPackageList(list *external.WorkPackagesModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tSUBJECT\tTYPE\tSTATUS\tASSIGNEE")
	for _, wp := range list.UnderscoreEmbedded.Elements {
		assignee := "-"
		if wp.UnderscoreLinks.Assignee != nil {
			assignee = dStr(wp.UnderscoreLinks.Assignee.Title)
		}
		typeName := dStr(wp.UnderscoreLinks.Type.Title)
		statusName := dStr(wp.UnderscoreLinks.Status.Title)
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", dInt(wp.Id), truncate(wp.Subject, 40), typeName, statusName, assignee)
	}
	return w.Flush()
}

func outputWorkPackage(wp *external.WorkPackageModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", dInt(wp.Id))
	_, _ = fmt.Fprintf(outputWriter, "Subject: %s\n", wp.Subject)
	if wp.Description != nil && wp.Description.Raw != nil {
		_, _ = fmt.Fprintf(outputWriter, "Description: %s\n", *wp.Description.Raw)
	}
	_, _ = fmt.Fprintf(outputWriter, "Type: %s\n", dStr(wp.UnderscoreLinks.Type.Title))
	_, _ = fmt.Fprintf(outputWriter, "Status: %s\n", dStr(wp.UnderscoreLinks.Status.Title))
	_, _ = fmt.Fprintf(outputWriter, "Priority: %s\n", dStr(wp.UnderscoreLinks.Priority.Title))
	if wp.UnderscoreLinks.Assignee != nil {
		_, _ = fmt.Fprintf(outputWriter, "Assignee: %s\n", dStr(wp.UnderscoreLinks.Assignee.Title))
	}
	if wp.PercentageDone != nil {
		_, _ = fmt.Fprintf(outputWriter, "Progress: %d%%\n", *wp.PercentageDone)
	}
	if wp.EstimatedTime != nil {
		_, _ = fmt.Fprintf(outputWriter, "Estimated: %s\n", *wp.EstimatedTime)
	}
	if wp.StartDate != nil {
		_, _ = fmt.Fprintf(outputWriter, "Start: %s\n", wp.StartDate.String())
	}
	if wp.DueDate != nil {
		_, _ = fmt.Fprintf(outputWriter, "Due: %s\n", wp.DueDate.String())
	}
	if wp.CreatedAt != nil {
		_, _ = fmt.Fprintf(outputWriter, "Created: %s\n", wp.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- User output ---

func outputUserList(list *external.UserCollectionModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tNAME\tEMAIL\tSTATUS")
	for _, u := range list.UnderscoreEmbedded.Elements {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", u.Id, u.Name, dStr(u.Email), dStr(u.Status))
	}
	return w.Flush()
}

func outputUser(u *external.UserModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", u.Id)
	_, _ = fmt.Fprintf(outputWriter, "Name: %s\n", u.Name)
	if u.Login != nil {
		_, _ = fmt.Fprintf(outputWriter, "Login: %s\n", *u.Login)
	}
	if u.Email != nil {
		_, _ = fmt.Fprintf(outputWriter, "Email: %s\n", *u.Email)
	}
	if u.Admin != nil {
		_, _ = fmt.Fprintf(outputWriter, "Admin: %t\n", *u.Admin)
	}
	if u.Status != nil {
		_, _ = fmt.Fprintf(outputWriter, "Status: %s\n", *u.Status)
	}
	if u.Language != nil {
		_, _ = fmt.Fprintf(outputWriter, "Language: %s\n", *u.Language)
	}
	if u.CreatedAt != nil {
		_, _ = fmt.Fprintf(outputWriter, "Created: %s\n", u.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Membership output ---

func outputMembershipList(list *external.MembershipCollectionModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tROLES")
	for _, m := range list.UnderscoreEmbedded.Elements {
		roles := "-"
		if len(m.UnderscoreLinks.Roles) > 0 {
			names := make([]string, len(m.UnderscoreLinks.Roles))
			for i, r := range m.UnderscoreLinks.Roles {
				names[i] = dStr(r.Title)
			}
			roles = strings.Join(names, ", ")
		}
		_, _ = fmt.Fprintf(w, "%d\t%s\n", m.Id, roles)
	}
	return w.Flush()
}

func outputMembership(m *external.MembershipReadModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", m.Id)
	roles := "-"
	if len(m.UnderscoreLinks.Roles) > 0 {
		names := make([]string, len(m.UnderscoreLinks.Roles))
		for i, r := range m.UnderscoreLinks.Roles {
			names[i] = dStr(r.Title)
		}
		roles = strings.Join(names, ", ")
	}
	_, _ = fmt.Fprintf(outputWriter, "Roles: %s\n", roles)
	_, _ = fmt.Fprintf(outputWriter, "Created: %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))
	return nil
}

// --- Version output ---

func outputVersionList(list *external.VersionCollectionModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tSTART\tEND")
	for _, v := range list.UnderscoreEmbedded.Elements {
		start := "-"
		if v.StartDate != nil {
			start = v.StartDate.String()
		}
		end := "-"
		if v.EndDate != nil {
			end = v.EndDate.String()
		}
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", v.Id, v.Name, v.Status, start, end)
	}
	return w.Flush()
}

func outputVersion(v *external.VersionReadModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", v.Id)
	_, _ = fmt.Fprintf(outputWriter, "Name: %s\n", v.Name)
	_, _ = fmt.Fprintf(outputWriter, "Status: %s\n", v.Status)
	if v.StartDate != nil {
		_, _ = fmt.Fprintf(outputWriter, "Start: %s\n", v.StartDate.String())
	}
	if v.EndDate != nil {
		_, _ = fmt.Fprintf(outputWriter, "End: %s\n", v.EndDate.String())
	}
	_, _ = fmt.Fprintf(outputWriter, "Created: %s\n", v.CreatedAt.Format("2006-01-02 15:04:05"))
	return nil
}

// --- Notification output ---

func outputNotificationList(list *external.NotificationCollectionModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tREASON\tREAD\tCREATED")
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
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", id, reason, read, created)
	}
	return w.Flush()
}

func outputNotification(n *external.NotificationModel) error {
	if n.Id != nil {
		_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", *n.Id)
	}
	if n.Reason != nil {
		_, _ = fmt.Fprintf(outputWriter, "Reason: %s\n", *n.Reason)
	}
	read := "No"
	if n.ReadIAN != nil && *n.ReadIAN {
		read = "Yes"
	}
	_, _ = fmt.Fprintf(outputWriter, "Read: %s\n", read)
	if n.CreatedAt != nil {
		_, _ = fmt.Fprintf(outputWriter, "Created: %s\n", n.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// --- Relation output ---

func outputRelationList(list *external.RelationCollectionModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tTYPE\tDELAY")
	for _, r := range list.UnderscoreEmbedded.Elements {
		delay := "-"
		if r.Lag != nil && *r.Lag > 0 {
			delay = fmt.Sprintf("%d days", *r.Lag)
		}
		rType := "-"
		if r.Type != nil {
			rType = string(*r.Type)
		}
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\n", dInt(r.Id), rType, delay)
	}
	return w.Flush()
}

func outputRelation(r *external.RelationReadModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", dInt(r.Id))
	if r.Type != nil {
		_, _ = fmt.Fprintf(outputWriter, "Type: %s\n", *r.Type)
	}
	if r.Description != nil {
		_, _ = fmt.Fprintf(outputWriter, "Description: %s\n", *r.Description)
	}
	if r.Lag != nil && *r.Lag > 0 {
		_, _ = fmt.Fprintf(outputWriter, "Delay: %d days\n", *r.Lag)
	}
	return nil
}

// --- Status output ---

func outputStatusList(list *external.StatusCollectionModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tNAME\tDEFAULT\tCLOSED\tREADONLY")
	for _, s := range list.UnderscoreEmbedded.Elements {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%t\t%t\t%t\n", dInt(s.Id), dStr(s.Name), dBool(s.IsDefault), dBool(s.IsClosed), dBool(s.IsReadonly))
	}
	return w.Flush()
}

func outputStatus(s *external.StatusModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", dInt(s.Id))
	_, _ = fmt.Fprintf(outputWriter, "Name: %s\n", dStr(s.Name))
	if s.Position != nil {
		_, _ = fmt.Fprintf(outputWriter, "Position: %d\n", *s.Position)
	}
	_, _ = fmt.Fprintf(outputWriter, "Default: %t\n", dBool(s.IsDefault))
	_, _ = fmt.Fprintf(outputWriter, "Closed: %t\n", dBool(s.IsClosed))
	_, _ = fmt.Fprintf(outputWriter, "Readonly: %t\n", dBool(s.IsReadonly))
	if s.Color != nil {
		_, _ = fmt.Fprintf(outputWriter, "Color: %s\n", *s.Color)
	}
	return nil
}

// --- Type output ---

func outputTypeList(list *external.TypesByWorkspaceModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tNAME\tDEFAULT\tMILESTONE")
	if list.UnderscoreEmbedded.Elements != nil {
		for _, t := range *list.UnderscoreEmbedded.Elements {
			_, _ = fmt.Fprintf(w, "%d\t%s\t%t\t%t\n", dInt(t.Id), dStr(t.Name), dBool(t.IsDefault), dBool(t.IsMilestone))
		}
	}
	return w.Flush()
}

func outputType(t *external.TypeModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", dInt(t.Id))
	_, _ = fmt.Fprintf(outputWriter, "Name: %s\n", dStr(t.Name))
	if t.Position != nil {
		_, _ = fmt.Fprintf(outputWriter, "Position: %d\n", *t.Position)
	}
	_, _ = fmt.Fprintf(outputWriter, "Default: %t\n", dBool(t.IsDefault))
	_, _ = fmt.Fprintf(outputWriter, "Milestone: %t\n", dBool(t.IsMilestone))
	if t.Color != nil {
		_, _ = fmt.Fprintf(outputWriter, "Color: %s\n", *t.Color)
	}
	return nil
}

// --- Priority output ---

func outputPriorityList(list *external.PriorityCollectionModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tNAME\tDEFAULT")
	for _, p := range list.UnderscoreEmbedded.Elements {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%t\n", dInt(p.Id), dStr(p.Name), dBool(p.IsDefault))
	}
	return w.Flush()
}

func outputPriority(p *external.PriorityModel) error {
	_, _ = fmt.Fprintf(outputWriter, "ID: %d\n", dInt(p.Id))
	_, _ = fmt.Fprintf(outputWriter, "Name: %s\n", dStr(p.Name))
	if p.Position != nil {
		_, _ = fmt.Fprintf(outputWriter, "Position: %d\n", *p.Position)
	}
	_, _ = fmt.Fprintf(outputWriter, "Default: %t\n", dBool(p.IsDefault))
	return nil
}

// --- Role output ---

func outputRoleList(roles []external.RoleModel) error {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, "ID\tNAME")
	for _, r := range roles {
		_, _ = fmt.Fprintf(w, "%d\t%s\n", dInt(r.Id), r.Name)
	}
	return w.Flush()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
