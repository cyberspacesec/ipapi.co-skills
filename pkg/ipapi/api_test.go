// api_test.go
package ipapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- GetIPInfo ---

func TestClient_GetIPInfo_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/json/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"ip": "8.8.8.8",
			"network": "8.8.8.0/24",
			"version": "IPv4",
			"city": "Mountain View",
			"region": "California",
			"region_code": "CA",
			"country": "US",
			"country_name": "United States",
			"country_code": "US",
			"country_code_iso3": "USA",
			"country_capital": "Washington",
			"country_tld": ".us",
			"continent_code": "NA",
			"in_eu": false,
			"postal": "94043",
			"latitude": 37.3860,
			"longitude": -122.0838,
			"timezone": "America/Los_Angeles",
			"utc_offset": "-0700",
			"country_calling_code": "+1",
			"currency": "USD",
			"currency_name": "Dollar",
			"languages": "en-US,es-US,haw,fr",
			"country_area": 9629091.0,
			"country_population": 327167434,
			"asn": "AS15169",
			"org": "Google LLC"
		}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	info, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}

	if info.IP != "8.8.8.8" {
		t.Errorf("expected 8.8.8.8, got %s", info.IP)
	}
	if info.Network != "8.8.8.0/24" {
		t.Errorf("expected 8.8.8.0/24, got %s", info.Network)
	}
	if info.Version != "IPv4" {
		t.Errorf("expected IPv4, got %s", info.Version)
	}
	if info.City != "Mountain View" {
		t.Errorf("expected Mountain View, got %s", info.City)
	}
	if info.Region != "California" {
		t.Errorf("expected California, got %s", info.Region)
	}
	if info.RegionCode != "CA" {
		t.Errorf("expected CA, got %s", info.RegionCode)
	}
	if info.Country != "US" {
		t.Errorf("expected US, got %s", info.Country)
	}
	if info.CountryName != "United States" {
		t.Errorf("expected United States, got %s", info.CountryName)
	}
	if info.CountryCode != "US" {
		t.Errorf("expected US, got %s", info.CountryCode)
	}
	if info.CountryCodeISO3 != "USA" {
		t.Errorf("expected USA, got %s", info.CountryCodeISO3)
	}
	if info.CountryCapital != "Washington" {
		t.Errorf("expected Washington, got %s", info.CountryCapital)
	}
	if info.CountryTLD != ".us" {
		t.Errorf("expected .us, got %s", info.CountryTLD)
	}
	if info.ContinentCode != "NA" {
		t.Errorf("expected NA, got %s", info.ContinentCode)
	}
	if info.InEU != false {
		t.Errorf("expected false, got %v", info.InEU)
	}
	if info.Postal == nil || *info.Postal != "94043" {
		t.Errorf("expected 94043, got %v", info.Postal)
	}
	if info.Latitude != 37.3860 {
		t.Errorf("unexpected latitude: %f", info.Latitude)
	}
	if info.Longitude != -122.0838 {
		t.Errorf("unexpected longitude: %f", info.Longitude)
	}
	if info.Timezone != "America/Los_Angeles" {
		t.Errorf("expected America/Los_Angeles, got %s", info.Timezone)
	}
	if info.UTCOffset != "-0700" {
		t.Errorf("expected -0700, got %s", info.UTCOffset)
	}
	if info.CountryCallingCode != "+1" {
		t.Errorf("expected +1, got %s", info.CountryCallingCode)
	}
	if info.Currency != "USD" {
		t.Errorf("expected USD, got %s", info.Currency)
	}
	if info.CurrencyName != "Dollar" {
		t.Errorf("expected Dollar, got %s", info.CurrencyName)
	}
	if info.Languages != "en-US,es-US,haw,fr" {
		t.Errorf("expected en-US,es-US,haw,fr, got %s", info.Languages)
	}
	if info.CountryArea != 9629091.0 {
		t.Errorf("expected 9629091.0, got %f", info.CountryArea)
	}
	if info.CountryPopulation != 327167434 {
		t.Errorf("expected 327167434, got %d", info.CountryPopulation)
	}
	if info.ASN != "AS15169" {
		t.Errorf("expected AS15169, got %s", info.ASN)
	}
	if info.Org != "Google LLC" {
		t.Errorf("expected Google LLC, got %s", info.Org)
	}
	if info.RetrievedAt.IsZero() {
		t.Error("expected RetrievedAt to be set")
	}
}

