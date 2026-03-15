package openproject

import (
	"context"
	"fmt"
)

// ListVersions retrieves all versions for a project.
func (c *Client) ListVersions(ctx context.Context, projectID int) (*VersionList, error) {
	var result VersionList
	if err := c.Get(ctx, fmt.Sprintf("/projects/%d/versions", projectID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateVersionOptions contains options for creating a version.
type CreateVersionOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	EndDate     string `json:"endDate,omitempty"`
	ProjectID   int    `json:"-"`
}

// CreateVersion creates a new version.
func (c *Client) CreateVersion(ctx context.Context, opts *CreateVersionOptions) (*Version, error) {
	payload := map[string]interface{}{
		"name":        opts.Name,
		"description": opts.Description,
		"startDate":   opts.StartDate,
		"endDate":     opts.EndDate,
		"_links": map[string]interface{}{
			"project": map[string]string{
				"href": fmt.Sprintf("/api/v3/projects/%d", opts.ProjectID),
			},
		},
	}

	var version Version
	if err := c.Post(ctx, "/versions", payload, &version); err != nil {
		return nil, err
	}
	return &version, nil
}
