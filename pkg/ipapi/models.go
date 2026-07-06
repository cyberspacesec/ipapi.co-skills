// models.go
package ipapi

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type IPInfo struct {
	IP                 string    `json:"ip"`
	Network            string    `json:"network"`
	Version            string    `json:"version"`
	City               string    `json:"city"`
	Region             string    `json:"region"`
	RegionCode         string    `json:"region_code"`
	Country            string    `json:"country"`
	CountryName        string    `json:"country_name"`
	CountryCode        string    `json:"country_code"`
	CountryCodeISO3    string    `json:"country_code_iso3"`
	CountryCapital     string    `json:"country_capital"`
	CountryTLD         string    `json:"country_tld"`
	ContinentCode      string    `json:"continent_code"`
	InEU               bool      `json:"in_eu"`
	Postal             *string   `json:"postal"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	LatLong            string    `json:"latlong"`
	Timezone           string    `json:"timezone"`
	UTCOffset          string    `json:"utc_offset"`
	CountryCallingCode string    `json:"country_calling_code"`
	Currency           string    `json:"currency"`
	CurrencyName       string    `json:"currency_name"`
	Languages          string    `json:"languages"`
	CountryArea        float64   `json:"country_area"`
	CountryPopulation  int       `json:"country_population"`
	ASN                string    `json:"asn"`
	Org                string    `json:"org"`
	Hostname           string    `json:"hostname,omitempty"`
	RetrievedAt        time.Time `json:"-"`
}

// Quota holds the remaining IP lookup quota for the configured API key.
//
// It is returned by GetQuota, which queries the (undocumented but stable)
// GET /quota/ endpoint. The API responds with one of:
//   - {"available": <number>}  — remaining lookups for a valid key
//   - {"available": "API key needed"} — no API key configured on the client
//   - {"error": true, "reason": "Invalid Key", ...} — rejected key
//
// Available is the raw "available" value from the response; AvailableInt is
// the parsed integer count (only meaningful when Available == AvailableInt's
// string form, i.e. a numeric response).
type Quota struct {
	Available string `json:"available"`
}

// AvailableInt returns the remaining lookup count as an integer. It returns
// ok=false when the "available" field is not numeric (e.g. "API key needed"
// or an error response). Use this to drive quota-monitoring logic.
func (q Quota) AvailableInt() (n int, ok bool) {
	if q.Available == "" {
		return 0, false
	}
	i, err := strconv.Atoi(q.Available)
	if err != nil {
		return 0, false
	}
	return i, true
}

type APIError struct {
	HasError bool   `json:"error"`
	Reason   string `json:"reason"`
	Message  string `json:"message"`
	IP       string `json:"ip"`
	Reserved bool   `json:"reserved"`
	Version  string `json:"version"`
}

// 实现error接口
func (e *APIError) Error() string {
	if e.Reserved {
		return fmt.Sprintf("ipapi error: %s (reason: %s, ip: %s, reserved: %v)", e.Message, e.Reason, e.IP, e.Reserved)
	}
	return fmt.Sprintf("ipapi error: %s (reason: %s)", e.Message, e.Reason)
}

// 保留原有方法但修改实现
func (e *APIError) ToError() error {
	return e // 现在直接返回自己
}

func (info *IPInfo) ParseLatLong() (float64, float64, error) {
	parts := strings.Split(info.LatLong, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid latlong format")
	}
	var lat, lon float64
	_, err := fmt.Sscanf(parts[0], "%f", &lat)
	if err != nil {
		return 0, 0, err
	}
	_, err = fmt.Sscanf(parts[1], "%f", &lon)
	if err != nil {
		return 0, 0, err
	}
	return lat, lon, nil
}

// GetPostal returns the postal code as a string, or empty string if nil
func (info *IPInfo) GetPostal() string {
	if info.Postal == nil {
		return ""
	}
	return *info.Postal
}

func ValidateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return ErrInvalidIP
	}
	return nil
}
