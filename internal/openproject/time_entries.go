package openproject

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ListTimeEntriesOptions contains options for listing time entries.
type ListTimeEntriesOptions struct {
	Offset   int
	PageSize int
	SortBy   string
	Select   []string
	Filters  []TimeEntryFilter
}

// TimeEntryFilter represents a filter for time entries.
type TimeEntryFilter struct {
	Name     string   `json:"name"`
	Values   []string `json:"values"`
	Operator string   `json:"operator,omitempty"`
}

// ListTimeEntries retrieves all time entries.
func (c *Client) ListTimeEntries(ctx context.Context, opts *ListTimeEntriesOptions) (*TimeEntryList, error) {
	if opts == nil {
		opts = &ListTimeEntriesOptions{}
	}

	params := url.Values{}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(opts.PageSize))
	}
	if opts.SortBy != "" {
		params.Set("sortBy", normalizeSortBy(opts.SortBy))
	}
	if len(opts.Select) > 0 {
		params.Set("select", strings.Join(opts.Select, ","))
	}
	if len(opts.Filters) > 0 {
		filterJSON, err := jsonMarshalTimeEntryFilters(opts.Filters)
		if err != nil {
			return nil, err
		}
		params.Set("filters", filterJSON)
	}

	path := "/time_entries"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result TimeEntryList
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateTimeEntryOptions contains options for creating a time entry.
type CreateTimeEntryOptions struct {
	Hours       string `json:"hours"`
	Comment     string `json:"comment,omitempty"`
	SpentOn     string `json:"spentOn,omitempty"`
	ProjectID   int    `json:"-"`
	WorkPackage int    `json:"-"`
	ActivityID  int    `json:"-"`
	UserID      int    `json:"-"`
}

// CreateTimeEntry creates a new time entry.
func (c *Client) CreateTimeEntry(ctx context.Context, opts *CreateTimeEntryOptions) (*TimeEntry, error) {
	payload := map[string]interface{}{
		"hours":   opts.Hours,
		"comment": opts.Comment,
		"spentOn": opts.SpentOn,
		"_links":  map[string]interface{}{},
	}

	links := payload["_links"].(map[string]interface{})
	if opts.ProjectID > 0 {
		links["project"] = map[string]string{
			"href": fmt.Sprintf("/api/v3/projects/%d", opts.ProjectID),
		}
	}
	if opts.WorkPackage > 0 {
		links["workPackage"] = map[string]string{
			"href": fmt.Sprintf("/api/v3/work_packages/%d", opts.WorkPackage),
		}
	}
	if opts.ActivityID > 0 {
		links["activity"] = map[string]string{
			"href": fmt.Sprintf("/api/v3/time_entry_activities/%d", opts.ActivityID),
		}
	}
	if opts.UserID > 0 {
		links["user"] = map[string]string{
			"href": fmt.Sprintf("/api/v3/users/%d", opts.UserID),
		}
	}

	var entry TimeEntry
	if err := c.Post(ctx, "/time_entries", payload, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// UpdateTimeEntryOptions contains options for updating a time entry.
type UpdateTimeEntryOptions struct {
	Hours      string `json:"hours,omitempty"`
	Comment    string `json:"comment,omitempty"`
	SpentOn    string `json:"spentOn,omitempty"`
	ActivityID int    `json:"-"`
}

// UpdateTimeEntry updates an existing time entry.
func (c *Client) UpdateTimeEntry(ctx context.Context, id int, opts *UpdateTimeEntryOptions) (*TimeEntry, error) {
	payload := map[string]interface{}{}

	if opts.Hours != "" {
		payload["hours"] = opts.Hours
	}
	if opts.Comment != "" {
		payload["comment"] = opts.Comment
	}
	if opts.SpentOn != "" {
		payload["spentOn"] = opts.SpentOn
	}
	if opts.ActivityID > 0 {
		payload["_links"] = map[string]interface{}{
			"activity": map[string]string{
				"href": fmt.Sprintf("/api/v3/time_entry_activities/%d", opts.ActivityID),
			},
		}
	}

	var entry TimeEntry
	if err := c.Patch(ctx, fmt.Sprintf("/time_entries/%d", id), payload, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// DeleteTimeEntry deletes a time entry.
func (c *Client) DeleteTimeEntry(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/time_entries/%d", id))
}

// ListTimeEntryActivities retrieves all time entry activities.
func (c *Client) ListTimeEntryActivities(ctx context.Context) (*TimeEntryActivityList, error) {
	var result TimeEntryActivityList
	if err := c.Get(ctx, "/time_entry_activities", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// jsonMarshalTimeEntryFilters marshals time entry filters to JSON string.
func jsonMarshalTimeEntryFilters(filters []TimeEntryFilter) (string, error) {
	encoded := make([]map[string]map[string]interface{}, 0, len(filters))
	for _, f := range filters {
		if f.Name == "" {
			continue
		}
		op := f.Operator
		if op == "" {
			op = "="
		}
		encoded = append(encoded, map[string]map[string]interface{}{
			f.Name: {
				"operator": op,
				"values":   f.Values,
			},
		})
	}

	data, err := json.Marshal(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