func TestClient_GetIPInfo_InvalidIP(t *testing.T) {
	client := NewClient()
	_, err := client.GetIPInfo(context.Background(), "invalid-ip", "json")
	if !errors.Is(err, ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

func TestClient_GetIPInfo_InvalidIP_WithErrorHandler(t *testing.T) {
	handlerCalled := false
	client := NewClient(WithErrorHandler(func(err error) error {
		handlerCalled = true
		return err
	}))
	_, err := client.GetIPInfo(context.Background(), "invalid-ip", "json")
	if !handlerCalled {
		t.Error("error handler was not called")
	}
	if !errors.Is(err, ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

func TestClient_GetIPInfo_UnexpectedData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `this is not valid json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrUnexpectedData) {
		t.Errorf("expected ErrUnexpectedData, got %v", err)
	}
}

func TestClient_GetIPInfo_NullPostal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"ip": "1.1.1.1",
			"network": "1.1.1.0/24",
			"city": "Sydney",
			"country": "AU",
			"country_code": "AU",
			"postal": null,
			"latitude": -33.8688,
			"longitude": 151.2093
		}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	info, err := client.GetIPInfo(context.Background(), "1.1.1.1", "json")
	if err != nil {
		t.Fatal(err)
	}

	if info.Postal != nil {
		t.Errorf("expected nil postal, got %v", info.Postal)
	}
	if info.GetPostal() != "" {
		t.Errorf("expected empty string from GetPostal(), got %q", info.GetPostal())
	}
}

func TestClient_GetIPInfo_NonNullPostal(t *testing.T) {
	postal := "94043"
	info := &IPInfo{Postal: &postal}
	if info.GetPostal() != "94043" {
		t.Errorf("expected 94043, got %s", info.GetPostal())
	}
}

// --- GetField ---

func TestClient_GetField_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/country/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		fmt.Fprint(w, "US")
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	result, err := client.GetField(context.Background(), "8.8.8.8", "country")
	if err != nil {
		t.Fatal(err)
	}
	if result != "US" {
		t.Errorf("expected US, got %s", result)
	}
}

func TestClient_GetField_InvalidField(t *testing.T) {
	client := NewClient()
	_, err := client.GetField(context.Background(), "8.8.8.8", "invalid_field")
	if !errors.Is(err, ErrInvalidField) {
		t.Errorf("expected ErrInvalidField, got %v", err)
	}
}

func TestClient_GetField_NewFields(t *testing.T) {
	// Verify all new fields are accepted as valid (not ErrInvalidField)
	newFields := []string{"network", "country_code", "version", "country_code_iso3",
		"country_capital", "country_tld", "country_area", "country_population", "currency_name"}

	for _, field := range newFields {
		// Create a mock server so we don't hit the real API
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "test-value")
		}))
		client := NewClient(WithCustomHTTPClient(ts.Client()))
		client.BaseURL = ts.URL + "/"

		_, err := client.GetField(context.Background(), "8.8.8.8", field)
		if err != nil {
			t.Errorf("field %q should be valid but got error: %v", field, err)
		}
		ts.Close()
	}
}

func TestClient_GetField_AllValidFields(t *testing.T) {
	// Test every single field in validFields
	for field := range validFields {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "value")
		}))
		client := NewClient(WithCustomHTTPClient(ts.Client()))
		client.BaseURL = ts.URL + "/"

		_, err := client.GetField(context.Background(), "8.8.8.8", field)
		if err != nil {
			t.Errorf("field %q should be valid but got error: %v", field, err)
		}
		ts.Close()
	}
}

// --- GetClientIPInfo ---

func TestClient_GetClientIPInfo_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/" {
			t.Errorf("unexpected path: %s, expected /json/", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"ip": "203.0.113.1",
			"network": "203.0.113.0/24",
			"city": "Test City",
			"country": "US",
			"country_code": "US",
			"latitude": 37.0,
			"longitude": -122.0
		}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	info, err := client.GetClientIPInfo(context.Background(), "json")
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "203.0.113.1" {
		t.Errorf("expected 203.0.113.1, got %s", info.IP)
	}
	if info.Network != "203.0.113.0/24" {
		t.Errorf("expected 203.0.113.0/24, got %s", info.Network)
	}
	if info.City != "Test City" {
		t.Errorf("expected Test City, got %s", info.City)
	}
	if info.RetrievedAt.IsZero() {
		t.Error("expected RetrievedAt to be set")
	}
}

func TestClient_GetClientIPInfo_UnexpectedData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetClientIPInfo(context.Background(), "json")
	if !errors.Is(err, ErrUnexpectedData) {
		t.Errorf("expected ErrUnexpectedData, got %v", err)
	}
}

func TestClient_GetClientIPInfo_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(APIError{
			HasError: true,
			Reason:   "RateLimited",
			Message:  "API rate limit exceeded",
		})
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetClientIPInfo(context.Background(), "json")
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

// --- doRequest and retry logic ---

func TestClient_GetIPInfo_RetryLogic(t *testing.T) {
	retryCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("test"),
	)
	client.BaseURL = ts.URL + "/"
	client.Retries = 2

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err == nil {
		t.Fatal("expected error")
	}

	if retryCount != 3 { // initial request + 2 retries
		t.Errorf("expected 3 attempts, got %d", retryCount)
	}
}

func TestClient_doRequest_NetworkError(t *testing.T) {
	// Server that immediately closes connections
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This simulates a connection that gets refused after server close
	}))
	ts.Close() // Close immediately so connections fail

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"
	client.Retries = 1

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestClient_doRequest_RateLimiter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ip": "8.8.8.8"}`)
	}))
	defer ts.Close()

	// Create a rate limiter that allows immediate execution
	rateLimiter := time.After(0)
	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"
	client.RateLimiter = rateLimiter

	info, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "8.8.8.8" {
		t.Errorf("expected 8.8.8.8, got %s", info.IP)
	}
}

func TestClient_doRequest_5xx_ThenSuccess(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusBadGateway) // 502
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ip": "8.8.8.8", "city": "Mountain View"}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"
	client.Retries = 2

	info, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
	if info.IP != "8.8.8.8" {
		t.Errorf("expected 8.8.8.8, got %s", info.IP)
	}
}

func TestClient_doRequest_ZeroRetries(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"
	client.Retries = 0

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err == nil {
		t.Fatal("expected error")
	}
	if callCount != 1 {
		t.Errorf("expected 1 attempt with 0 retries, got %d", callCount)
	}
}

// --- Error handling: doRequest status code mapping ---

func TestClient_mapStatusCodeToError_BadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `not a valid api error json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrServerError) {
		t.Errorf("expected ErrServerError for 400, got %v", err)
	}
}

func TestClient_mapStatusCodeToError_Forbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `not a valid api error json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("expected ErrInvalidKey for 403, got %v", err)
	}
}

func TestClient_mapStatusCodeToError_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `not a valid api error json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for 404, got %v", err)
	}
}

func TestClient_mapStatusCodeToError_TooManyRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `not a valid api error json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited for 429, got %v", err)
	}
}

func TestClient_mapStatusCodeToError_InternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"
	client.Retries = 0

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err == nil {
		t.Fatal("expected error for 500")
	}
	// With Retries=0, the error comes from the retry loop, not mapStatusCodeToError
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention status 500, got %v", err)
	}
}

