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
	"time"
)

var validFields = map[string]struct{}{
	"ip": {}, "city": {}, "region": {}, "region_code": {}, "country": {},
	"country_name": {}, "continent_code": {}, "in_eu": {}, "postal": {},
	"latitude": {}, "longitude": {}, "latlong": {}, "timezone": {},
	"utc_offset": {}, "languages": {}, "country_calling_code": {},
	"currency": {}, "asn": {}, "org": {}, "hostname": {},
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

	if err != nil {
		return nil, err
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

	u, _ := url.Parse(c.BaseURL)
	u.Path = path.Join(u.Path, ip, format) + "/"

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, c.handleError(err)
	}

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

func (c *Client) GetField(ctx context.Context, ip, field string) (string, error) {
	if _, valid := validFields[field]; !valid {
		return "", c.handleError(ErrInvalidField)
	}

	u, _ := url.Parse(c.BaseURL)
	u.Path = path.Join(u.Path, ip, field) + "/"

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return "", c.handleError(err)
	}

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
	u, _ := url.Parse(c.BaseURL)
	u.Path = path.Join(u.Path, format) + "/"

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, c.handleError(err)
	}

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

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
}

func (c *Client) mapStatusCodeToError(code int) error {
	switch code {
	case http.StatusBadRequest:
		return fmt.Errorf("%w: invalid request", ErrServerError)
	case http.StatusForbidden:
		return fmt.Errorf("%w: check API key", ErrServerError)
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusInternalServerError:
		return ErrServerError
	default:
		return fmt.Errorf("unexpected status code: %d", code)
	}
}
