package openproject

import (
	"context"
	"fmt"
)

// SearchResult represents a single search result.
type SearchResult struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"_type,omitempty"`
	Links       *Links                 `json:"_links,omitempty"`
	Embedded    map[string]interface{} `json:"_embedded,omitempty"`
}

// SearchResults represents a list of search results.
type SearchResults struct {
	Embedded struct {
		Elements []SearchResult `json:"elements"`
	} `json:"_embedded"`
	Total int `json:"total"`
	Count int `json:"count"`
}

// SearchOptions contains options for searching.
type SearchOptions struct {
	Query    string
	Type     string // "project", "work_package", "user", or "" for all
	Limit    int
	Offset   int
	PageSize int
}

// Search performs a search across OpenProject resources.
func (c *Client) Search(ctx context.Context, query, resourceType string, limit int) (*SearchResults, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if resourceType == "" {
		resourceType = "all"
	}

	path := fmt.Sprintf("/search?query=%s&limit=%d&type=%s", query, limit, resourceType)
	var results SearchResults
	if err := c.Get(ctx, path, &results); err != nil {
		return nil, err
	}
	return &results, nil
}
