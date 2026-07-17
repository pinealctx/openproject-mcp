package openproject

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadResponseReturnsEnglishAPIError(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusForbidden,
		Body: io.NopCloser(strings.NewReader(
			`{"_type":"Error","errorIdentifier":"urn:openproject-org:api:v3:errors:Unauthenticated","message":"您无权访问此资源。"}`,
		)),
	}

	err := ReadResponse(response, nil)
	if err == nil {
		t.Fatal("ReadResponse returned no error")
	}
	const expected = "OpenProject API request failed with HTTP 403 Forbidden (error identifier: urn:openproject-org:api:v3:errors:Unauthenticated)"
	if err.Error() != expected {
		t.Fatalf("ReadResponse error = %q, want %q", err, expected)
	}
	if strings.Contains(err.Error(), "您无权") {
		t.Fatalf("ReadResponse exposed a localized server message: %q", err)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("ReadResponse error type = %T, want *APIError", err)
	}
	if apiErr.Message != "您无权访问此资源。" {
		t.Fatalf("APIError.Message = %q", apiErr.Message)
	}
}

func TestAPIErrorOmitsUnsafeIdentifier(t *testing.T) {
	err := (&APIError{
		StatusCode: http.StatusUnprocessableEntity,
		ErrorID:    "错误标识",
	}).Error()

	if err != "OpenProject API request failed with HTTP 422 Unprocessable Entity" {
		t.Fatalf("APIError.Error() = %q", err)
	}
}
