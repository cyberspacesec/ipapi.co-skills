// api.go
package ipapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"time"
)

var validFields = map[string]struct{}{
	"ip": {}, "network": {}, "version": {}, "city": {}, "region": {},
	"region_code": {}, "country": {}, "country_name": {}, "country_code": {},
	"country_code_iso3": {}, "country_capital": {}, "country_tld": {},
	"continent_code": {}, "in_eu": {}, "postal": {},
	"latitude": {}, "longitude": {}, "latlong": {}, "timezone": {},
	"utc_offset": {}, "languages": {}, "country_calling_code": {},
	"currency": {}, "currency_name": {}, "country_area": {},
	"country_population": {}, "asn": {}, "org": {}, "hostname": {},
}

// ValidateFormat checks whether the given format string is a valid API response format.
func ValidateFormat(format string) error {
	if _, ok := validFormats[Format(format)]; !ok {
		return ErrInvalidFormat
	}
	return nil
}

// ValidFields returns the list of all field names accepted by GetField and
// GetClientField. The order is stable (sorted), which makes it suitable for
// CLI enumeration and machine consumption.
func ValidFields() []string {
	out := make([]string, 0, len(validFields))
	for f := range validFields {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// IsValidField reports whether the given field name is a valid API field.
func IsValidField(field string) bool {
	_, ok := validFields[field]
	return ok
}

// ValidFormats returns the list of all response formats accepted by the API.
func ValidFormats() []Format {
	out := make([]Format, 0, len(validFormats))
	for f := range validFormats {
		out = append(out, f)
	}
	// keep a stable, conventional order
	order := []Format{FormatJSON, FormatJSONP, FormatXML, FormatCSV, FormatYAML}
	seen := map[Format]bool{}
	result := make([]Format, 0, len(validFormats))
	for _, f := range order {
		if _, ok := validFormats[f]; ok && !seen[f] {
			result = append(result, f)
			seen[f] = true
		}
	}
	for _, f := range out {
		if !seen[f] {
			result = append(result, f)
			seen[f] = true
		}
	}
	return result
}

// newGetRequest creates an HTTP GET request for the given base URL and path segments.
// It combines URL construction and request creation, returning an error if either fails.
func newGetRequest(ctx context.Context, baseURL string, segments ...string) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, path.Join(segments...)) + "/"
	return http.NewRequestWithContext(ctx, "GET", u.String(), nil)
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.RateLimiter != nil {
		<-c.RateLimiter
	}

	var resp *http.Response
	var err error

	for i := 0; i <= c.Retries; i++ {
		resp, err = c.HTTPClient.Do(req)
		if err == nil && resp.StatusCode < 500 { // 仅网络错误和5xx错误重试
			break
		}
		if err == nil { // 处理5xx错误
			resp.Body.Close()
		}

		if i == c.Retries {
			if err != nil {
				return nil, fmt.Errorf("request failed after %d retries: %w", c.Retries, err)
			}
			return nil, fmt.Errorf("server error after %d retries (status: %d)", c.Retries, resp.StatusCode)
		}

		time.Sleep(defaultRetryDelay)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var apiErr *APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.HasError {
			return nil, apiErr // 直接返回APIError实例
		}
		return nil, c.mapStatusCodeToError(resp.StatusCode)
	}

	return resp, nil
}

func (c *Client) GetIPInfo(ctx context.Context, ip string, format string) (*IPInfo, error) {
	if err := ValidateIP(ip); err != nil {
		return nil, c.handleError(err)
	}

	if err := ValidateFormat(format); err != nil {
		return nil, c.handleError(err)
	}

	req, err := newGetRequest(ctx, c.BaseURL, ip, format)
	if err != nil {
		return nil, c.handleError(err)
	}

	c.applyAuth(req)
	c.setHeaders(req)
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, c.handleError(err)
	}
	defer resp.Body.Close()

	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, c.handleError(fmt.Errorf("%w: %v", ErrUnexpectedData, err))
	}

	info.RetrievedAt = time.Now().UTC()
	return &info, nil
}

