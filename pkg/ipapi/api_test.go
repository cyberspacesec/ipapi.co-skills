// api_test.go
package ipapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetIPInfo_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/json/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"ip": "8.8.8.8",
			"city": "Mountain View",
			"country": "US",
			"latitude": 37.3860,
			"longitude": -122.0838
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
	if info.City != "Mountain View" {
		t.Errorf("expected Mountain View, got %s", info.City)
	}
	if info.Latitude != 37.3860 {
		t.Errorf("unexpected latitude: %f", info.Latitude)
	}
}

func TestClient_GetIPInfo_InvalidIP(t *testing.T) {
	client := NewClient()
	_, err := client.GetIPInfo(context.Background(), "invalid-ip", "json")
	if !errors.Is(err, ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

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

func TestClient_GetField_InvalidField(t *testing.T) {
	client := NewClient()
	_, err := client.GetField(context.Background(), "8.8.8.8", "invalid_field")
	if !errors.Is(err, ErrInvalidField) {
		t.Errorf("expected ErrInvalidField, got %v", err)
	}
}

func TestClient_HandleAPIError(t *testing.T) {
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

	client := NewClient(
		WithCustomHTTPClient(ts.Client()),
	)
	client.BaseURL = ts.URL + "/"

	_, err := client.GetIPInfo(ctx, "8.8.8.8", "json")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context canceled error, got %v", err)
	}
}

func TestParseLatLong(t *testing.T) {
	info := &IPInfo{LatLong: "37.3860,-122.0838"}
	lat, lon, err := info.ParseLatLong()
	if err != nil {
		t.Fatal(err)
	}
	if lat != 37.3860 || lon != -122.0838 {
		t.Errorf("unexpected lat/long: %f, %f", lat, lon)
	}

	info.LatLong = "invalid"
	_, _, err = info.ParseLatLong()
	if err == nil {
		t.Error("expected error for invalid latlong")
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"8.8.8.8", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"invalid", false},
		{"192.168.0.256", false},
	}

	for _, tt := range tests {
		err := ValidateIP(tt.ip)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateIP(%q) = %v, want valid=%t", tt.ip, err, tt.valid)
		}
	}
}