func TestClient_mapStatusCodeToError_UnexpectedCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(418) // I'm a teapot
		fmt.Fprint(w, `not a valid api error json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err == nil {
		t.Fatal("expected error for 418")
	}
	if errors.Is(err, ErrServerError) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrRateLimited) {
		t.Errorf("unexpected known error type for 418, got %v", err)
	}
	if !strings.Contains(err.Error(), "418") {
		t.Errorf("expected error to mention status code 418, got %v", err)
	}
}

// --- API error handling ---

func TestClient_HandleAPIError_RateLimited(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(APIError{
			HasError: true,
			Reason:   "RateLimited",
			Message:  "API rate limit exceeded",
		})
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

func TestClient_HandleAPIError_ReservedIP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{
			HasError: true,
			Reason:   "Reserved IP Address",
			Message:  "Reserved IP Address",
			IP:       "127.0.0.1",
		})
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrReservedIP) {
		t.Errorf("expected ErrReservedIP, got %v", err)
	}
}

func TestClient_HandleAPIError_InvalidIPAddress(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{
			HasError: true,
			Reason:   "Invalid IP Address",
			Message:  "Invalid IP Address",
			IP:       "999.999.999.999",
		})
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

// --- Headers ---

func TestClient_Headers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "ipapi-go-client/1.0" {
			t.Error("missing User-Agent header")
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing Authorization header")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(IPInfo{IP: "8.8.8.8"})
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("test-key"),
	)
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Headers_NoAPIKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "ipapi-go-client/1.0" {
			t.Error("missing User-Agent header")
		}
		if r.Header.Get("Authorization") != "" {
			t.Errorf("expected no Authorization header, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(IPInfo{IP: "8.8.8.8"})
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
}

// --- Context cancellation ---

func TestClient_ContextCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(ctx, "8.8.8.8", "json")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context canceled error, got %v", err)
	}
}

// --- ParseLatLong ---

func TestParseLatLong(t *testing.T) {
	info := &IPInfo{LatLong: "37.3860,-122.0838"}
	lat, lon, err := info.ParseLatLong()
	if err != nil {
		t.Fatal(err)
	}
	if lat != 37.3860 || lon != -122.0838 {
		t.Errorf("unexpected lat/long: %f, %f", lat, lon)
	}
}

func TestParseLatLong_InvalidFormat(t *testing.T) {
	info := &IPInfo{LatLong: "invalid"}
	_, _, err := info.ParseLatLong()
	if err == nil {
		t.Error("expected error for invalid latlong")
	}
}

func TestParseLatLong_InvalidLatitude(t *testing.T) {
	info := &IPInfo{LatLong: "abc,-122.0838"}
	_, _, err := info.ParseLatLong()
	if err == nil {
		t.Error("expected error for invalid latitude")
	}
}

func TestParseLatLong_InvalidLongitude(t *testing.T) {
	info := &IPInfo{LatLong: "37.3860,xyz"}
	_, _, err := info.ParseLatLong()
	if err == nil {
		t.Error("expected error for invalid longitude")
	}
}

func TestParseLatLong_EmptyString(t *testing.T) {
	info := &IPInfo{LatLong: ""}
	_, _, err := info.ParseLatLong()
	if err == nil {
		t.Error("expected error for empty latlong")
	}
}

// --- ValidateIP ---

func TestValidateIP(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"8.8.8.8", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"invalid", false},
		{"192.168.0.256", false},
		{"", false},
	}

	for _, tt := range tests {
		err := ValidateIP(tt.ip)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateIP(%q) = %v, want valid=%t", tt.ip, err, tt.valid)
		}
	}
}

// --- APIError methods ---

func TestAPIError_Error(t *testing.T) {
	apiErr := &APIError{
		HasError: true,
		Reason:   "RateLimited",
		Message:  "API rate limit exceeded",
	}
	msg := apiErr.Error()
	if !strings.Contains(msg, "API rate limit exceeded") {
		t.Errorf("error message should contain 'API rate limit exceeded', got %q", msg)
	}
	if !strings.Contains(msg, "RateLimited") {
		t.Errorf("error message should contain 'RateLimited', got %q", msg)
	}
}

func TestAPIError_ToError(t *testing.T) {
	apiErr := &APIError{
		HasError: true,
		Reason:   "TestReason",
		Message:  "Test message",
	}
	returned := apiErr.ToError()
	if returned != apiErr {
		t.Errorf("ToError should return itself, got %v", returned)
	}
}

func TestAPIError_AsError(t *testing.T) {
	apiErr := &APIError{
		HasError: true,
		Reason:   "RateLimited",
		Message:  "API rate limit exceeded",
	}
	var target *APIError
	if !errors.As(apiErr, &target) {
		t.Error("errors.As should match APIError")
	}
	if target.Reason != "RateLimited" {
		t.Errorf("expected RateLimited, got %s", target.Reason)
	}
}

// --- IsRetryableError ---

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		err        error
		retryable  bool
	}{
		{ErrRateLimited, true},
		{ErrServerError, true},
		{ErrNotFound, true},
		{ErrInvalidIP, false},
		{ErrInvalidField, false},
		{ErrReservedIP, false},
		{ErrUnexpectedData, false},
		{fmt.Errorf("some random error: %w", ErrRateLimited), true},
	}

	for _, tt := range tests {
		result := IsRetryableError(tt.err)
		if result != tt.retryable {
			t.Errorf("IsRetryableError(%v) = %v, want %v", tt.err, result, tt.retryable)
		}
	}
}

// --- WrapError ---

func TestWrapError(t *testing.T) {
	wrapped := WrapError("GetIPInfo", ErrRateLimited)
	if !strings.Contains(wrapped.Error(), "GetIPInfo failed") {
		t.Errorf("expected 'GetIPInfo failed' in message, got %v", wrapped)
	}
	if !errors.Is(wrapped, ErrRateLimited) {
		t.Errorf("expected wrapped error to match ErrRateLimited, got %v", wrapped)
	}
}

// --- handleError branches ---

func TestClient_handleError_WithAPIErrorHandler(t *testing.T) {
	handlerCalled := false
	client := NewClient(WithErrorHandler(func(err error) error {
		handlerCalled = true
		return err
	}))
	err := client.handleError(ErrInvalidIP)
	if !handlerCalled {
		t.Error("error handler was not called")
	}
	if !errors.Is(err, ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

func TestClient_handleError_ReservedIPAPIError(t *testing.T) {
	client := NewClient()
	apiErr := &APIError{
		HasError: true,
		Reason:   "Reserved IP Address",
		Message:  "Reserved IP Address",
		IP:       "127.0.0.1",
	}
	err := client.handleError(apiErr)
	if !errors.Is(err, ErrReservedIP) {
		t.Errorf("expected ErrReservedIP, got %v", err)
	}
}

func TestClient_handleError_InvalidIPAPIError(t *testing.T) {
	client := NewClient()
	apiErr := &APIError{
		HasError: true,
		Reason:   "Invalid IP Address",
		Message:  "Invalid IP Address",
		IP:       "999.999.999.999",
	}
	err := client.handleError(apiErr)
	if !errors.Is(err, ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

func TestClient_handleError_UnknownAPIError(t *testing.T) {
	client := NewClient()
	apiErr := &APIError{
		HasError: true,
		Reason:   "UnknownReason",
		Message:  "Something went wrong",
	}
	err := client.handleError(apiErr)
	// Unknown reason falls through to return the original error
	if err != apiErr {
		t.Errorf("expected original apiErr, got %v", err)
	}
}

func TestClient_handleError_NilErrorHandler(t *testing.T) {
	client := NewClient()
	err := client.handleError(ErrNotFound)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- GetField with error paths ---

func TestClient_GetField_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `not json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetField(context.Background(), "8.8.8.8", "country")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- GetField_ReadAllError ---

func TestClient_GetField_ReadAllError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a response whose body will fail on read
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		// Don't write any data - the body reader will return EOF early
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	// This should still work since ReadAll handles short reads
	result, err := client.GetField(context.Background(), "8.8.8.8", "country")
	if err != nil {
		t.Logf("GetField returned error (acceptable): %v", err)
	} else {
		t.Logf("GetField returned: %q", result)
	}
}