// GetIPInfoRaw returns the raw response body for a specific IP lookup.
// This is useful for non-JSON formats (xml, csv, yaml, jsonp) where you need
// the raw bytes rather than parsed IPInfo. The format must be one of the
// valid formats (json, jsonp, xml, csv, yaml).
func (c *Client) GetIPInfoRaw(ctx context.Context, ip string, format string) ([]byte, error) {
	if err := ValidateIP(ip); err != nil {
		return nil, c.handleError(err)
	}

	if err := ValidateFormat(format); err != nil {
		return nil, c.handleError(err)
	}

	req, err := newGetRequest(ctx, c.BaseURL, ip, format)
	if err != nil {
		return nil, c.handleError(err)
	}

	c.applyAuth(req)
	c.setHeaders(req)
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, c.handleError(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, c.handleError(err)
	}

	return data, nil
}

func (c *Client) GetField(ctx context.Context, ip, field string) (string, error) {
	if _, valid := validFields[field]; !valid {
		return "", c.handleError(ErrInvalidField)
	}

	req, err := newGetRequest(ctx, c.BaseURL, ip, field)
	if err != nil {
		return "", c.handleError(err)
	}

	c.applyAuth(req)
	c.setHeaders(req)
	resp, err := c.doRequest(req)
	if err != nil {
		return "", c.handleError(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", c.handleError(err)
	}

	return string(data), nil
}

func (c *Client) GetClientIPInfo(ctx context.Context, format string) (*IPInfo, error) {
	if err := ValidateFormat(format); err != nil {
		return nil, c.handleError(err)
	}

	req, err := newGetRequest(ctx, c.BaseURL, format)
	if err != nil {
		return nil, c.handleError(err)
	}

	c.applyAuth(req)
	c.setHeaders(req)
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, c.handleError(err)
	}
	defer resp.Body.Close()

	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, c.handleError(fmt.Errorf("%w: %v", ErrUnexpectedData, err))
	}

	info.RetrievedAt = time.Now().UTC()
	return &info, nil
}

// GetClientIPInfoRaw returns the raw response body for the client's IP lookup.
// This is useful for non-JSON formats (xml, csv, yaml, jsonp) where you need
// the raw bytes rather than parsed IPInfo. The format must be one of the
// valid formats (json, jsonp, xml, csv, yaml).
func (c *Client) GetClientIPInfoRaw(ctx context.Context, format string) ([]byte, error) {
	if err := ValidateFormat(format); err != nil {
		return nil, c.handleError(err)
	}

	req, err := newGetRequest(ctx, c.BaseURL, format)
	if err != nil {
		return nil, c.handleError(err)
	}

	c.applyAuth(req)
	c.setHeaders(req)
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, c.handleError(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, c.handleError(err)
	}

	return data, nil
}

// GetClientField returns a single location field for the client's IP address.
// The field must be one of the valid fields (e.g., "ip", "city", "country", "asn").
// This corresponds to the API endpoint: GET https://ipapi.co/{field}/
func (c *Client) GetClientField(ctx context.Context, field string) (string, error) {
	if _, valid := validFields[field]; !valid {
		return "", c.handleError(ErrInvalidField)
	}

	req, err := newGetRequest(ctx, c.BaseURL, field)
	if err != nil {
		return "", c.handleError(err)
	}

	c.applyAuth(req)
	c.setHeaders(req)
	resp, err := c.doRequest(req)
	if err != nil {
		return "", c.handleError(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", c.handleError(err)
	}

	return string(data), nil
}

// setHeaders sets common HTTP headers on the request.
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.UserAgent)
}

// applyAuth applies the API key authentication and JSONP callback to the request.
// If APIKeyMode is APIKeyQuery, the key is added as a ?key= query parameter.
// Otherwise, the key is set as a Bearer Authorization header.
// If a JSONP callback is set, it is added as a ?callback= query parameter.
func (c *Client) applyAuth(req *http.Request) {
	if c.APIKey != "" {
		switch c.APIKeyMode {
		case APIKeyQuery:
			q := req.URL.Query()
			q.Set("key", c.APIKey)
			req.URL.RawQuery = q.Encode()
		default:
			req.Header.Set("Authorization", "Bearer "+c.APIKey)
		}
	}

	// Apply JSONP callback if specified
	if c.Callback != "" {
		q := req.URL.Query()
		q.Set("callback", c.Callback)
		req.URL.RawQuery = q.Encode()
	}
}

func (c *Client) mapStatusCodeToError(code int) error {
	switch code {
	case http.StatusBadRequest:
		return fmt.Errorf("%w: invalid request", ErrServerError)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", ErrInvalidKey, "check API key")
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusMethodNotAllowed:
		return ErrMethodNotAllowed
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusInternalServerError:
		return ErrServerError
	default:
		return fmt.Errorf("unexpected status code: %d", code)
	}
}
