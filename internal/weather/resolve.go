package weather

import (
	"fmt"
	"strings"

	"github.com/ViolaPeracia/Weather-cli/internal/location"
)

// CacheStore is the minimal cache contract Resolve needs.
//
// It exists to break an import cycle: internal/config imports internal/weather
// (CacheEntry embeds WeatherData), so internal/weather cannot import
// internal/config. Callers (main, TUI) pass their own implementation instead.
// A nil CacheStore disables caching entirely.
type CacheStore interface {
	LoadCache(lat, lon float64, unit string, days int) (WeatherData, bool)
	SaveCache(lat, lon float64, locName, unit string, days int, data WeatherData) error
}

// ResolveResult is the outcome of a single resolve+fetch cycle.
type ResolveResult struct {
	Name      string
	Data      WeatherData
	FromCache bool
}

// Resolve turns a location request into weather data.
//
// When city is empty the user's location is detected from their IP address,
// otherwise the city is geocoded. The result is served from cache when a cache
// store is supplied, days matches and force is false; otherwise it is fetched
// from the API and written back to the cache. A cache write failure is
// non-fatal: the freshly fetched data is still returned.
func Resolve(city string, unit string, days int, force bool, cache CacheStore) (ResolveResult, error) {
	lat, lon, name, err := resolveLocation(city)
	if err != nil {
		return ResolveResult{}, err
	}

	if cache != nil && !force {
		if data, ok := cache.LoadCache(lat, lon, unit, days); ok {
			return ResolveResult{Name: name, Data: data, FromCache: true}, nil
		}
	}

	data, err := FetchWeather(lat, lon, unit, days)
	if err != nil {
		return ResolveResult{}, err
	}

	if cache != nil {
		_ = cache.SaveCache(lat, lon, name, unit, days, data)
	}

	return ResolveResult{Name: name, Data: data, FromCache: false}, nil
}

// resolveLocation returns coordinates and a display name for the request.
// An empty city falls back to IP-based location detection.
func resolveLocation(city string) (lat float64, lon float64, name string, err error) {
	if strings.TrimSpace(city) == "" {
		loc, err := location.DetectLocation()
		if err != nil {
			return 0, 0, "", fmt.Errorf("failed to detect location: %w", err)
		}
		return loc.Lat, loc.Lon, fmt.Sprintf("%s, %s", loc.City, loc.Country), nil
	}

	lat, lon, name, err = GeocodeCity(city)
	if err != nil {
		return 0, 0, "", err
	}
	return lat, lon, name, nil
}