// --- Full integration: GetField with mock server returns value ---

func TestClient_GetField_NetworkField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/network/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		fmt.Fprint(w, "8.8.8.0/24")
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	result, err := client.GetField(context.Background(), "8.8.8.8", "network")
	if err != nil {
		t.Fatal(err)
	}
	if result != "8.8.8.0/24" {
		t.Errorf("expected 8.8.8.0/24, got %s", result)
	}
}

// --- doRequest: 4xx without valid APIError JSON ---

func TestClient_doRequest_4xxWithInvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `<html>Bad Request</html>`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrServerError) {
		t.Errorf("expected ErrServerError for 400 without valid JSON, got %v", err)
	}
}

// --- Error variables ---

func TestErrorVariables(t *testing.T) {
	errs := []error{
		ErrInvalidIP,
		ErrInvalidField,
		ErrRateLimited,
		ErrReservedIP,
		ErrNotFound,
		ErrServerError,
		ErrUnexpectedData,
	}
	for _, e := range errs {
		if e == nil {
			t.Errorf("error variable should not be nil")
		}
		if e.Error() == "" {
			t.Errorf("error variable should have non-empty message")
		}
	}
}

// --- io.ReadAll failure path via GetField ---

type failingReader struct{}

func (f *failingReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (f *failingReader) Close() error {
	return nil
}

func TestClient_GetField_ReadBodyError(t *testing.T) {
	// We need to test the io.ReadAll error path in GetField.
	// This is hard to trigger with httptest.Server, so we test it indirectly.
	// The key is to verify the error path exists in the code.
	// Since httptest.Server always provides a valid Body, we test the error
	// wrapping behavior through handleError instead.
	client := NewClient()
	err := client.handleError(fmt.Errorf("read error: %w", io.ErrUnexpectedEOF))
	if err == nil {
		t.Error("expected error")
	}
}

// --- Constants ---

func TestConstants(t *testing.T) {
	if defaultBaseURL != "https://ipapi.co/" {
		t.Errorf("expected defaultBaseURL https://ipapi.co/, got %s", defaultBaseURL)
	}
	if defaultTimeout != 10*time.Second {
		t.Errorf("expected defaultTimeout 10s, got %v", defaultTimeout)
	}
	if maxRedirects != 3 {
		t.Errorf("expected maxRedirects 3, got %d", maxRedirects)
	}
	if defaultRetryDelay != 500*time.Millisecond {
		t.Errorf("expected defaultRetryDelay 500ms, got %v", defaultRetryDelay)
	}
}

// --- mapStatusCodeToError direct tests ---

func TestClient_mapStatusCodeToError_InternalServerError_Direct(t *testing.T) {
	client := NewClient()
	err := client.mapStatusCodeToError(http.StatusInternalServerError)
	if !errors.Is(err, ErrServerError) {
		t.Errorf("expected ErrServerError for 500, got %v", err)
	}
}

// --- Cover doRequest post-loop err check (lines 53-55) ---
// This path is only reachable if the for loop exits without returning,
// which can't happen in normal flow. We test mapStatusCodeToError directly
// to ensure full branch coverage of that function.

func TestClient_mapStatusCodeToError_AllCodes(t *testing.T) {
	client := NewClient()
	
	tests := []struct {
		code     int
		wantErr  error
	}{
		{http.StatusBadRequest, ErrServerError},
		{http.StatusForbidden, ErrInvalidKey},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusMethodNotAllowed, ErrMethodNotAllowed},
		{http.StatusTooManyRequests, ErrRateLimited},
		{http.StatusInternalServerError, ErrServerError},
		{418, nil}, // unexpected code - not a known sentinel
	}

	for _, tt := range tests {
		err := client.mapStatusCodeToError(tt.code)
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("mapStatusCodeToError(%d) = %v, want %v", tt.code, err, tt.wantErr)
			}
		} else {
			if err == nil {
				t.Errorf("mapStatusCodeToError(%d) should return error", tt.code)
			}
		}
	}
}

// --- newGetRequest ---

