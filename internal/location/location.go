package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Location represents the geographic location of the user.
type Location struct {
	City    string  `json:"city"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// IPApiResponse maps the JSON response from ipwho.is
type IPApiResponse struct {
	Success     bool    `json:"success"`
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Message     string  `json:"message"`
	Reason      string  `json:"reason"`
}

// defaultAPIURL is the production geolocation endpoint. It must stay HTTPS.
const defaultAPIURL = "https://ipwho.is/"

// apiURL is a package-level var so tests can override it with an httptest
// server. Do NOT downgrade it to plain HTTP: that would leak the user's IP
// address in cleartext and let a network attacker forge the response.
var apiURL = defaultAPIURL

// DetectLocation uses ipwho.is over HTTPS to detect the user's location based
// on their IP address.
func DetectLocation() (Location, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return Location{}, fmt.Errorf("failed to fetch IP geolocation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Location{}, fmt.Errorf("IP geolocation API returned status: %d", resp.StatusCode)
	}

	var ipResp IPApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&ipResp); err != nil {
		return Location{}, fmt.Errorf("failed to decode IP geolocation response: %w", err)
	}

	if !ipResp.Success {
		detail := ipResp.Message
		if detail == "" {
			detail = ipResp.Reason
		}
		return Location{}, fmt.Errorf("IP geolocation failed: %s", detail)
	}

	return Location{
		City:    ipResp.City,
		Country: ipResp.Country,
		Lat:     ipResp.Latitude,
		Lon:     ipResp.Longitude,
	}, nil
}
