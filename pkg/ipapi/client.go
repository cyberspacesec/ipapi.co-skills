// client.go
package ipapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultBaseURL    = "https://ipapi.co/"
	defaultTimeout    = 10 * time.Second
	maxRedirects      = 3
	defaultRetryDelay = 500 * time.Millisecond
)

var (
	ErrInvalidIP      = errors.New("invalid IP address")
	ErrInvalidField   = errors.New("invalid field name")
	ErrRateLimited    = errors.New("API rate limit exceeded")
	ErrReservedIP     = errors.New("reserved IP address")
	ErrNotFound       = errors.New("resource not found")
	ErrServerError    = errors.New("server error")
	ErrUnexpectedData = errors.New("unexpected response data")
)

type Client struct {
	HTTPClient   *http.Client
	BaseURL      string
	APIKey       string
	UserAgent    string
	Retries      int
	RateLimiter  <-chan time.Time
	errorHandler func(error) error
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				return nil
			},
		},
		BaseURL:   defaultBaseURL,
		UserAgent: "ipapi-go-client/1.0",
		Retries:   2,
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.APIKey = key
	}
}

func WithCustomHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = client
	}
}

func WithErrorHandler(handler func(error) error) ClientOption {
	return func(c *Client) {
		c.errorHandler = handler
	}
}