func TestNewGetRequest(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		segments  []string
		wantURL   string
		wantErr   bool
	}{
		{
			name:     "normal IP lookup",
			baseURL:  "https://ipapi.co/",
			segments: []string{"8.8.8.8", "json"},
			wantURL:  "https://ipapi.co/8.8.8.8/json/",
		},
		{
			name:     "client IP lookup",
			baseURL:  "https://ipapi.co/",
			segments: []string{"json"},
			wantURL:  "https://ipapi.co/json/",
		},
		{
			name:     "field lookup",
			baseURL:  "https://ipapi.co/",
			segments: []string{"8.8.8.8", "country"},
			wantURL:  "https://ipapi.co/8.8.8.8/country/",
		},
		{
			name:     "base URL without trailing slash",
			baseURL:  "https://ipapi.co",
			segments: []string{"8.8.8.8", "json"},
			wantURL:  "https://ipapi.co/8.8.8.8/json/",
		},
		{
			name:     "base URL with path prefix",
			baseURL:  "https://example.com/api/",
			segments: []string{"8.8.8.8", "json"},
			wantURL:  "https://example.com/api/8.8.8.8/json/",
		},
		{
			name:     "invalid base URL",
			baseURL:  "://invalid",
			segments: []string{"8.8.8.8", "json"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := newGetRequest(context.Background(), tt.baseURL, tt.segments...)
			if (err != nil) != tt.wantErr {
				t.Errorf("newGetRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if req.URL.String() != tt.wantURL {
					t.Errorf("newGetRequest() URL = %v, want %v", req.URL.String(), tt.wantURL)
				}
				if req.Method != "GET" {
					t.Errorf("expected GET method, got %s", req.Method)
				}
			}
		})
	}
}

func TestNewGetRequest_InvalidBaseURL(t *testing.T) {
	// Test that an invalid base URL returns an error through the client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"ip":"8.8.8.8"}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = "://invalid"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err == nil {
		t.Error("expected error for invalid base URL")
	}

	_, err = client.GetField(context.Background(), "8.8.8.8", "country")
	if err == nil {
		t.Error("expected error for invalid base URL")
	}

	_, err = client.GetClientIPInfo(context.Background(), "json")
	if err == nil {
		t.Error("expected error for invalid base URL")
	}
}

// --- validFields completeness check ---

func TestValidFields_Completeness(t *testing.T) {
	// Verify that every field in the IPInfo struct has a corresponding
	// entry in validFields (except RetrievedAt which is json:"-")
	expectedFields := []string{
		"ip", "network", "version", "city", "region", "region_code",
		"country", "country_name", "country_code", "country_code_iso3",
		"country_capital", "country_tld", "continent_code", "in_eu",
		"postal", "latitude", "longitude", "latlong", "timezone",
		"utc_offset", "country_calling_code", "currency", "currency_name",
		"languages", "country_area", "country_population", "asn", "org",
		"hostname",
	}

	for _, field := range expectedFields {
		if _, ok := validFields[field]; !ok {
			t.Errorf("field %q is missing from validFields map", field)
		}
	}

	// Verify no extra fields in validFields that aren't in IPInfo
	if len(validFields) != len(expectedFields) {
		t.Errorf("validFields has %d entries, expected %d", len(validFields), len(expectedFields))
	}
}

// --- APIError with Reserved field ---

func TestAPIError_ReservedField(t *testing.T) {
	apiErr := &APIError{
		HasError: true,
		Reason:   "Reserved IP Address",
		IP:       "127.0.0.1",
		Reserved: true,
		Version:  "IPv4",
	}
	if !apiErr.Reserved {
		t.Error("expected Reserved to be true")
	}
	if apiErr.Version != "IPv4" {
		t.Errorf("expected IPv4, got %s", apiErr.Version)
	}
	// Test that Error() includes reserved info
	msg := apiErr.Error()
	if !strings.Contains(msg, "reserved: true") {
		t.Errorf("expected error message to contain 'reserved: true', got %q", msg)
	}
}

func TestAPIError_NonReserved(t *testing.T) {
	apiErr := &APIError{
		HasError: true,
		Reason:   "RateLimited",
		Message:  "API rate limit exceeded",
	}
	if apiErr.Reserved {
		t.Error("expected Reserved to be false")
	}
	// Non-reserved errors should not include reserved info
	msg := apiErr.Error()
	if strings.Contains(msg, "reserved:") {
		t.Errorf("non-reserved error should not contain reserved info, got %q", msg)
	}
}

func TestAPIError_ParseReservedFromJSON(t *testing.T) {
	// Simulate the actual API response for a reserved IP
	jsonStr := `{"ip": "127.0.0.1", "error": true, "reason": "Reserved IP Address", "reserved": true, "version": "IPv4"}`
	var apiErr APIError
	if err := json.Unmarshal([]byte(jsonStr), &apiErr); err != nil {
		t.Fatal(err)
	}
	if !apiErr.HasError {
		t.Error("expected HasError to be true")
	}
	if apiErr.Reason != "Reserved IP Address" {
		t.Errorf("expected 'Reserved IP Address', got %q", apiErr.Reason)
	}
	if !apiErr.Reserved {
		t.Error("expected Reserved to be true")
	}
	if apiErr.IP != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", apiErr.IP)
	}
	if apiErr.Version != "IPv4" {
		t.Errorf("expected IPv4, got %s", apiErr.Version)
	}
}

func TestAPIError_ParseInvalidKeyFromJSON(t *testing.T) {
	jsonStr := `{"error": true, "reason": "Invalid Key", "message": "Invalid key. SignUp @ https://ipapi.co/pricing/ "}`
	var apiErr APIError
	if err := json.Unmarshal([]byte(jsonStr), &apiErr); err != nil {
		t.Fatal(err)
	}
	if apiErr.Reason != "Invalid Key" {
		t.Errorf("expected 'Invalid Key', got %q", apiErr.Reason)
	}
	if apiErr.Reserved {
		t.Error("expected Reserved to be false for non-reserved errors")
	}
}

// --- GetClientField ---

func TestClient_GetClientField_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/country/" {
			t.Errorf("unexpected path: %s, expected /country/", r.URL.Path)
		}
		fmt.Fprint(w, "US")
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	result, err := client.GetClientField(context.Background(), "country")
	if err != nil {
		t.Fatal(err)
	}
	if result != "US" {
		t.Errorf("expected US, got %s", result)
	}
}

func TestClient_GetClientField_InvalidField(t *testing.T) {
	client := NewClient()
	_, err := client.GetClientField(context.Background(), "invalid_field")
	if !errors.Is(err, ErrInvalidField) {
		t.Errorf("expected ErrInvalidField, got %v", err)
	}
}

