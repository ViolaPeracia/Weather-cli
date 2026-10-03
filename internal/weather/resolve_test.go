package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// fakeCache is an in-memory CacheStore so Resolve can be tested without
// touching the filesystem or internal/config.
type fakeCache struct {
	entries map[string]WeatherData
	loads   int
	saves   int
}

func newFakeCache() *fakeCache {
	return &fakeCache{entries: map[string]WeatherData{}}
}

func cacheKey(lat, lon float64, unit string, days int) string {
	return fmt.Sprintf("%.4f|%.4f|%s|%d", lat, lon, unit, days)
}

func (c *fakeCache) LoadCache(lat, lon float64, unit string, days int) (WeatherData, bool) {
	c.loads++
	data, ok := c.entries[cacheKey(lat, lon, unit, days)]
	return data, ok
}

func (c *fakeCache) SaveCache(lat, lon float64, locName, unit string, days int, data WeatherData) error {
	c.saves++
	c.entries[cacheKey(lat, lon, unit, days)] = data
	return nil
}

// withStubAPIs points geocodeBaseURL and weatherBaseURL at httptest servers for
// the duration of the test.
func withStubAPIs(t *testing.T, geoHandler, weatherHandler http.HandlerFunc) {
	t.Helper()

	geoServer := httptest.NewServer(geoHandler)
	t.Cleanup(geoServer.Close)
	weatherServer := httptest.NewServer(weatherHandler)
	t.Cleanup(weatherServer.Close)

	oldGeo, oldWeather := geocodeBaseURL, weatherBaseURL
	geocodeBaseURL, weatherBaseURL = geoServer.URL, weatherServer.URL
	t.Cleanup(func() { geocodeBaseURL, weatherBaseURL = oldGeo, oldWeather })
}

func stubGeocodeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func stubWeatherHandler(requests *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if requests != nil {
			*requests++
		}
		// Emulate Open-Meteo: forecast_days=N returns N entries, today included.
		entries := 4
		if raw := r.URL.Query().Get("forecast_days"); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n > 0 {
				entries = n
			}
		}
		times, max, min, codes := buildDailySeries(entries)
		writeDailyJSON(w, times, max, min, codes)
	}
}

// TestResolveGeocodeThenCache checks the full geocode+fetch+cache path: the
// first call must hit the network (FromCache=false) and populate the cache, the
// second call with the same days must be served from cache (FromCache=true).
func TestResolveGeocodeThenCache(t *testing.T) {
	var weatherRequests int
	withStubAPIs(t, stubGeocodeHandler(), stubWeatherHandler(&weatherRequests))

	cache := newFakeCache()

	first, err := Resolve("Hanoi", "celsius", 2, false, cache)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}
	if first.Name != "Hanoi, Vietnam" {
		t.Errorf("Expected name 'Hanoi, Vietnam', got '%s'", first.Name)
	}
	if first.FromCache {
		t.Error("Expected FromCache=false on the first (network) call")
	}
	if first.Data.Conditions != "Thunderstorm" {
		t.Errorf("Expected current conditions 'Thunderstorm', got '%s'", first.Data.Conditions)
	}
	if len(first.Data.Forecast) != 2 {
		t.Errorf("Expected 2 forecast rows for days=2, got %d", len(first.Data.Forecast))
	}
	if weatherRequests != 1 {
		t.Errorf("Expected exactly 1 weather API request, got %d", weatherRequests)
	}
	if cache.saves != 1 {
		t.Errorf("Expected the result to be cached once, got %d saves", cache.saves)
	}

	second, err := Resolve("Hanoi", "celsius", 2, false, cache)
	if err != nil {
		t.Fatalf("Resolve() (cached) failed: %v", err)
	}
	if !second.FromCache {
		t.Error("Expected FromCache=true on the second call with matching days")
	}
	if weatherRequests != 1 {
		t.Errorf("Expected no additional weather API request on cache hit, got %d total", weatherRequests)
	}
	if second.Data.Temperature != first.Data.Temperature {
		t.Errorf("Expected cached temperature %f, got %f", first.Data.Temperature, second.Data.Temperature)
	}
}

// TestResolveCacheKeyIncludesDays guards the cache key: a different --forecast
// value must not reuse a cached response.
func TestResolveCacheKeyIncludesDays(t *testing.T) {
	var weatherRequests int
	withStubAPIs(t, stubGeocodeHandler(), stubWeatherHandler(&weatherRequests))

	cache := newFakeCache()

	if _, err := Resolve("Hanoi", "celsius", 1, false, cache); err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}
	if _, err := Resolve("Hanoi", "celsius", 3, false, cache); err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	if weatherRequests != 2 {
		t.Errorf("Expected a network fetch per distinct forecast length, got %d requests", weatherRequests)
	}

	result, err := Resolve("Hanoi", "celsius", 3, false, cache)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}
	if !result.FromCache {
		t.Error("Expected FromCache=true when repeating the same forecast length")
	}
	if weatherRequests != 2 {
		t.Errorf("Expected no new request on cache hit, got %d total", weatherRequests)
	}
}

