package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var geocodeBaseURL = "https://geocoding-api.open-meteo.com"
var weatherBaseURL = "https://api.open-meteo.com"

// ErrCityNotFound reports that the geocoding API returned no match for the
// requested city. Callers match it with errors.Is to tell an unknown city -
// which may fall back to a configured default - apart from a transport or API
// failure, whose cause is unrelated to the city name.
var ErrCityNotFound = errors.New("city not found")

// maxForecastDays caps the number of upcoming forecast days requested from
// the API (Open-Meteo itself accepts up to 16).
const maxForecastDays = 7

// DailyForecast represents the forecast for a single day.
type DailyForecast struct {
	Date       string
	MinTemp    float64
	MaxTemp    float64
	Conditions string
	Icon       string
}

// WeatherData represents current and future weather conditions.
type WeatherData struct {
	Temperature float64
	Humidity    int
	Conditions  string
	Icon        string
	Unit        string
	Forecast    []DailyForecast
}

// GeocodeResponse represents the response from Open-Meteo Geocoding API.
type GeocodeResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Name      string  `json:"name"`
		Country   string  `json:"country"`
	} `json:"results"`
}

// WeatherResponse represents the response from Open-Meteo Weather API.
type WeatherResponse struct {
	Current struct {
		Temperature2m    float64 `json:"temperature_2m"`
		RelativeHumidity int     `json:"relative_humidity_2m"`
		WeatherCode      int     `json:"weather_code"`
	} `json:"current"`
	Daily struct {
		Time             []string  `json:"time"`
		Temperature2mMax []float64 `json:"temperature_2m_max"`
		Temperature2mMin []float64 `json:"temperature_2m_min"`
		WeatherCode      []int     `json:"weather_code"`
	} `json:"daily"`
}

// GeocodeCity converts a city name into coordinates using Open-Meteo Geocoding API.
func GeocodeCity(city string) (float64, float64, string, error) {
	apiURL := fmt.Sprintf("%s/v1/search?name=%s&count=1&language=en&format=json", geocodeBaseURL, url.QueryEscape(city))

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to fetch geocoding data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, "", fmt.Errorf("geocoding API returned status: %d", resp.StatusCode)
	}

	var geoResp GeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return 0, 0, "", fmt.Errorf("failed to decode geocoding response: %w", err)
	}

	if len(geoResp.Results) == 0 {
		return 0, 0, "", fmt.Errorf("%w: %s", ErrCityNotFound, city)
	}

	result := geoResp.Results[0]
	locationName := fmt.Sprintf("%s, %s", result.Name, result.Country)
	return result.Latitude, result.Longitude, locationName, nil
}

// FetchWeather fetches weather data for a given latitude and longitude using Open-Meteo API.
// unit should be "celsius" or "fahrenheit". days is the number of UPCOMING forecast days (0-7);
// today's conditions are returned via the current fields, so days=0 yields no forecast rows.
// Values outside 0-7 are clamped defensively.
func FetchWeather(lat, lon float64, unit string, days int) (WeatherData, error) {
	tempUnit := ""
	unitSymbol := "C"
	if unit == "fahrenheit" {
		tempUnit = "&temperature_unit=fahrenheit"
		unitSymbol = "F"
	}

	// The forecast panel already shows today's current conditions, so it must
	// contain only upcoming days. Open-Meteo's forecast_days=N includes today,
	// therefore request N+1 entries and drop the first one.
	if days < 0 {
		days = 0
	}
	if days > maxForecastDays {
		days = maxForecastDays
	}

	forecastQuery := ""
	if days > 0 {
		forecastQuery = fmt.Sprintf("&daily=temperature_2m_max,temperature_2m_min,weather_code&forecast_days=%d", days+1)
	}

	apiURL := fmt.Sprintf("%s/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,relative_humidity_2m,weather_code%s%s", weatherBaseURL, lat, lon, tempUnit, forecastQuery)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return WeatherData{}, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherData{}, fmt.Errorf("weather API returned status: %d", resp.StatusCode)
	}

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return WeatherData{}, fmt.Errorf("failed to decode weather response: %w", err)
	}

	current := weatherResp.Current
	conditions, icon := decodeWeatherCode(current.WeatherCode)

	var forecasts []DailyForecast
	if days > 0 {
		// The API's parallel daily arrays can come back with unequal lengths
		// (partial responses, upstream/proxy glitches). Only trust the prefix
		// that exists in every array.
		safeLen := len(weatherResp.Daily.Time)
		for _, series := range []int{
			len(weatherResp.Daily.Temperature2mMax),
			len(weatherResp.Daily.Temperature2mMin),
			len(weatherResp.Daily.WeatherCode),
		} {
			if series < safeLen {
				safeLen = series
			}
		}

		// Index 0 is today, which is already rendered as the current panel.
		for i := 1; i < safeLen; i++ {
			fCond, fIcon := decodeWeatherCode(weatherResp.Daily.WeatherCode[i])
			forecasts = append(forecasts, DailyForecast{
				Date:       weatherResp.Daily.Time[i],
				MinTemp:    weatherResp.Daily.Temperature2mMin[i],
				MaxTemp:    weatherResp.Daily.Temperature2mMax[i],
				Conditions: fCond,
				Icon:       fIcon,
			})
		}
	}

	return WeatherData{
		Temperature: current.Temperature2m,
		Humidity:    current.RelativeHumidity,
		Conditions:  conditions,
		Icon:        icon,
		Unit:        unitSymbol,
		Forecast:    forecasts,
	}, nil
}

// decodeWeatherCode converts WMO weather codes to strings and Nerd Font icons.
func decodeWeatherCode(code int) (string, string) {
	switch {
	case code == 0:
		return "Clear sky", "󰖙"
	case code == 1 || code == 2 || code == 3:
		return "Partly cloudy", "󰖕"
	case code == 45 || code == 48:
		return "Fog", "󰖑"
	case code >= 51 && code <= 55:
		return "Drizzle", "󰖗"
	case code >= 56 && code <= 57:
		return "Freezing Drizzle", "󰖗"
	case code >= 61 && code <= 65:
		return "Rain", "󰖖"
	case code >= 66 && code <= 67:
		return "Freezing Rain", "󰖖"
	case code >= 71 && code <= 75:
		return "Snow fall", "󰖘"
	case code == 77:
		return "Snow grains", "󰖘"
	case code >= 80 && code <= 82:
		return "Rain showers", "󰖗"
	case code >= 85 && code <= 86:
		return "Snow showers", "󰖘"
	case code >= 95 && code <= 99:
		return "Thunderstorm", "󰖓"
	default:
		return "Unknown", "󰖕"
	}
}
