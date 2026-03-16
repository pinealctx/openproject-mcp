package openproject

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ListProjectsOptions contains options for listing projects.
type ListProjectsOptions struct {
	Offset     int
	PageSize   int
	SortBy     string
	Select     []string
	Filters    []ProjectFilter
	RawFilters string // overrides Filters when non-empty
	ShowRoot   bool
}

// ProjectFilter represents a filter for projects.
type ProjectFilter struct {
	Name     string   `json:"name"`
	Values   []string `json:"values"`
	Operator string   `json:"operator,omitempty"`
}

// ListProjects retrieves all projects.
func (c *Client) ListProjects(ctx context.Context, opts *ListProjectsOptions) (*ProjectList, error) {
	if opts == nil {
		opts = &ListProjectsOptions{}
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
		filterJSON, err := jsonMarshalProjectFilters(opts.Filters)
		if err != nil {
			return nil, err
		}
		params.Set("filters", filterJSON)
	}
	if opts.ShowRoot {
		params.Set("showRoot", "true")
	}

	path := "/projects"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result ProjectList
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// jsonMarshalProjectFilters marshals project filters to the API query format.
func jsonMarshalProjectFilters(filters []ProjectFilter) (string, error) {
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

// GetProject retrieves a specific project by ID.
func (c *Client) GetProject(ctx context.Context, id int) (*Project, error) {
	var project Project
	if err := c.Get(ctx, fmt.Sprintf("/projects/%d", id), &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProjectByIdentifier retrieves a project by its identifier.
func (c *Client) GetProjectByIdentifier(ctx context.Context, identifier string) (*Project, error) {
	var project Project
	if err := c.Get(ctx, fmt.Sprintf("/projects/%s", identifier), &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProjectOptions contains options for creating a project.
type CreateProjectOptions struct {
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
	Public      bool   `json:"public,omitempty"`
	ParentID    *int   `json:"parentId,omitempty"`
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, opts *CreateProjectOptions) (*Project, error) {
	var project Project
	if err := c.Post(ctx, "/projects", opts, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProjectOptions contains options for updating a project.
type UpdateProjectOptions struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Public      *bool  `json:"public,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

// UpdateProject updates an existing project.
func (c *Client) UpdateProject(ctx context.Context, id int, opts *UpdateProjectOptions) (*Project, error) {
	var project Project
	if err := c.Patch(ctx, fmt.Sprintf("/projects/%d", id), opts, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject deletes a project.
func (c *Client) DeleteProject(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/projects/%d", id))
}
