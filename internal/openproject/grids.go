package openproject

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListGridsOptions contains options for listing grids.
type ListGridsOptions struct {
	Offset   int
	PageSize int
	Filters  string // raw JSON filter string
}

// CreateGridOptions contains options for creating a grid (board).
type CreateGridOptions struct {
	RowCount    int           `json:"rowCount"`
	ColumnCount int           `json:"columnCount"`
	Embedded    *GridEmbedded `json:"_embedded,omitempty"`
	Links       *GridLinks    `json:"_links,omitempty"`
}

// UpdateGridOptions contains options for updating a grid.
type UpdateGridOptions struct {
	RowCount    int           `json:"rowCount,omitempty"`
	ColumnCount int           `json:"columnCount,omitempty"`
	Embedded    *GridEmbedded `json:"_embedded,omitempty"`
}

// ListGrids retrieves all grids, optionally filtered.
func (c *Client) ListGrids(ctx context.Context, opts *ListGridsOptions) (*GridList, error) {
	if opts == nil {
		opts = &ListGridsOptions{}
	}

	params := url.Values{}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(opts.PageSize))
	}
	if opts.Filters != "" {
		params.Set("filters", opts.Filters)
	}

	path := "/grids"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result GridList
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGrid retrieves a specific grid by ID.
func (c *Client) GetGrid(ctx context.Context, id int) (*Grid, error) {
	var grid Grid
	if err := c.Get(ctx, fmt.Sprintf("/grids/%d", id), &grid); err != nil {
		return nil, err
	}
	return &grid, nil
}

// CreateGrid creates a new grid (board) scoped to a project.
func (c *Client) CreateGrid(ctx context.Context, projectID int, opts *CreateGridOptions) (*Grid, error) {
	if opts == nil {
		opts = &CreateGridOptions{}
	}
	if opts.RowCount == 0 {
		opts.RowCount = 1
	}
	if opts.ColumnCount == 0 {
		opts.ColumnCount = 3
	}
	if opts.Links == nil {
		opts.Links = &GridLinks{}
	}
	opts.Links.Scope = &Link{Href: fmt.Sprintf("/api/v3/projects/%d", projectID)}

	var grid Grid
	if err := c.Post(ctx, "/grids", opts, &grid); err != nil {
		return nil, err
	}
	return &grid, nil
}

// UpdateGrid updates an existing grid.
func (c *Client) UpdateGrid(ctx context.Context, id int, opts *UpdateGridOptions) (*Grid, error) {
	var grid Grid
	if err := c.Patch(ctx, fmt.Sprintf("/grids/%d", id), opts, &grid); err != nil {
		return nil, err
	}
	return &grid, nil
}

// DeleteGrid deletes a grid.
func (c *Client) DeleteGrid(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/grids/%d", id))
}

// AddGridWidget appends a widget to a grid by fetching current state and patching.
func (c *Client) AddGridWidget(ctx context.Context, gridID int, widget GridWidget) (*Grid, error) {
	current, err := c.GetGrid(ctx, gridID)
	if err != nil {
		return nil, err
	}
	widgets := []GridWidget{}
	if current.Embedded != nil {
		widgets = current.Embedded.Widgets
	}
	widgets = append(widgets, widget)
	return c.UpdateGrid(ctx, gridID, &UpdateGridOptions{
		Embedded: &GridEmbedded{Widgets: widgets},
	})
}

// RemoveGridWidget removes a widget from a grid by ID.
func (c *Client) RemoveGridWidget(ctx context.Context, gridID, widgetID int) (*Grid, error) {
	current, err := c.GetGrid(ctx, gridID)
	if err != nil {
		return nil, err
	}
	var widgets []GridWidget
	if current.Embedded != nil {
		for _, w := range current.Embedded.Widgets {
			if w.ID != widgetID {
				widgets = append(widgets, w)
			}
		}
	}
	return c.UpdateGrid(ctx, gridID, &UpdateGridOptions{
		Embedded: &GridEmbedded{Widgets: widgets},
	})
}
