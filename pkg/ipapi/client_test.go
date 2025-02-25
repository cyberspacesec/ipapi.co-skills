// client_test.go
package ipapi

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient()
	if client.BaseURL != defaultBaseURL {
		t.Errorf("expected default base URL %q, got %q", defaultBaseURL, client.BaseURL)
	}
	if client.UserAgent != "ipapi-go-client/1.0" {
		t.Errorf("unexpected User-Agent: %q", client.UserAgent)
	}
	if client.Retries != 2 {
		t.Errorf("expected 2 retries, got %d", client.Retries)
	}
}

func TestWithCustomHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 5 * time.Second}
	client := NewClient(WithCustomHTTPClient(customClient))
	if client.HTTPClient != customClient {
		t.Error("custom HTTP client not set")
	}
}

func TestWithAPIKey(t *testing.T) {
	client := NewClient(WithAPIKey("test-key"))
	if client.APIKey != "test-key" {
		t.Error("API key not set")
	}
}

func TestErrorHandler(t *testing.T) {
	handlerCalled := false
	customHandler := func(err error) error {
		handlerCalled = true
		return err
	}

	client := NewClient(WithErrorHandler(customHandler))
	err := client.handleError(ErrInvalidIP)

	if !handlerCalled {
		t.Error("custom error handler not called")
	}
	if !errors.Is(err, ErrInvalidIP) {
		t.Error("unexpected error returned")
	}
}
