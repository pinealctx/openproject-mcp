package openproject

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	external "github.com/pinealctx/openproject"
)

const (
	SearchTypeProject     = "project"
	SearchTypeWorkPackage = "work_package"
	SearchTypeUser        = "user"

	DefaultSearchLimit = 10
	MaxSearchLimit     = 100
)

var searchTypes = []string{SearchTypeWorkPackage, SearchTypeProject, SearchTypeUser}

// SearchInput describes a search across one or all supported resource types.
type SearchInput struct {
	Query string
	Type  string
	Limit int
}

// SearchResults is a concise, stable representation of OpenProject search results.
type SearchResults struct {
	Query        string          `json:"query"`
	ResourceType string          `json:"resourceType,omitempty"`
	Limit        int             `json:"limit"`
	Count        int             `json:"count"`
	Total        int             `json:"total"`
	Results      []SearchResult  `json:"results"`
	Warnings     []SearchWarning `json:"warnings,omitempty"`
}

// SearchResult identifies one matching OpenProject resource.
type SearchResult struct {
	Type       string `json:"type"`
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Identifier string `json:"identifier,omitempty"`
	Status     string `json:"status,omitempty"`
}

// SearchWarning reports a resource type that could not be searched in an aggregate request.
type SearchWarning struct {
	ResourceType string `json:"resourceType"`
	Message      string `json:"message"`
}

// Search queries OpenProject collection endpoints with their supported filters.
// Aggregate searches are best-effort so one inaccessible resource type does not
// hide successful results from the others.
func (c *Client) Search(ctx context.Context, input SearchInput) (*SearchResults, error) {
	query, resourceType, limit, err := normalizeSearchInput(input)
	if err != nil {
		return nil, err
	}

	result := &SearchResults{
		Query:        query,
		ResourceType: resourceType,
		Limit:        limit,
		Results:      make([]SearchResult, 0),
	}
	types := searchTypes
	if resourceType != "" {
		types = []string{resourceType}
	}

	var searchErrors []error
	for _, currentType := range types {
		items, total, searchErr := c.searchResource(ctx, currentType, query, limit)
		if searchErr != nil {
			wrapped := fmt.Errorf("search %s: %w", currentType, searchErr)
			if resourceType != "" {
				return nil, wrapped
			}
			searchErrors = append(searchErrors, wrapped)
			result.Warnings = append(result.Warnings, SearchWarning{
				ResourceType: currentType,
				Message:      wrapped.Error(),
			})
			continue
		}
		result.Results = append(result.Results, items...)
		result.Total += total
	}

	if len(searchErrors) == len(types) {
		return nil, fmt.Errorf("search failed for every resource type: %w", errors.Join(searchErrors...))
	}
	result.Count = len(result.Results)
	return result, nil
}

func (c *Client) searchResource(ctx context.Context, resourceType, query string, limit int) ([]SearchResult, int, error) {
	switch resourceType {
	case SearchTypeWorkPackage:
		filter, err := buildSearchFilter("search", "**", query)
		if err != nil {
			return nil, 0, err
		}
		list, err := c.ListWorkPackages(ctx, WorkPackageListInput{PageSize: limit, Filters: filter})
		if err != nil {
			return nil, 0, err
		}
		items := make([]SearchResult, 0, min(limit, len(list.UnderscoreEmbedded.Elements)))
		for _, workPackage := range list.UnderscoreEmbedded.Elements {
			if len(items) == limit {
				break
			}
			status := ""
			if workPackage.UnderscoreLinks.Status.Title != nil {
				status = *workPackage.UnderscoreLinks.Status.Title
			}
			items = append(items, SearchResult{
				Type:   SearchTypeWorkPackage,
				ID:     valueOrZero(workPackage.Id),
				Title:  workPackage.Subject,
				Status: status,
			})
		}
		return items, list.Total, nil

	case SearchTypeProject:
		filter, err := buildSearchFilter("name_and_identifier", "~", query)
		if err != nil {
			return nil, 0, err
		}
		list, err := c.ListProjects(ctx, ProjectListInput{Filters: filter})
		if err != nil {
			return nil, 0, err
		}
		items := make([]SearchResult, 0, min(limit, len(list.UnderscoreEmbedded.Elements)))
		for _, project := range list.UnderscoreEmbedded.Elements {
			if len(items) == limit {
				break
			}
			items = append(items, SearchResult{
				Type:       SearchTypeProject,
				ID:         valueOrZero(project.Id),
				Title:      valueOrEmpty(project.Name),
				Identifier: valueOrEmpty(project.Identifier),
			})
		}
		return items, list.Total, nil

	case SearchTypeUser:
		filter, err := buildSearchFilter("name", "~", query)
		if err != nil {
			return nil, 0, err
		}
		params := &external.ListUsersParams{PageSize: ptr(limit), Filters: ptr(filter)}
		var list external.UserCollectionModel
		resp, requestErr := c.apiClient.ListUsers(ctx, params)
		if resp != nil {
			defer func() { _ = resp.Body.Close() }()
		}
		if decodeErr := DecodeResponse(resp, requestErr, &list); decodeErr != nil {
			return nil, 0, decodeErr
		}
		items := make([]SearchResult, 0, min(limit, len(list.UnderscoreEmbedded.Elements)))
		for _, user := range list.UnderscoreEmbedded.Elements {
			if len(items) == limit {
				break
			}
			items = append(items, SearchResult{Type: SearchTypeUser, ID: user.Id, Title: user.Name})
		}
		return items, list.Total, nil

	default:
		return nil, 0, fmt.Errorf("unsupported resource type %q", resourceType)
	}
}

func normalizeSearchInput(input SearchInput) (string, string, int, error) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return "", "", 0, fmt.Errorf("search query is required")
	}

	resourceType := strings.TrimSpace(input.Type)
	if resourceType != "" && resourceType != SearchTypeProject && resourceType != SearchTypeWorkPackage && resourceType != SearchTypeUser {
		return "", "", 0, fmt.Errorf("search type must be project, work_package, or user")
	}

	limit := input.Limit
	if limit == 0 {
		limit = DefaultSearchLimit
	}
	if limit < 1 || limit > MaxSearchLimit {
		return "", "", 0, fmt.Errorf("search limit must be between 1 and %d", MaxSearchLimit)
	}
	return query, resourceType, limit, nil
}

func buildSearchFilter(name, operator, query string) (string, error) {
	filter := []map[string]map[string]any{
		{name: {"operator": operator, "values": []string{query}}},
	}
	encoded, err := json.Marshal(filter)
	if err != nil {
		return "", fmt.Errorf("encode search filter: %w", err)
	}
	return string(encoded), nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
