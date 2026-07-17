package openproject

import (
	"context"

	external "github.com/pinealctx/openproject"
)

type DocumentCollection struct {
	Embedded struct {
		Elements []external.DocumentModel `json:"elements"`
	} `json:"_embedded"`
	Count int `json:"count"`
	Total int `json:"total"`
}

type DocumentListInput struct {
	Offset   int
	PageSize int
	SortBy   string
}

type DocumentUpdateInput struct {
	ID          int
	Title       string
	Description string
}

func (c *Client) ListDocuments(ctx context.Context, input DocumentListInput) (*DocumentCollection, error) {
	params := &external.ListDocumentsParams{}
	if input.Offset > 0 {
		params.Offset = ptr(input.Offset)
	}
	if input.PageSize > 0 {
		params.PageSize = ptr(input.PageSize)
	}
	if input.SortBy != "" {
		params.SortBy = ptr(input.SortBy)
	}

	var documents DocumentCollection
	resp, err := c.apiClient.ListDocuments(ctx, params)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &documents, DecodeResponse(resp, err, &documents)
}

func (c *Client) GetDocument(ctx context.Context, id int) (*external.DocumentModel, error) {
	var document external.DocumentModel
	resp, err := c.apiClient.ViewDocument(ctx, id)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &document, DecodeResponse(resp, err, &document)
}

func (c *Client) UpdateDocument(ctx context.Context, input DocumentUpdateInput) (*external.DocumentModel, error) {
	body := external.UpdateDocumentJSONRequestBody{}
	if input.Title != "" {
		body.Title = ptr(input.Title)
	}
	if input.Description != "" {
		body.Description = &struct {
			Raw *string `json:"raw,omitempty"`
		}{
			Raw: ptr(input.Description),
		}
	}

	var document external.DocumentModel
	resp, err := c.apiClient.UpdateDocument(ctx, input.ID, body)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	return &document, DecodeResponse(resp, err, &document)
}
