package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestDecodeWeatherCode(t *testing.T) {
	tests := []struct {
		code         int
		expectedCond string
		expectedIcon string
	}{
		{0, "Clear sky", "󰖙"},
		{1, "Partly cloudy", "󰖕"},
		{2, "Partly cloudy", "󰖕"},
		{3, "Partly cloudy", "󰖕"},
		{45, "Fog", "󰖑"},
		{48, "Fog", "󰖑"},
		{51, "Drizzle", "󰖗"},
		{55, "Drizzle", "󰖗"},
		{56, "Freezing Drizzle", "󰖗"},
		{61, "Rain", "󰖖"},
		{65, "Rain", "󰖖"},
		{66, "Freezing Rain", "󰖖"},
		{71, "Snow fall", "󰖘"},
		{77, "Snow grains", "󰖘"},
		{80, "Rain showers", "󰖗"},
		{85, "Snow showers", "󰖘"},
		{95, "Thunderstorm", "󰖓"},
		{99, "Thunderstorm", "󰖓"},
		{-1, "Unknown", "󰖕"},
	}

	for _, tt := range tests {
		cond, icon := decodeWeatherCode(tt.code)
		if cond != tt.expectedCond {
			t.Errorf("For code %d, expected condition '%s', got '%s'", tt.code, tt.expectedCond, cond)
		}
		if icon != tt.expectedIcon {
			t.Errorf("For code %d, expected icon '%s', got '%s'", tt.code, tt.expectedIcon, icon)
		}
	}
}

func TestGeocodeCitySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := GeocodeResponse{}
		resp.Results = append(resp.Results, struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Name      string  `json:"name"`
			Country   string  `json:"country"`
		}{
			Latitude:  21.0245,
			Longitude: 105.8412,
			Name:      "Hanoi",
			Country:   "Vietnam",
		})
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	oldBaseURL := geocodeBaseURL
	geocodeBaseURL = server.URL
	defer func() { geocodeBaseURL = oldBaseURL }()

	lat, lon, name, err := GeocodeCity("Hanoi")
	if err != nil {
		t.Fatalf("GeocodeCity() failed: %v", err)
	}

	if lat != 21.0245 || lon != 105.8412 {
		t.Errorf("Expected coords 21.0245, 105.8412, got %f, %f", lat, lon)
	}
	if name != "Hanoi, Vietnam" {
		t.Errorf("Expected name 'Hanoi, Vietnam', got '%s'", name)
	}
}

func TestGeocodeCityNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := GeocodeResponse{Results: nil}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	oldBaseURL := geocodeBaseURL
	geocodeBaseURL = server.URL
	defer func() { geocodeBaseURL = oldBaseURL }()

	_, _, _, err := GeocodeCity("NonExistentCity")
	if err == nil {
		t.Error("Expected error for non-existent city, got nil")
	}
}

func TestFetchWeatherSuccess(t *testing.T) {
	// Mock honours the requested forecast_days so the fixture reflects what the
	// real API would return: today plus the requested upcoming days.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := WeatherResponse{}
		resp.Current.Temperature2m = 28.2
		resp.Current.RelativeHumidity = 91
		resp.Current.WeatherCode = 95

		resp.Daily.Time = []string{"2026-05-30", "2026-05-31"}
		resp.Daily.Temperature2mMax = []float64{32.0, 31.0}
		resp.Daily.Temperature2mMin = []float64{25.0, 24.0}
		resp.Daily.WeatherCode = []int{95, 3}

		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	oldBaseURL := weatherBaseURL
	weatherBaseURL = server.URL
	defer func() { weatherBaseURL = oldBaseURL }()

	data, err := FetchWeather(21.0245, 105.8412, "celsius", 1)
	if err != nil {
		t.Fatalf("FetchWeather() failed: %v", err)
	}

	if data.Temperature != 28.2 {
		t.Errorf("Expected temperature 28.2, got %f", data.Temperature)
	}
	if data.Humidity != 91 {
		t.Errorf("Expected humidity 91, got %d", data.Humidity)
	}
	if data.Conditions != "Thunderstorm" || data.Icon != "󰖓" {
		t.Errorf("Expected 'Thunderstorm' / '󰖓', got '%s' / '%s'", data.Conditions, data.Icon)
	}
	if data.Unit != "C" {
		t.Errorf("Expected unit symbol 'C', got '%s'", data.Unit)
	}

	// days=1 must yield exactly one UPCOMING day (2026-05-31), not today's entry.
	if len(data.Forecast) != 1 {
		t.Errorf("Expected 1 forecast day, got %d", len(data.Forecast))
	} else {
		f := data.Forecast[0]
		if f.Date != "2026-05-31" {
			t.Errorf("Expected forecast date '2026-05-31', got '%s'", f.Date)
		}
		if f.MaxTemp != 31.0 || f.MinTemp != 24.0 {
			t.Errorf("Expected Min/Max 24.0/31.0, got %f/%f", f.MinTemp, f.MaxTemp)
		}
		if f.Conditions != "Partly cloudy" {
			t.Errorf("Expected forecast condition 'Partly cloudy', got '%s'", f.Conditions)
		}
	}
}

