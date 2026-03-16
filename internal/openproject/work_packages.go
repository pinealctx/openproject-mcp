package openproject

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ListWorkPackagesOptions contains options for listing work packages.
type ListWorkPackagesOptions struct {
	Offset     int
	PageSize   int
	SortBy     string
	Select     []string
	Filters    []WorkPackageFilter
	RawFilters string // overrides Filters when non-empty
}

// WorkPackageFilter represents a filter for work packages.
type WorkPackageFilter struct {
	Name     string `json:"name"`
	Values   []any  `json:"values"`
	Operator string `json:"operator,omitempty"`
}

// ListWorkPackages retrieves work packages with optional filters.
func (c *Client) ListWorkPackages(ctx context.Context, opts *ListWorkPackagesOptions) (*WorkPackageList, error) {
	if opts == nil {
		opts = &ListWorkPackagesOptions{}
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
	if opts.RawFilters != "" {
		params.Set("filters", opts.RawFilters)
	} else if len(opts.Filters) > 0 {
		filterJSON, err := jsonMarshalFilters(opts.Filters)
		if err != nil {
			return nil, err
		}
		params.Set("filters", filterJSON)
	}

	path := "/work_packages"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result WorkPackageList
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListProjectWorkPackages retrieves work packages for a specific project.
func (c *Client) ListProjectWorkPackages(ctx context.Context, projectID int, opts *ListWorkPackagesOptions) (*WorkPackageList, error) {
	if opts == nil {
		opts = &ListWorkPackagesOptions{}
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
	if opts.RawFilters != "" {
		params.Set("filters", opts.RawFilters)
	} else if len(opts.Filters) > 0 {
		filterJSON, err := jsonMarshalFilters(opts.Filters)
		if err != nil {
			return nil, err
		}
		params.Set("filters", filterJSON)
	}

	path := fmt.Sprintf("/projects/%d/work_packages", projectID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result WorkPackageList
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWorkPackage retrieves a specific work package by ID.
func (c *Client) GetWorkPackage(ctx context.Context, id int) (*WorkPackage, error) {
	var wp WorkPackage
	if err := c.Get(ctx, fmt.Sprintf("/work_packages/%d", id), &wp); err != nil {
		return nil, err
	}
	return &wp, nil
}

// CreateWorkPackageOptions contains options for creating a work package.
type CreateWorkPackageOptions struct {
	Subject       string                  `json:"subject"`
	Description   *RichText               `json:"description,omitempty"`
	Type          string                  `json:"_type,omitempty"`
	StartDate     string                  `json:"startDate,omitempty"`
	DueDate       string                  `json:"dueDate,omitempty"`
	EstimatedTime string                  `json:"estimatedTime,omitempty"`
	Assignee      *WorkPackageLink        `json:"assignee,omitempty"`
	Project       *WorkPackageLink        `json:"project,omitempty"`
	Status        *WorkPackageLink        `json:"status,omitempty"`
	Priority      *WorkPackageLink        `json:"priority,omitempty"`
	Version       *WorkPackageLink        `json:"version,omitempty"`
	Links         *CreateWorkPackageLinks `json:"_links,omitempty"`
}

// CreateWorkPackageLinks contains links for creating a work package.
type CreateWorkPackageLinks struct {
	Type     *WorkPackageLink `json:"type,omitempty"`
	Status   *WorkPackageLink `json:"status,omitempty"`
	Priority *WorkPackageLink `json:"priority,omitempty"`
	Project  *WorkPackageLink `json:"project,omitempty"`
	Assignee *WorkPackageLink `json:"assignee,omitempty"`
	Version  *WorkPackageLink `json:"version,omitempty"`
	Parent   *WorkPackageLink `json:"parent,omitempty"`
}

// WorkPackageLink represents a link to another resource.
type WorkPackageLink struct {
	Href string `json:"href"`
}

// CreateWorkPackage creates a new work package.
func (c *Client) CreateWorkPackage(ctx context.Context, projectID int, opts *CreateWorkPackageOptions) (*WorkPackage, error) {
	// Ensure project link is set
	if opts.Links == nil {
		opts.Links = &CreateWorkPackageLinks{}
	}
	if opts.Links.Project == nil {
		opts.Links.Project = &WorkPackageLink{
			Href: fmt.Sprintf("/api/v3/projects/%d", projectID),
		}
	}

	var wp WorkPackage
	if err := c.Post(ctx, "/work_packages", opts, &wp); err != nil {
		return nil, err
	}
	return &wp, nil
}

// UpdateWorkPackageOptions contains options for updating a work package.
type UpdateWorkPackageOptions struct {
	Subject        string                  `json:"subject,omitempty"`
	Description    *RichText               `json:"description,omitempty"`
	StartDate      string                  `json:"startDate,omitempty"`
	DueDate        string                  `json:"dueDate,omitempty"`
	EstimatedTime  string                  `json:"estimatedTime,omitempty"`
	PercentageDone *int                    `json:"percentageDone,omitempty"`
	LockVersion    int                     `json:"lockVersion"` // always sent; 0 is valid
	Links          *UpdateWorkPackageLinks `json:"_links,omitempty"`
}

// UpdateWorkPackageLinks contains links for updating a work package.
type UpdateWorkPackageLinks struct {
	Status   *WorkPackageLink `json:"status,omitempty"`
	Priority *WorkPackageLink `json:"priority,omitempty"`
	Assignee *WorkPackageLink `json:"assignee,omitempty"`
	Version  *WorkPackageLink `json:"version,omitempty"`
	Parent   *WorkPackageLink `json:"parent,omitempty"`
}

// UpdateWorkPackage updates an existing work package. It automatically fetches
// the current lockVersion to satisfy OpenProject's optimistic concurrency check.
func (c *Client) UpdateWorkPackage(ctx context.Context, id int, opts *UpdateWorkPackageOptions) (*WorkPackage, error) {
	current, err := c.GetWorkPackage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current work package for lockVersion: %w", err)
	}
	opts.LockVersion = current.LockVersion

	var wp WorkPackage
	if err := c.Patch(ctx, fmt.Sprintf("/work_packages/%d", id), opts, &wp); err != nil {
		return nil, err
	}
	return &wp, nil
}

// DeleteWorkPackage deletes a work package.
func (c *Client) DeleteWorkPackage(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/work_packages/%d", id))
}

// ListTypes retrieves all work package types.
func (c *Client) ListTypes(ctx context.Context) (*TypeList, error) {
	var result TypeList
	if err := c.Get(ctx, "/types", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListProjectTypes retrieves work package types for a project.
func (c *Client) ListProjectTypes(ctx context.Context, projectID int) (*TypeList, error) {
	var result TypeList
	if err := c.Get(ctx, fmt.Sprintf("/projects/%d/types", projectID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListStatuses retrieves all statuses.
func (c *Client) ListStatuses(ctx context.Context) (*StatusList, error) {
	var result StatusList
	if err := c.Get(ctx, "/statuses", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListPriorities retrieves all priorities.
func (c *Client) ListPriorities(ctx context.Context) (*PriorityList, error) {
	var result PriorityList
	if err := c.Get(ctx, "/priorities", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// jsonMarshalFilters marshals filters to JSON string.
func jsonMarshalFilters(filters []WorkPackageFilter) (string, error) {
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

// normalizeSortBy converts shorthand syntax like "updatedAt:desc" into
// OpenProject's JSON sort format: [["updatedAt","desc"]].
func normalizeSortBy(sortBy string) string {
	s := strings.TrimSpace(sortBy)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "[") {
		return s
	}

	parts := strings.Split(s, ",")
	encoded := make([][]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		field := p
		order := "asc"
		if idx := strings.LastIndex(p, ":"); idx > 0 {
			field = strings.TrimSpace(p[:idx])
			candidate := strings.ToLower(strings.TrimSpace(p[idx+1:]))
			if candidate == "asc" || candidate == "desc" {
				order = candidate
			}
		}
		if field != "" {
			encoded = append(encoded, []string{field, order})
		}
	}
	if len(encoded) == 0 {
		return s
	}
	b, err := json.Marshal(encoded)
	if err != nil {
		return s
	}
	return string(b)
}
