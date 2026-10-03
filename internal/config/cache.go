package config

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/ViolaPeracia/Weather-cli/internal/weather"
)

// CacheEntry represents the structure of the cached weather data.
type CacheEntry struct {
	Lat         float64             `json:"lat"`
	Lon         float64             `json:"lon"`
	Location    string              `json:"location"`
	Unit        string              `json:"unit"`
	Days        int                 `json:"days"`
	WeatherData weather.WeatherData `json:"weather_data"`
	CachedAt    time.Time           `json:"cached_at"`
}

// getCachePath resolves the path to ~/.weather-cli/cache.json
func getCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".weather-cli", "cache.json"), nil
}

// SaveCache writes weather data to cache.json.
func SaveCache(lat, lon float64, locName string, unit string, days int, data weather.WeatherData) error {
	path, err := getCachePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	entry := CacheEntry{
		Lat:         lat,
		Lon:         lon,
		Location:    locName,
		Unit:        unit,
		Days:        days,
		WeatherData: data,
		CachedAt:    time.Now(),
	}

	raw, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode cache JSON: %w", err)
	}

	if err := os.WriteFile(path, raw, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// LoadCache reads cached data. Returns entry and true if cache exists, is valid for same coords, unit & forecast days, and is under 10 minutes old.
func LoadCache(lat, lon float64, unit string, days int) (CacheEntry, bool) {
	path, err := getCachePath()
	if err != nil {
		return CacheEntry{}, false
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return CacheEntry{}, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return CacheEntry{}, false
	}

	// 1. Check age (TTL = 10 minutes)
	if time.Since(entry.CachedAt) > 10*time.Minute {
		return CacheEntry{}, false
	}

	// 2. Check temperature unit matches
	if entry.Unit != unit {
		return CacheEntry{}, false
	}

	// 3. Check forecast days requested matches
	if entry.Days != days {
		return CacheEntry{}, false
	}

	// 4. Check coordinates match (within small delta, e.g. 0.01 degrees)
	const delta = 0.01
	if math.Abs(entry.Lat-lat) > delta || math.Abs(entry.Lon-lon) > delta {
		return CacheEntry{}, false
	}

	return entry, true
}