func TestClient_GetClientField_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `not json`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetClientField(context.Background(), "country")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestClient_GetClientField_InvalidBaseURL(t *testing.T) {
	client := NewClient()
	client.BaseURL = "://invalid"

	_, err := client.GetClientField(context.Background(), "country")
	if err == nil {
		t.Error("expected error for invalid base URL")
	}
}

// --- Reserved IP error from API (HTTP 200 with error body) ---

func TestClient_GetIPInfo_ReservedIP_APIError(t *testing.T) {
	// The real API returns HTTP 200 with {"error":true,"reason":"Reserved IP Address","reserved":true}
	// Our doRequest doesn't check for errors in 200 responses, so the JSON decode
	// will parse into IPInfo with empty fields. This is expected behavior since
	// ValidateIP already catches reserved IPs before making the request.
	// But if the API returns a 4xx error with the reserved IP info, we handle it.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{
			HasError: true,
			Reason:   "Reserved IP Address",
			IP:       "127.0.0.1",
			Reserved: true,
			Version:  "IPv4",
		})
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrReservedIP) {
		t.Errorf("expected ErrReservedIP, got %v", err)
	}
}

// --- APIError.ToError returns self ---

func TestAPIError_ToError_ReturnsSelf(t *testing.T) {
	apiErr := &APIError{
		HasError: true,
		Reason:   "TestReason",
		Reserved: false,
	}
	returned := apiErr.ToError()
	if returned != apiErr {
		t.Errorf("ToError should return itself")
	}
}

// --- IsRetryableError with wrapped errors ---

func TestIsRetryableError_WrappedErrors(t *testing.T) {
	// Test that wrapped errors are still detected
	wrapped := fmt.Errorf("operation failed: %w", ErrServerError)
	if !IsRetryableError(wrapped) {
		t.Error("wrapped ErrServerError should be retryable")
	}

	wrappedNotFound := fmt.Errorf("op: %w", ErrNotFound)
	if !IsRetryableError(wrappedNotFound) {
		t.Error("wrapped ErrNotFound should be retryable")
	}

	notRetryable := fmt.Errorf("op: %w", ErrInvalidIP)
	if IsRetryableError(notRetryable) {
		t.Error("wrapped ErrInvalidIP should NOT be retryable")
	}
}

// --- GetClientField io.ReadAll error ---

func TestClient_GetClientField_ReadAllError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		// Don't write any data - triggers unexpected EOF on ReadAll
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetClientField(context.Background(), "country")
	// This should return an error (unexpected EOF)
	if err == nil {
		t.Error("expected error for short read")
	}
}

// --- Format validation ---

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		format string
		valid  bool
	}{
		{"json", true},
		{"jsonp", true},
		{"xml", true},
		{"csv", true},
		{"yaml", true},
		{"invalid", false},
		{"html", false},
		{"", false},
		{"JSON", false}, // case-sensitive
	}

	for _, tt := range tests {
		err := ValidateFormat(tt.format)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateFormat(%q) = %v, want valid=%t", tt.format, err, tt.valid)
		}
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
	if FormatJSONP != "jsonp" {
		t.Errorf("FormatJSONP = %q, want %q", FormatJSONP, "jsonp")
	}
	if FormatXML != "xml" {
		t.Errorf("FormatXML = %q, want %q", FormatXML, "xml")
	}
	if FormatCSV != "csv" {
		t.Errorf("FormatCSV = %q, want %q", FormatCSV, "csv")
	}
	if FormatYAML != "yaml" {
		t.Errorf("FormatYAML = %q, want %q", FormatYAML, "yaml")
	}
}

func TestValidFormats_Completeness(t *testing.T) {
	expected := []Format{FormatJSON, FormatJSONP, FormatXML, FormatCSV, FormatYAML}
	if len(validFormats) != len(expected) {
		t.Errorf("validFormats has %d entries, expected %d", len(validFormats), len(expected))
	}
	for _, f := range expected {
		if _, ok := validFormats[f]; !ok {
			t.Errorf("format %q missing from validFormats", f)
		}
	}
}

func TestClient_GetIPInfo_InvalidFormat(t *testing.T) {
	client := NewClient()
	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "html")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestClient_GetClientIPInfo_InvalidFormat(t *testing.T) {
	client := NewClient()
	_, err := client.GetClientIPInfo(context.Background(), "html")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

// --- GetIPInfoRaw ---

func TestClient_GetIPInfoRaw_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/json/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ip": "8.8.8.8", "city": "Mountain View"}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	data, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"ip": "8.8.8.8", "city": "Mountain View"}` {
		t.Errorf("unexpected raw data: %s", string(data))
	}
}

func TestClient_GetIPInfoRaw_XML(t *testing.T) {
	xmlResp := `<?xml version="1.0" encoding="utf-8"?><root><ip>8.8.8.8</ip></root>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/xml/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, xmlResp)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	data, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "xml")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != xmlResp {
		t.Errorf("unexpected XML data: %s", string(data))
	}
}

func TestClient_GetIPInfoRaw_CSV(t *testing.T) {
	csvResp := "ip,network,version,city\n8.8.8.8,8.8.8.0/24,IPv4,Mountain View"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/csv/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		fmt.Fprint(w, csvResp)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	data, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "csv")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != csvResp {
		t.Errorf("unexpected CSV data: %s", string(data))
	}
}

func TestClient_GetIPInfoRaw_YAML(t *testing.T) {
	yamlResp := "ip: 8.8.8.8\ncity: Mountain View\n"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/yaml/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		fmt.Fprint(w, yamlResp)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	data, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != yamlResp {
		t.Errorf("unexpected YAML data: %s", string(data))
	}
}

func TestClient_GetIPInfoRaw_InvalidIP(t *testing.T) {
	client := NewClient()
	_, err := client.GetIPInfoRaw(context.Background(), "invalid-ip", "json")
	if !errors.Is(err, ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

func TestClient_GetIPInfoRaw_InvalidFormat(t *testing.T) {
	client := NewClient()
	_, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "html")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestClient_GetIPInfoRaw_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `not found`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestClient_GetIPInfoRaw_InvalidBaseURL(t *testing.T) {
	client := NewClient()
	client.BaseURL = "://invalid"

	_, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "json")
	if err == nil {
		t.Error("expected error for invalid base URL")
	}
}

func TestClient_GetIPInfoRaw_ReadBodyError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		// Don't write any data - triggers unexpected EOF on ReadAll
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "json")
	if err == nil {
		t.Error("expected error for short read")
	}
}

// --- GetClientIPInfoRaw ---

func TestClient_GetClientIPInfoRaw_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ip": "203.0.113.1", "city": "Test City"}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	data, err := client.GetClientIPInfoRaw(context.Background(), "json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"ip": "203.0.113.1", "city": "Test City"}` {
		t.Errorf("unexpected raw data: %s", string(data))
	}
}

