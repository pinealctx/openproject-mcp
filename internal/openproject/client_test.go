package openproject

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConnectionUsesCurrentUserEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v3/users/me" {
			t.Errorf("TestConnection requested %q, want /api/v3/users/me", request.URL.Path)
			http.NotFound(response, request)
			return
		}
		username, password, ok := request.BasicAuth()
		if !ok || username != "apikey" || password != "token" {
			t.Errorf("TestConnection sent unexpected credentials")
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"_type":"User","id":7,"login":"team-member"}`)
	}))
	defer server.Close()

	client := NewClientDirect(server.URL, "token", time.Second)
	user, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("TestConnection returned error: %v", err)
	}
	if user.Id != 7 || user.Login == nil || *user.Login != "team-member" {
		t.Fatalf("TestConnection returned unexpected user: %#v", user)
	}
}

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