// TestResolveForceBypassesCache ensures --force skips the cache read and
// re-populates it.
func TestResolveForceBypassesCache(t *testing.T) {
	var weatherRequests int
	withStubAPIs(t, stubGeocodeHandler(), stubWeatherHandler(&weatherRequests))

	cache := newFakeCache()

	if _, err := Resolve("Hanoi", "celsius", 0, false, cache); err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}
	forced, err := Resolve("Hanoi", "celsius", 0, true, cache)
	if err != nil {
		t.Fatalf("Resolve(force) failed: %v", err)
	}
	if forced.FromCache {
		t.Error("Expected FromCache=false when force is set")
	}
	if weatherRequests != 2 {
		t.Errorf("Expected 2 weather API requests, got %d", weatherRequests)
	}
	if cache.loads != 1 {
		t.Errorf("Expected the cache to be consulted only when force is false, got %d loads", cache.loads)
	}
}

// TestResolveWithoutCacheStore verifies a nil CacheStore disables caching
// without panicking.
func TestResolveWithoutCacheStore(t *testing.T) {
	var weatherRequests int
	withStubAPIs(t, stubGeocodeHandler(), stubWeatherHandler(&weatherRequests))

	first, err := Resolve("Hanoi", "celsius", 0, false, nil)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}
	if first.FromCache {
		t.Error("Expected FromCache=false without a cache store")
	}

	second, err := Resolve("Hanoi", "celsius", 0, false, nil)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}
	if second.FromCache {
		t.Error("Expected FromCache=false without a cache store")
	}
	if weatherRequests != 2 {
		t.Errorf("Expected 2 weather API requests (no caching), got %d", weatherRequests)
	}
}

// TestResolvePropagatesGeocodeError ensures a bad city name surfaces as an
// error rather than an empty result.
func TestResolvePropagatesGeocodeError(t *testing.T) {
	notFound := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GeocodeResponse{Results: nil})
	}
	withStubAPIs(t, notFound, stubWeatherHandler(nil))

	result, err := Resolve("NoSuchCity", "celsius", 0, false, newFakeCache())
	if err == nil {
		t.Fatal("Expected an error for an unknown city, got nil")
	}
	if result.Name != "" {
		t.Errorf("Expected an empty result on error, got name '%s'", result.Name)
	}
}

// TestErrCityNotFoundIsMatchable locks in the sentinel contract: callers (the
// TUI) must be able to tell "unknown city" from every other failure with
// errors.Is, even after Resolve has wrapped the error, while the message still
// names the city for display.
func TestErrCityNotFoundIsMatchable(t *testing.T) {
	notFound := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GeocodeResponse{Results: nil})
	}
	withStubAPIs(t, notFound, stubWeatherHandler(nil))

	_, err := Resolve("NoSuchCity", "celsius", 0, false, newFakeCache())
	if err == nil {
		t.Fatal("Expected an error for an unknown city, got nil")
	}
	if !errors.Is(err, ErrCityNotFound) {
		t.Errorf("Expected errors.Is(err, ErrCityNotFound) to be true, got %v", err)
	}
	// Still matchable through an extra wrap, as any outer layer may add one.
	if !errors.Is(fmt.Errorf("resolving city: %w", err), ErrCityNotFound) {
		t.Error("Expected the sentinel to survive an additional outer wrap")
	}
	if !strings.Contains(err.Error(), "NoSuchCity") {
		t.Errorf("Expected the error message to name the city, got %q", err.Error())
	}
}

// TestErrCityNotFoundNotUsedForOtherFailures guards the other direction: a
// transport failure must not be mistaken for an unknown city, or the TUI would
// silently revert instead of reporting the outage.
func TestErrCityNotFoundNotUsedForOtherFailures(t *testing.T) {
	broken := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	withStubAPIs(t, broken, stubWeatherHandler(nil))

	_, err := Resolve("Hanoi", "celsius", 0, false, newFakeCache())
	if err == nil {
		t.Fatal("Expected an error when the geocoding API fails, got nil")
	}
	if errors.Is(err, ErrCityNotFound) {
		t.Errorf("A geocoding HTTP failure must not match ErrCityNotFound, got %v", err)
	}
}