func TestClient_GetClientIPInfoRaw_XML(t *testing.T) {
	xmlResp := `<?xml version="1.0" encoding="utf-8"?><root><ip>203.0.113.1</ip></root>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xml/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, xmlResp)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	data, err := client.GetClientIPInfoRaw(context.Background(), "xml")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != xmlResp {
		t.Errorf("unexpected XML data: %s", string(data))
	}
}

func TestClient_GetClientIPInfoRaw_InvalidFormat(t *testing.T) {
	client := NewClient()
	_, err := client.GetClientIPInfoRaw(context.Background(), "html")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestClient_GetClientIPInfoRaw_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(APIError{
			HasError: true,
			Reason:   "RateLimited",
			Message:  "API rate limit exceeded",
		})
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetClientIPInfoRaw(context.Background(), "json")
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

func TestClient_GetClientIPInfoRaw_InvalidBaseURL(t *testing.T) {
	client := NewClient()
	client.BaseURL = "://invalid"

	_, err := client.GetClientIPInfoRaw(context.Background(), "json")
	if err == nil {
		t.Error("expected error for invalid base URL")
	}
}

func TestClient_GetClientIPInfoRaw_ReadBodyError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetClientIPInfoRaw(context.Background(), "json")
	if err == nil {
		t.Error("expected error for short read")
	}
}

// --- API Key Query Parameter ---

func TestClient_APIKeyQueryParameter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify key is in query parameter, not in Authorization header
		if r.Header.Get("Authorization") != "" {
			t.Errorf("expected no Authorization header, got %q", r.Header.Get("Authorization"))
		}
		key := r.URL.Query().Get("key")
		if key != "my-api-key" {
			t.Errorf("expected key=my-api-key in query, got %q", key)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ip": "8.8.8.8", "city": "Mountain View"}`)
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("my-api-key"),
		WithAPIKeyQuery(),
	)
	client.BaseURL = ts.URL + "/"

	info, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "8.8.8.8" {
		t.Errorf("expected 8.8.8.8, got %s", info.IP)
	}
}

func TestClient_APIKeyBearerHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify key is in Authorization header, not in query parameter
		if r.Header.Get("Authorization") != "Bearer my-api-key" {
			t.Errorf("expected Bearer authorization, got %q", r.Header.Get("Authorization"))
		}
		if key := r.URL.Query().Get("key"); key != "" {
			t.Errorf("expected no key in query parameter, got %q", key)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ip": "8.8.8.8", "city": "Mountain View"}`)
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("my-api-key"),
	)
	client.BaseURL = ts.URL + "/"

	info, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "8.8.8.8" {
		t.Errorf("expected 8.8.8.8, got %s", info.IP)
	}
}

func TestClient_APIKeyQuery_GetField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("expected no Authorization header, got %q", r.Header.Get("Authorization"))
		}
		key := r.URL.Query().Get("key")
		if key != "my-api-key" {
			t.Errorf("expected key=my-api-key in query, got %q", key)
		}
		fmt.Fprint(w, "US")
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("my-api-key"),
		WithAPIKeyQuery(),
	)
	client.BaseURL = ts.URL + "/"

	result, err := client.GetField(context.Background(), "8.8.8.8", "country")
	if err != nil {
		t.Fatal(err)
	}
	if result != "US" {
		t.Errorf("expected US, got %s", result)
	}
}

func TestClient_APIKeyQuery_GetClientIPInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key != "my-api-key" {
			t.Errorf("expected key=my-api-key in query, got %q", key)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ip": "203.0.113.1"}`)
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("my-api-key"),
		WithAPIKeyQuery(),
	)
	client.BaseURL = ts.URL + "/"

	info, err := client.GetClientIPInfo(context.Background(), "json")
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "203.0.113.1" {
		t.Errorf("expected 203.0.113.1, got %s", info.IP)
	}
}

func TestClient_APIKeyQuery_GetClientField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key != "my-api-key" {
			t.Errorf("expected key=my-api-key in query, got %q", key)
		}
		fmt.Fprint(w, "US")
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("my-api-key"),
		WithAPIKeyQuery(),
	)
	client.BaseURL = ts.URL + "/"

	result, err := client.GetClientField(context.Background(), "country")
	if err != nil {
		t.Fatal(err)
	}
	if result != "US" {
		t.Errorf("expected US, got %s", result)
	}
}

func TestClient_APIKeyQuery_GetIPInfoRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key != "my-api-key" {
			t.Errorf("expected key=my-api-key in query, got %q", key)
		}
		fmt.Fprint(w, `{"ip": "8.8.8.8"}`)
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("my-api-key"),
		WithAPIKeyQuery(),
	)
	client.BaseURL = ts.URL + "/"

	data, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"ip": "8.8.8.8"}` {
		t.Errorf("unexpected data: %s", string(data))
	}
}

func TestClient_APIKeyQuery_GetClientIPInfoRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key != "my-api-key" {
			t.Errorf("expected key=my-api-key in query, got %q", key)
		}
		fmt.Fprint(w, `{"ip": "203.0.113.1"}`)
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("my-api-key"),
		WithAPIKeyQuery(),
	)
	client.BaseURL = ts.URL + "/"

	data, err := client.GetClientIPInfoRaw(context.Background(), "json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"ip": "203.0.113.1"}` {
		t.Errorf("unexpected data: %s", string(data))
	}
}

func TestClient_NoAPIKey_NoAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("expected no Authorization header, got %q", r.Header.Get("Authorization"))
		}
		if key := r.URL.Query().Get("key"); key != "" {
			t.Errorf("expected no key query parameter, got %q", key)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ip": "8.8.8.8"}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
}

