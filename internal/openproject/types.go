// Package openproject provides a client for interacting with the OpenProject API.
package openproject

import (
	"encoding/json"
	"time"
)

// RichText represents a formatted text field. OpenProject returns these as
// {"format":"markdown","raw":"...","html":"..."} in responses, but accepts
// plain strings in write requests. RichText.MarshalJSON serialises as a plain
// string so it can be safely used in POST/PATCH bodies.
type RichText struct {
	Raw string
}

func (r *RichText) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if b[0] == '{' {
		var obj struct {
			Raw string `json:"raw"`
		}
		if err := json.Unmarshal(b, &obj); err != nil {
			return err
		}
		r.Raw = obj.Raw
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	r.Raw = s
	return nil
}

func (r RichText) MarshalJSON() ([]byte, error) {
	if r.Raw == "" {
		return []byte("null"), nil
	}
	return json.Marshal(struct {
		Format string `json:"format"`
		Raw    string `json:"raw"`
	}{Format: "markdown", Raw: r.Raw})
}

func (r RichText) String() string {
	return r.Raw
}

// NewRichText creates a RichText from a plain markdown string.
// Returns nil when s is empty so omitempty skips the field.
func NewRichText(s string) *RichText {
	if s == "" {
		return nil
	}
	return &RichText{Raw: s}
}

// Pagination represents common pagination fields in API responses.
type Pagination struct {
	Offset int `json:"offset"`
	Total  int `json:"total"`
	Count  int `json:"count"`
}

// Link represents a HAL link.
type Link struct {
	Href  string `json:"href"`
	Title string `json:"title,omitempty"`
}

// Links represents a collection of HAL links.
type Links struct {
	Self        *Link `json:"self,omitempty"`
	Next        *Link `json:"next,omitempty"`
	Prev        *Link `json:"prev,omitempty"`
	First       *Link `json:"first,omitempty"`
	Last        *Link `json:"last,omitempty"`
	Parent      *Link `json:"parent,omitempty"`
	Project     *Link `json:"project,omitempty"`
	User        *Link `json:"user,omitempty"`
	Status      *Link `json:"status,omitempty"`
	Type        *Link `json:"type,omitempty"`
	Priority    *Link `json:"priority,omitempty"`
	Author      *Link `json:"author,omitempty"`
	Assignee    *Link `json:"assignee,omitempty"`
	Version     *Link `json:"version,omitempty"`
	WorkPackage *Link `json:"workPackage,omitempty"`
}

// Project represents an OpenProject project.
type Project struct {
	ID          int                    `json:"id"`
	Identifier  string                 `json:"identifier"`
	Name        string                 `json:"name"`
	Description RichText               `json:"description"`
	Active      bool                   `json:"active"`
	Public      bool                   `json:"public"`
	CreatedAt   *time.Time             `json:"createdAt"`
	UpdatedAt   *time.Time             `json:"updatedAt"`
	Links       *Links                 `json:"_links,omitempty"`
	Embedded    map[string]interface{} `json:"_embedded,omitempty"`
}