// writeDailyJSON encodes a weather response whose "time", "temperature_2m_max",
// "temperature_2m_min" and "weather_code" arrays can have independent lengths so
// tests can reproduce malformed/partial upstream payloads.
func writeDailyJSON(w http.ResponseWriter, times []string, max, min []float64, codes []int) {
	resp := WeatherResponse{}
	resp.Current.Temperature2m = 28.2
	resp.Current.RelativeHumidity = 91
	resp.Current.WeatherCode = 95
	resp.Daily.Time = times
	resp.Daily.Temperature2mMax = max
	resp.Daily.Temperature2mMin = min
	resp.Daily.WeatherCode = codes
	_ = json.NewEncoder(w).Encode(resp)
}

// buildDailySeries produces n daily entries starting at 2026-05-30.
func buildDailySeries(n int) ([]string, []float64, []float64, []int) {
	times := make([]string, 0, n)
	max := make([]float64, 0, n)
	min := make([]float64, 0, n)
	codes := make([]int, 0, n)
	for i := 0; i < n; i++ {
		day := 30 + i
		times = append(times, fmt.Sprintf("2026-05-%02d", day))
		max = append(max, 32.0-float64(i))
		min = append(min, 25.0-float64(i))
		codes = append(codes, 3)
	}
	return times, max, min, codes
}

// TestFetchWeatherForecastCountMatchesRequestedDays catches the off-by-one bug
// where the API's forecast_days=N (which already includes today) was requested
// directly, producing N-1 upcoming days and zero days for --forecast 1.
func TestFetchWeatherForecastCountMatchesRequestedDays(t *testing.T) {
	for _, days := range []int{1, 2, 3, 7} {
		t.Run(fmt.Sprintf("days=%d", days), func(t *testing.T) {
			var gotForecastDays string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				gotForecastDays = r.URL.Query().Get("forecast_days")

				// Emulate Open-Meteo: forecast_days=N returns N entries, today included.
				times, max, min, codes := buildDailySeries(days + 1)
				writeDailyJSON(w, times, max, min, codes)
			}))
			defer server.Close()

			oldBaseURL := weatherBaseURL
			weatherBaseURL = server.URL
			defer func() { weatherBaseURL = oldBaseURL }()

			data, err := FetchWeather(21.0245, 105.8412, "celsius", days)
			if err != nil {
				t.Fatalf("FetchWeather() failed: %v", err)
			}

			if want := strconv.Itoa(days + 1); gotForecastDays != want {
				t.Errorf("Expected forecast_days=%s (today + %d upcoming), got %q", want, days, gotForecastDays)
			}
			if len(data.Forecast) != days {
				t.Fatalf("Expected exactly %d forecast rows, got %d", days, len(data.Forecast))
			}
			if data.Forecast[0].Date == "2026-05-30" {
				t.Errorf("Forecast must not include today (2026-05-30), got first row %q", data.Forecast[0].Date)
			}
		})
	}
}