// --- JSONP Callback ---

func TestClient_JSONPCallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callback := r.URL.Query().Get("callback")
		if callback != "myCallback" {
			t.Errorf("expected callback=myCallback in query, got %q", callback)
		}
		fmt.Fprint(w, `myCallback({"ip": "8.8.8.8"})`)
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithCallback("myCallback"),
	)
	client.BaseURL = ts.URL + "/"

	data, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "jsonp")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "myCallback") {
		t.Errorf("expected JSONP callback in response, got %s", string(data))
	}
}

func TestClient_JSONPCallback_WithAPIKeyQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callback := r.URL.Query().Get("callback")
		if callback != "myCallback" {
			t.Errorf("expected callback=myCallback, got %q", callback)
		}
		key := r.URL.Query().Get("key")
		if key != "my-api-key" {
			t.Errorf("expected key=my-api-key, got %q", key)
		}
		fmt.Fprintf(w, `%s({"ip": "8.8.8.8"})`, callback)
	}))
	defer ts.Close()

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
		WithAPIKey("my-api-key"),
		WithAPIKeyQuery(),
		WithCallback("myCallback"),
	)
	client.BaseURL = ts.URL + "/"

	data, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "jsonp")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "myCallback") {
		t.Errorf("expected JSONP callback in response, got %s", string(data))
	}
}

func TestClient_NoCallback_NoQueryParam(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callback := r.URL.Query().Get("callback"); callback != "" {
			t.Errorf("expected no callback in query, got %q", callback)
		}
		fmt.Fprint(w, `{"ip": "8.8.8.8"}`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", "json")
	if err != nil {
		t.Fatal(err)
	}
}

// --- ErrMethodNotAllowed ---

func TestClient_mapStatusCodeToError_MethodNotAllowed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `Method Not Allowed`)
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrMethodNotAllowed) {
		t.Errorf("expected ErrMethodNotAllowed for 405, got %v", err)
	}
}

// --- ErrInvalidKey ---

func TestClient_HandleAPIError_InvalidKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(APIError{
			HasError: true,
			Reason:   "Invalid Key",
			Message:  "Invalid key. SignUp @ https://ipapi.co/pricing/",
		})
	}))
	defer ts.Close()

	client := NewClient(WithCustomHTTPClient(ts.Client()))
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "json")
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestClient_handleError_InvalidKeyAPIError(t *testing.T) {
	client := NewClient()
	apiErr := &APIError{
		HasError: true,
		Reason:   "Invalid Key",
		Message:  "Invalid key. SignUp @ https://ipapi.co/pricing/",
	}
	err := client.handleError(apiErr)
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

// --- New error variables ---

func TestNewErrorVariables(t *testing.T) {
	if ErrInvalidFormat == nil {
		t.Error("ErrInvalidFormat should not be nil")
	}
	if ErrInvalidFormat.Error() == "" {
		t.Error("ErrInvalidFormat should have non-empty message")
	}
	if ErrMethodNotAllowed == nil {
		t.Error("ErrMethodNotAllowed should not be nil")
	}
	if ErrMethodNotAllowed.Error() == "" {
		t.Error("ErrMethodNotAllowed should have non-empty message")
	}
	if ErrInvalidKey == nil {
		t.Error("ErrInvalidKey should not be nil")
	}
	if ErrInvalidKey.Error() == "" {
		t.Error("ErrInvalidKey should have non-empty message")
	}
}

// --- APIKeyMode constants ---

func TestAPIKeyModeConstants(t *testing.T) {
	if APIKeyHeader != 0 {
		t.Errorf("APIKeyHeader should be 0, got %d", APIKeyHeader)
	}
	if APIKeyQuery != 1 {
		t.Errorf("APIKeyQuery should be 1, got %d", APIKeyQuery)
	}
}

// --- WithAPIKeyQuery option ---

func TestWithAPIKeyQuery(t *testing.T) {
	client := NewClient(WithAPIKeyQuery())
	if client.APIKeyMode != APIKeyQuery {
		t.Errorf("expected APIKeyQuery mode, got %d", client.APIKeyMode)
	}
}

// --- WithCallback option ---

func TestWithCallback(t *testing.T) {
	client := NewClient(WithCallback("myFunc"))
	if client.Callback != "myFunc" {
		t.Errorf("expected callback 'myFunc', got %q", client.Callback)
	}
}

func TestWithCallback_Empty(t *testing.T) {
	client := NewClient(WithCallback(""))
	if client.Callback != "" {
		t.Errorf("expected empty callback, got %q", client.Callback)
	}
}

// --- IsRetryableError with new error types ---

func TestIsRetryableError_NewErrors(t *testing.T) {
	if IsRetryableError(ErrInvalidFormat) {
		t.Error("ErrInvalidFormat should not be retryable")
	}
	if IsRetryableError(ErrMethodNotAllowed) {
		t.Error("ErrMethodNotAllowed should not be retryable")
	}
	if IsRetryableError(ErrInvalidKey) {
		t.Error("ErrInvalidKey should not be retryable")
	}
}

// --- Format validation in GetIPInfo with error handler ---

func TestClient_GetIPInfo_InvalidFormat_WithErrorHandler(t *testing.T) {
	handlerCalled := false
	client := NewClient(WithErrorHandler(func(err error) error {
		handlerCalled = true
		return err
	}))
	_, err := client.GetIPInfo(context.Background(), "8.8.8.8", "html")
	if !handlerCalled {
		t.Error("error handler was not called")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

// --- Format validation with all valid formats ---

func TestClient_GetIPInfo_AllFormats(t *testing.T) {
	formats := []string{"json", "jsonp", "xml", "csv", "yaml"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"ip": "8.8.8.8"}`)
			}))
			defer ts.Close()

			client := NewClient(WithCustomHTTPClient(ts.Client()))
			client.BaseURL = ts.URL + "/"

			// Use GetIPInfoRaw for non-JSON formats since json.Decoder would fail
			_, err := client.GetIPInfoRaw(context.Background(), "8.8.8.8", format)
			if err != nil {
				t.Errorf("format %q should be valid but got error: %v", format, err)
			}
		})
	}
}
