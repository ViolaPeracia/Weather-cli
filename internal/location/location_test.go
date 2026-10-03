package location

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetectLocationSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := IPApiResponse{
			Success:     true,
			IP:          "113.190.231.55",
			Country:     "Vietnam",
			CountryCode: "VN",
			City:        "Hanoi",
			Latitude:    21.0245,
			Longitude:   105.8412,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	oldAPIURL := apiURL
	apiURL = server.URL + "/"
	defer func() { apiURL = oldAPIURL }()

	loc, err := DetectLocation()
	if err != nil {
		t.Fatalf("DetectLocation() failed: %v", err)
	}

	if loc.City != "Hanoi" {
		t.Errorf("Expected city 'Hanoi', got '%s'", loc.City)
	}
	if loc.Country != "Vietnam" {
		t.Errorf("Expected country 'Vietnam', got '%s'", loc.Country)
	}
	if loc.Lat != 21.0245 || loc.Lon != 105.8412 {
		t.Errorf("Expected lat/lon 21.0245, 105.8412, got %f, %f", loc.Lat, loc.Lon)
	}
}

func TestDetectLocationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := IPApiResponse{
			Success: false,
			Message: "invalid query",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	oldAPIURL := apiURL
	apiURL = server.URL + "/"
	defer func() { apiURL = oldAPIURL }()

	_, err := DetectLocation()
	if err == nil {
		t.Error("Expected error from DetectLocation when API reports success=false, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid query") {
		t.Errorf("Expected error to surface the provider message 'invalid query', got %q", err.Error())
	}
}

func TestDetectLocationFailureFallsBackToReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := IPApiResponse{
			Success: false,
			Reason:  "Reserved range",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	oldAPIURL := apiURL
	apiURL = server.URL + "/"
	defer func() { apiURL = oldAPIURL }()

	_, err := DetectLocation()
	if err == nil {
		t.Fatal("Expected error from DetectLocation when success=false and reason is set, got nil")
	}
	if !strings.Contains(err.Error(), "Reserved range") {
		t.Errorf("Expected error to fall back to the reason field, got %q", err.Error())
	}
}

func TestDetectLocationHttpError(t *testing.T) {
	// Point at an invalid address so the request fails fast.
	oldAPIURL := apiURL
	apiURL = "http://invalid-endpoint-that-should-fail-immediately.local"
	defer func() { apiURL = oldAPIURL }()

	_, err := DetectLocation()
	if err == nil {
		t.Error("Expected error from DetectLocation when HTTP call fails, got nil")
	}
}

// TestDetectLocationUsesHTTPS is a regression guard: the geolocation lookup
// must never fall back to plaintext HTTP, otherwise the user's IP leaks and a
// man-in-the-middle can forge the returned weather location.
func TestDetectLocationUsesHTTPS(t *testing.T) {
	if !strings.HasPrefix(defaultAPIURL, "https://") {
		t.Errorf("default IP geolocation URL must use HTTPS, got %q", defaultAPIURL)
	}
	if !strings.HasPrefix(apiURL, "https://") {
		t.Errorf("active IP geolocation URL must use HTTPS, got %q", apiURL)
	}
}