// ProjectList represents a list of projects with pagination.
type ProjectList struct {
	Embedded struct {
		Elements []Project `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// WorkPackage represents an OpenProject work package (task/issue).
type WorkPackage struct {
	ID               int                    `json:"id"`
	Subject          string                 `json:"subject"`
	Description      RichText               `json:"description"`
	LockVersion      int                    `json:"lockVersion"`
	PercentageDone   int                    `json:"percentageDone,omitempty"`
	EstimatedTime    string                 `json:"estimatedTime,omitempty"`
	ScheduleManually bool                   `json:"scheduleManually,omitempty"`
	StartDate        string                 `json:"startDate,omitempty"`
	DueDate          string                 `json:"dueDate,omitempty"`
	CreatedAt        *time.Time             `json:"createdAt"`
	UpdatedAt        *time.Time             `json:"updatedAt"`
	Links            *WorkPackageLinks      `json:"_links,omitempty"`
	Embedded         map[string]interface{} `json:"_embedded,omitempty"`
}

// WorkPackageLinks represents links specific to work packages.
type WorkPackageLinks struct {
	Links
	Children    *Link   `json:"children,omitempty"`
	Relations   *Link   `json:"relations,omitempty"`
	Attachments *Link   `json:"attachments,omitempty"`
	Watchers    *Link   `json:"watchers,omitempty"`
	Categories  []*Link `json:"categories,omitempty"`
}

// WorkPackageList represents a list of work packages.
type WorkPackageList struct {
	Embedded struct {
		Elements []WorkPackage `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// User represents an OpenProject user.
type User struct {
	ID        int                    `json:"id"`
	Login     string                 `json:"login"`
	FirstName string                 `json:"firstName"`
	LastName  string                 `json:"lastName"`
	Name      string                 `json:"name"`
	Email     string                 `json:"email"`
	Admin     bool                   `json:"admin"`
	Avatar    string                 `json:"avatar"`
	Status    string                 `json:"status"`
	Language  string                 `json:"language"`
	CreatedAt *time.Time             `json:"createdAt"`
	UpdatedAt *time.Time             `json:"updatedAt"`
	Links     *Links                 `json:"_links,omitempty"`
	Embedded  map[string]interface{} `json:"_embedded,omitempty"`
}

// UserList represents a list of users.
type UserList struct {
	Embedded struct {
		Elements []User `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// Membership represents a project membership.
type Membership struct {
	ID        int                    `json:"id"`
	CreatedAt *time.Time             `json:"createdAt"`
	UpdatedAt *time.Time             `json:"updatedAt"`
	Links     *MembershipLinks       `json:"_links,omitempty"`
	Embedded  map[string]interface{} `json:"_embedded,omitempty"`
}

// MembershipLinks represents links specific to memberships.
type MembershipLinks struct {
	Links
	Project   *Link   `json:"project,omitempty"`
	Principal *Link   `json:"principal,omitempty"`
	Roles     []*Link `json:"roles,omitempty"`
}

// MembershipList represents a list of memberships.
type MembershipList struct {
	Embedded struct {
		Elements []Membership `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// Role represents a user role.
type Role struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Permissions []string               `json:"permissions"`
	CreatedAt   *time.Time             `json:"createdAt"`
	UpdatedAt   *time.Time             `json:"updatedAt"`
	Links       *Links                 `json:"_links,omitempty"`
	Embedded    map[string]interface{} `json:"_embedded,omitempty"`
}

// RoleList represents a list of roles.
type RoleList struct {
	Embedded struct {
		Elements []Role `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// TimeEntry represents a time tracking entry.
type TimeEntry struct {
	ID        int                    `json:"id"`
	Comment   string                 `json:"comment"`
	Hours     string                 `json:"hours"`
	SpentOn   string                 `json:"spentOn"`
	CreatedAt *time.Time             `json:"createdAt"`
	UpdatedAt *time.Time             `json:"updatedAt"`
	Links     *TimeEntryLinks        `json:"_links,omitempty"`
	Embedded  map[string]interface{} `json:"_embedded,omitempty"`
}

// TimeEntryLinks represents links specific to time entries.
type TimeEntryLinks struct {
	Links
	Project     *Link `json:"project,omitempty"`
	WorkPackage *Link `json:"workPackage,omitempty"`
	User        *Link `json:"user,omitempty"`
	Activity    *Link `json:"activity,omitempty"`
}

// TimeEntryList represents a list of time entries.
type TimeEntryList struct {
	Embedded struct {
		Elements []TimeEntry `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// TimeEntryActivity represents a time entry activity type.
type TimeEntryActivity struct {
	ID        int                    `json:"id"`
	Name      string                 `json:"name"`
	Position  int                    `json:"position"`
	IsDefault bool                   `json:"isDefault"`
	Active    bool                   `json:"active"`
	Links     *Links                 `json:"_links,omitempty"`
	Embedded  map[string]interface{} `json:"_embedded,omitempty"`
}

// TimeEntryActivityList represents a list of time entry activities.
type TimeEntryActivityList struct {
	Embedded struct {
		Elements []TimeEntryActivity `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// Version represents a project version/milestone.
type Version struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	StartDate   string                 `json:"startDate,omitempty"`
	EndDate     string                 `json:"endDate,omitempty"`
	CreatedAt   *time.Time             `json:"createdAt"`
	UpdatedAt   *time.Time             `json:"updatedAt"`
	Links       *Links                 `json:"_links,omitempty"`
	Embedded    map[string]interface{} `json:"_embedded,omitempty"`
}

// VersionList represents a list of versions.
type VersionList struct {
	Embedded struct {
		Elements []Version `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// Relation represents a relationship between work packages.
type Relation struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	ReverseType string                 `json:"reverseType"`
	Description string                 `json:"description"`
	Delay       int                    `json:"delay,omitempty"`
	CreatedAt   *time.Time             `json:"createdAt"`
	UpdatedAt   *time.Time             `json:"updatedAt"`
	Links       *RelationLinks         `json:"_links,omitempty"`
	Embedded    map[string]interface{} `json:"_embedded,omitempty"`
}

// RelationLinks represents links specific to relations.
type RelationLinks struct {
	Links
	From *Link `json:"from,omitempty"`
	To   *Link `json:"to,omitempty"`
}

// RelationList represents a list of relations.
type RelationList struct {
	Embedded struct {
		Elements []Relation `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// Status represents a work package status.
type Status struct {
	ID         int                    `json:"id"`
	Name       string                 `json:"name"`
	Position   int                    `json:"position"`
	IsDefault  bool                   `json:"isDefault"`
	IsClosed   bool                   `json:"isClosed"`
	IsReadonly bool                   `json:"isReadonly"`
	Color      string                 `json:"color"`
	CreatedAt  *time.Time             `json:"createdAt"`
	UpdatedAt  *time.Time             `json:"updatedAt"`
	Links      *Links                 `json:"_links,omitempty"`
	Embedded   map[string]interface{} `json:"_embedded,omitempty"`
}

// StatusList represents a list of statuses.
type StatusList struct {
	Embedded struct {
		Elements []Status `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// Type represents a work package type (e.g., Task, Bug, Feature).
type Type struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Position    int                    `json:"position"`
	IsDefault   bool                   `json:"isDefault"`
	IsMilestone bool                   `json:"isMilestone"`
	Color       string                 `json:"color"`
	CreatedAt   *time.Time             `json:"createdAt"`
	UpdatedAt   *time.Time             `json:"updatedAt"`
	Links       *Links                 `json:"_links,omitempty"`
	Embedded    map[string]interface{} `json:"_embedded,omitempty"`
}

// TypeList represents a list of types.
type TypeList struct {
	Embedded struct {
		Elements []Type `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// Priority represents a work package priority.
type Priority struct {
	ID        int                    `json:"id"`
	Name      string                 `json:"name"`
	Position  int                    `json:"position"`
	IsDefault bool                   `json:"isDefault"`
	Color     string                 `json:"color"`
	CreatedAt *time.Time             `json:"createdAt"`
	UpdatedAt *time.Time             `json:"updatedAt"`
	Links     *Links                 `json:"_links,omitempty"`
	Embedded  map[string]interface{} `json:"_embedded,omitempty"`
}

// PriorityList represents a list of priorities.
type PriorityList struct {
	Embedded struct {
		Elements []Priority `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// Permission represents user permissions.
type Permission struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Action []string `json:"action"`
}

// APIError represents an error from the OpenProject API.
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	ErrorID    string `json:"errorIdentifier,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

// ---------- Grid / Board types ----------

// GridWidgetLinks holds links for a board widget.
type GridWidgetLinks struct {
	Query *Link `json:"query,omitempty"`
}

// GridWidget represents a single widget (column) on a board.
type GridWidget struct {
	ID          int              `json:"id,omitempty"`
	Identifier  string           `json:"identifier"`
	StartRow    int              `json:"startRow"`
	EndRow      int              `json:"endRow"`
	StartColumn int              `json:"startColumn"`
	EndColumn   int              `json:"endColumn"`
	Options     map[string]any   `json:"options,omitempty"`
	Links       *GridWidgetLinks `json:"_links,omitempty"`
}

// GridLinks holds HAL links for a grid.
type GridLinks struct {
	Self      *Link `json:"self,omitempty"`
	Scope     *Link `json:"scope,omitempty"`
	AddWidget *Link `json:"addWidget,omitempty"`
}

// GridEmbedded holds the embedded resources for a grid.
type GridEmbedded struct {
	Widgets []GridWidget `json:"widgets"`
}

// Grid represents an OpenProject grid (board).
type Grid struct {
	ID          int           `json:"id"`
	RowCount    int           `json:"rowCount"`
	ColumnCount int           `json:"columnCount"`
	CreatedAt   *time.Time    `json:"createdAt"`
	UpdatedAt   *time.Time    `json:"updatedAt"`
	Links       *GridLinks    `json:"_links,omitempty"`
	Embedded    *GridEmbedded `json:"_embedded,omitempty"`
}

// GridList represents a list of grids.
type GridList struct {
	Embedded struct {
		Elements []Grid `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}
