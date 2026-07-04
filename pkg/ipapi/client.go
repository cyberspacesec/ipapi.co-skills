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

// Format represents the response format for API requests.
type Format string

const (
	FormatJSON  Format = "json"
	FormatJSONP Format = "jsonp"
	FormatXML   Format = "xml"
	FormatCSV   Format = "csv"
	FormatYAML  Format = "yaml"
)

// validFormats contains all supported response formats.
var validFormats = map[Format]struct{}{
	FormatJSON:  {},
	FormatJSONP: {},
	FormatXML:   {},
	FormatCSV:   {},
	FormatYAML:  {},
}

// APIKeyMode controls how the API key is sent to the server.
type APIKeyMode int

const (
	// APIKeyHeader sends the API key as a Bearer Authorization header (default).
	APIKeyHeader APIKeyMode = iota
	// APIKeyQuery sends the API key as a ?key= query parameter.
	APIKeyQuery
)

var (
	ErrInvalidIP       = errors.New("invalid IP address")
	ErrInvalidField    = errors.New("invalid field name")
	ErrInvalidFormat   = errors.New("invalid response format")
	ErrRateLimited     = errors.New("API rate limit exceeded")
	ErrReservedIP      = errors.New("reserved IP address")
	ErrNotFound        = errors.New("resource not found")
	ErrServerError     = errors.New("server error")
	ErrUnexpectedData  = errors.New("unexpected response data")
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrInvalidKey      = errors.New("invalid API key")
)

type Client struct {
	HTTPClient   *http.Client
	BaseURL      string
	APIKey       string
	APIKeyMode   APIKeyMode
	UserAgent    string
	Retries      int
	RateLimiter  <-chan time.Time
	Callback     string // JSONP callback function name
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

// WithAPIKeyQuery sets the API key to be sent as a ?key= query parameter
// instead of the default Bearer Authorization header.
func WithAPIKeyQuery() ClientOption {
	return func(c *Client) {
		c.APIKeyMode = APIKeyQuery
	}
}

// WithCallback sets the JSONP callback function name for JSONP requests.
// This is only effective when using FormatJSONP.
func WithCallback(callback string) ClientOption {
	return func(c *Client) {
		c.Callback = callback
	}
}

// WithBaseURL overrides the default API base URL ("https://ipapi.co/").
// Useful for pointing the client at a proxy, a self-hosted mirror, or a
// test double. An empty url is ignored so the default is preserved.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		if url != "" {
			c.BaseURL = url
		}
	}
}

// WithUserAgent overrides the default User-Agent header
// ("ipapi-go-client/1.0"). An empty ua is ignored so the default is
// preserved.
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		if ua != "" {
			c.UserAgent = ua
		}
	}
}

// WithRetries sets the number of retries on transient failures
// (network errors and HTTP 5xx). The default is 2, meaning a request is
// attempted at most Retries+1 times. 4xx responses (including 429) are
// never retried. A negative n is treated as 0.
func WithRetries(n int) ClientOption {
	return func(c *Client) {
		if n < 0 {
			n = 0
		}
		c.Retries = n
	}
}

// WithTimeout sets the per-request timeout on the underlying HTTP client
// (default 10s). It mutates the existing *http.Client in place rather than
// replacing it, so any CheckRedirect / Transport configuration is preserved.
//
// If WithCustomHTTPClient is also used, apply it BEFORE WithTimeout so the
// timeout lands on the custom client; otherwise WithTimeout acts on the
// default client and is then superseded when the custom client is installed.
// A non-positive d is ignored.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		if d > 0 && c.HTTPClient != nil {
			c.HTTPClient.Timeout = d
		}
	}
}

// WithRateLimiter installs a client-side throttle: before every request,
// doRequest blocks on a receive from ch. Pass a channel created with
// time.Tick to cap the global request rate, or nil to explicitly disable
// any previously-configured limiter.
func WithRateLimiter(ch <-chan time.Time) ClientOption {
	return func(c *Client) {
		c.RateLimiter = ch
	}
}
