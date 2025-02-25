// models.go
package ipapi

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type IPInfo struct {
	IP                 string    `json:"ip"`
	Version            string    `json:"version"`
	City               string    `json:"city"`
	Region             string    `json:"region"`
	RegionCode         string    `json:"region_code"`
	Country            string    `json:"country"`
	CountryName        string    `json:"country_name"`
	CountryCodeISO3    string    `json:"country_code_iso3"`
	CountryCapital     string    `json:"country_capital"`
	CountryTLD         string    `json:"country_tld"`
	ContinentCode      string    `json:"continent_code"`
	InEU               bool      `json:"in_eu"`
	Postal             string    `json:"postal"`
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

type APIError struct {
	HasError bool   `json:"error"`
	Reason   string `json:"reason"`
	Message  string `json:"message"`
	IP       string `json:"ip"`
}

// 实现error接口
func (e *APIError) Error() string {
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

func ValidateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return ErrInvalidIP
	}
	return nil
}