// TestFetchWeatherHandlesUnequalArrayLength catches the index-out-of-range panic
// when the parallel daily arrays have mismatched lengths.
func TestFetchWeatherHandlesUnequalArrayLength(t *testing.T) {
	tests := []struct {
		name      string
		days      int
		times     []string
		max       []float64
		min       []float64
		codes     []int
		wantCount int
	}{
		{
			name:      "weather_code shorter than time",
			days:      1,
			times:     []string{"2026-05-30", "2026-05-31"},
			max:       []float64{32.0, 31.0},
			min:       []float64{25.0, 24.0},
			codes:     []int{3},
			wantCount: 0,
		},
		{
			name:      "temperature arrays shorter than time",
			days:      1,
			times:     []string{"2026-05-30", "2026-05-31", "2026-06-01"},
			max:       []float64{32.0},
			min:       []float64{25.0},
			codes:     []int{3, 0, 61},
			wantCount: 0,
		},
		{
			name:      "only leading entries are complete",
			days:      1,
			times:     []string{"2026-05-30", "2026-05-31", "2026-06-01"},
			max:       []float64{32.0, 31.0, 30.0},
			min:       []float64{25.0, 24.0, 23.0},
			codes:     []int{3, 61},
			wantCount: 1,
		},
		{
			name:      "all arrays empty",
			days:      3,
			times:     nil,
			max:       nil,
			min:       nil,
			codes:     nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeDailyJSON(w, tt.times, tt.max, tt.min, tt.codes)
			}))
			defer server.Close()

			oldBaseURL := weatherBaseURL
			weatherBaseURL = server.URL
			defer func() { weatherBaseURL = oldBaseURL }()

			// A panic here fails the test naturally; the guard must prevent it.
			data, err := FetchWeather(21.0245, 105.8412, "celsius", tt.days)
			if err != nil {
				t.Fatalf("FetchWeather() failed: %v", err)
			}

			if len(data.Forecast) != tt.wantCount {
				t.Fatalf("Expected %d forecast rows, got %d", len(data.Forecast), tt.wantCount)
			}
			for i, f := range data.Forecast {
				if f.Date == "" {
					t.Errorf("Forecast row %d has empty date", i)
				}
			}
		})
	}
}

func TestFetchWeatherZeroDaysReturnsNoForecast(t *testing.T) {
	var sawForecastDays bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, ok := r.URL.Query()["forecast_days"]; ok {
			sawForecastDays = true
		}
		times, max, min, codes := buildDailySeries(4)
		writeDailyJSON(w, times, max, min, codes)
	}))
	defer server.Close()

	oldBaseURL := weatherBaseURL
	weatherBaseURL = server.URL
	defer func() { weatherBaseURL = oldBaseURL }()

	data, err := FetchWeather(21.0245, 105.8412, "celsius", 0)
	if err != nil {
		t.Fatalf("FetchWeather() failed: %v", err)
	}

	if len(data.Forecast) != 0 {
		t.Errorf("Expected no forecast rows for days=0, got %d", len(data.Forecast))
	}
	if sawForecastDays {
		t.Error("Expected no forecast_days parameter when days=0")
	}
	// Current conditions must still be present.
	if data.Conditions != "Thunderstorm" {
		t.Errorf("Expected current conditions 'Thunderstorm', got '%s'", data.Conditions)
	}
}

// TestFetchWeatherClampsDays ensures out-of-range input cannot produce a request
// the API would reject (and cannot inflate the forecast row count).
func TestFetchWeatherClampsDays(t *testing.T) {
	for _, tt := range []struct {
		name          string
		days          int
		wantRequested string
	}{
		{name: "negative", days: -3, wantRequested: ""},
		{name: "above max", days: 99, wantRequested: strconv.Itoa(maxForecastDays + 1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var gotForecastDays string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				gotForecastDays = r.URL.Query().Get("forecast_days")
				times, max, min, codes := buildDailySeries(maxForecastDays + 1)
				writeDailyJSON(w, times, max, min, codes)
			}))
			defer server.Close()

			oldBaseURL := weatherBaseURL
			weatherBaseURL = server.URL
			defer func() { weatherBaseURL = oldBaseURL }()

			data, err := FetchWeather(21.0245, 105.8412, "celsius", tt.days)
			if err != nil {
				t.Fatalf("FetchWeather() failed: %v", err)
			}

			if gotForecastDays != tt.wantRequested {
				t.Errorf("Expected forecast_days=%q, got %q", tt.wantRequested, gotForecastDays)
			}
			wantRows := 0
			if tt.days > maxForecastDays {
				wantRows = maxForecastDays
			}
			if len(data.Forecast) != wantRows {
				t.Errorf("Expected %d forecast rows, got %d", wantRows, len(data.Forecast))
			}
		})
	}
}
