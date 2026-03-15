package openproject

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ListUsersOptions contains options for listing users.
type ListUsersOptions struct {
	Offset   int
	PageSize int
	OrderBy  string
	Select   []string
	Filters  []UserFilter
}

// UserFilter represents a filter for users.
type UserFilter struct {
	Name     string   `json:"name"`
	Values   []string `json:"values"`
	Operator string   `json:"operator,omitempty"`
}

// ListUsers retrieves all users.
func (c *Client) ListUsers(ctx context.Context, opts *ListUsersOptions) (*UserList, error) {
	if opts == nil {
		opts = &ListUsersOptions{}
	}

	params := url.Values{}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(opts.PageSize))
	}
	if opts.OrderBy != "" {
		params.Set("orderBy", opts.OrderBy)
	}
	if len(opts.Select) > 0 {
		params.Set("select", strings.Join(opts.Select, ","))
	}
	if len(opts.Filters) > 0 {
		filterJSON, err := jsonMarshalUserFilters(opts.Filters)
		if err != nil {
			return nil, err
		}
		params.Set("filters", filterJSON)
	}

	path := "/users"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result UserList
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUser retrieves a specific user by ID.
func (c *Client) GetUser(ctx context.Context, id int) (*User, error) {
	var user User
	if err := c.Get(ctx, fmt.Sprintf("/users/%d", id), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// jsonMarshalUserFilters marshals user filters to JSON string.
func jsonMarshalUserFilters(filters []UserFilter) (string, error) {
	data, err := json.Marshal(filters)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
