// client_test.go
package ipapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
	if client.HTTPClient.Timeout != defaultTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultTimeout, client.HTTPClient.Timeout)
	}
	if client.APIKey != "" {
		t.Errorf("expected empty API key, got %q", client.APIKey)
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

func TestNewClient_CheckRedirect_TooManyRedirects(t *testing.T) {
	// Create a server that redirects forever
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect", http.StatusFound)
	}))
	defer ts.Close()

	client := NewClient()
	_, err := client.HTTPClient.Get(ts.URL)
	if err == nil {
		t.Error("expected error for too many redirects")
	}
}

func TestNewClient_CheckRedirect_AllowedRedirects(t *testing.T) {
	redirectCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/final" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("done"))
			return
		}
		redirectCount++
		if redirectCount < 3 { // Within maxRedirects (3)
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient()
	resp, err := client.HTTPClient.Get(ts.URL)
	if err != nil {
		t.Errorf("expected successful redirect, got error: %v", err)
	}
	resp.Body.Close()
}

func TestNewClient_MultipleOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 15 * time.Second}
	client := NewClient(
		WithAPIKey("my-key"),
		WithCustomHTTPClient(customHTTP),
	)
	if client.APIKey != "my-key" {
		t.Errorf("expected API key 'my-key', got %q", client.APIKey)
	}
	if client.HTTPClient != customHTTP {
		t.Error("custom HTTP client not set")
	}
}

func TestWithBaseURL(t *testing.T) {
	c := NewClient(WithBaseURL("https://proxy.example.com/"))
	if c.BaseURL != "https://proxy.example.com/" {
		t.Errorf("expected overridden base URL, got %q", c.BaseURL)
	}
}

func TestWithBaseURL_EmptyIgnored(t *testing.T) {
	c := NewClient(WithBaseURL(""))
	if c.BaseURL != defaultBaseURL {
		t.Errorf("empty url should keep default %q, got %q", defaultBaseURL, c.BaseURL)
	}
}

func TestWithUserAgent(t *testing.T) {
	c := NewClient(WithUserAgent("my-app/2.0"))
	if c.UserAgent != "my-app/2.0" {
		t.Errorf("expected overridden UA, got %q", c.UserAgent)
	}
}

func TestWithUserAgent_EmptyIgnored(t *testing.T) {
	c := NewClient(WithUserAgent(""))
	if c.UserAgent != "ipapi-go-client/1.0" {
		t.Errorf("empty ua should keep default, got %q", c.UserAgent)
	}
}

func TestWithRetries(t *testing.T) {
	c := NewClient(WithRetries(5))
	if c.Retries != 5 {
		t.Errorf("expected 5 retries, got %d", c.Retries)
	}
}

func TestWithRetries_NegativeClampedToZero(t *testing.T) {
	c := NewClient(WithRetries(-3))
	if c.Retries != 0 {
		t.Errorf("negative should clamp to 0, got %d", c.Retries)
	}
}

func TestWithRetries_ZeroAllowed(t *testing.T) {
	c := NewClient(WithRetries(0))
	if c.Retries != 0 {
		t.Errorf("0 should be honored (no retries), got %d", c.Retries)
	}
}

func TestWithTimeout(t *testing.T) {
	c := NewClient(WithTimeout(7 * time.Second))
	if c.HTTPClient.Timeout != 7*time.Second {
		t.Errorf("expected 7s timeout, got %v", c.HTTPClient.Timeout)
	}
}

func TestWithTimeout_PreservesDefaultClientRedirectPolicy(t *testing.T) {
	// WithTimeout mutates the default *http.Client in place; CheckRedirect
	// configured by NewClient must still be present.
	c := NewClient(WithTimeout(2 * time.Second))
	if c.HTTPClient.Timeout != 2*time.Second {
		t.Errorf("expected 2s, got %v", c.HTTPClient.Timeout)
	}
	if c.HTTPClient.CheckRedirect == nil {
		t.Error("CheckRedirect must be preserved after WithTimeout")
	}
}

func TestWithTimeout_NonPositiveIgnored(t *testing.T) {
	for _, d := range []time.Duration{0, -1 * time.Second} {
		c := NewClient(WithTimeout(d))
		if c.HTTPClient.Timeout != defaultTimeout {
			t.Errorf("non-positive %v should keep default %v, got %v", d, defaultTimeout, c.HTTPClient.Timeout)
		}
	}
}

func TestWithTimeout_AppliedAfterCustomHTTPClient(t *testing.T) {
	// Recommended order: WithCustomHTTPClient first, then WithTimeout lands
	// on the custom client.
	custom := &http.Client{}
	c := NewClient(WithCustomHTTPClient(custom), WithTimeout(9*time.Second))
	if c.HTTPClient != custom {
		t.Error("custom HTTP client should not be replaced by WithTimeout")
	}
	if c.HTTPClient.Timeout != 9*time.Second {
		t.Errorf("timeout should land on custom client, got %v", c.HTTPClient.Timeout)
	}
}

func TestWithRateLimiter(t *testing.T) {
	ch := make(chan time.Time, 1)
	c := NewClient(WithRateLimiter(ch))
	if c.RateLimiter == nil {
		t.Error("rate limiter channel not set")
	}
	// Drain sanity: a buffered tick channel should be receivable without blocking.
	select {
	case <-c.RateLimiter:
	default:
		// channel empty by default; that's fine, the pointer is what matters
	}
}

func TestWithRateLimiter_NilAllowed(t *testing.T) {
	// First install a limiter, then explicitly disable it with nil.
	c := NewClient(WithRateLimiter(make(chan time.Time)), WithRateLimiter(nil))
	if c.RateLimiter != nil {
		t.Error("nil should be honored to disable rate limiting")
	}
}
