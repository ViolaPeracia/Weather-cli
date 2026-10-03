package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ViolaPeracia/Weather-cli/internal/config"
	"github.com/ViolaPeracia/Weather-cli/internal/display"
	"github.com/ViolaPeracia/Weather-cli/internal/weather"
)

// configCache adapts the config package's package-level cache functions to the
// weather.CacheStore interface, mirroring the adapter used by cmd/weather, so
// internal/weather stays free of an import cycle back into internal/config.
type configCache struct{}

func (configCache) LoadCache(lat, lon float64, unit string, days int) (weather.WeatherData, bool) {
	entry, ok := config.LoadCache(lat, lon, unit, days)
	if !ok {
		return weather.WeatherData{}, false
	}
	return entry.WeatherData, true
}

func (configCache) SaveCache(lat, lon float64, locName, unit string, days int, data weather.WeatherData) error {
	return config.SaveCache(lat, lon, locName, unit, days, data)
}

// tuiErrorFor maps a weather.Resolve failure to the TUI banner message and
// reports whether the TUI must fall back to the configured default city.
//
// Only an unknown city is recoverable. IP-based location detection and every
// other failure (network, HTTP status, decode) leave the city untouched.
func tuiErrorFor(city string, resolveErr error) (message string, revertToDefault bool) {
	switch {
	case errors.Is(resolveErr, weather.ErrCityNotFound):
		return fmt.Sprintf("City '%s' not found. Reverting...", city), true
	case city == "":
		// Resolve already labels this failure "failed to detect location: ...";
		// unwrap once so the banner does not repeat that prefix.
		return fmt.Sprintf("Failed to detect location: %v", errors.Unwrap(resolveErr)), false
	default:
		return fmt.Sprintf("Failed to fetch weather: %v", resolveErr), false
	}
}

// StartTUI launches the interactive terminal user interface.
func StartTUI(cfg config.Config) {
	runTUI(os.Stdin, os.Stdout, cfg)
}

// runTUI is the testable implementation of the TUI loop, accepting custom input/output streams.
func runTUI(stdin io.Reader, stdout io.Writer, cfg config.Config) {
	scanner := bufio.NewScanner(stdin)
	city := cfg.DefaultCity
	unit := cfg.Unit
	if unit == "" {
		unit = "celsius"
	}
	forecastDays := 0
	showHelp := false
	var errorMessage string

	for {
		// 1. Clear terminal screen
		// \033[H moves cursor to top-left; \033[2J clears the screen
		_, _ = fmt.Fprint(stdout, "\033[H\033[2J")

		// 2. Render Header
		_, _ = fmt.Fprintln(stdout, "Weather CLI - Interactive Dashboard")
		_, _ = fmt.Fprintln(stdout, "====================================")

		if errorMessage != "" {
			_, _ = fmt.Fprintf(stdout, "\033[31m✖ Error: %s\033[0m\n\n", errorMessage)
			errorMessage = "" // clear error
		}

		if showHelp {
			renderHelpScreen(stdout)
		} else {
			// Resolve location (city name, or IP detection when blank), fetch
			// through the cache, then render - a single call shared with main.
			result, err := weather.Resolve(city, unit, forecastDays, false, configCache{})
			if err != nil {
				var revert bool
				errorMessage, revert = tuiErrorFor(city, err)
				if revert {
					city = cfg.DefaultCity // Revert to config default
				}
			} else {
				display.RenderWeather(stdout, result.Name, result.Data)
			}
		}

		// 3. Print navigation options
		_, _ = fmt.Fprintln(stdout, "Commands: [c] Change City | [u] Toggle Unit | [f] Toggle Forecast | [h] Toggle Help | [q] Quit")
		_, _ = fmt.Fprint(stdout, "Enter command: ")

		if !scanner.Scan() {
			break // EOF
		}

		cmd := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if cmd == "" {
			continue
		}

		switch cmd {
		case "q", "quit":
			_, _ = fmt.Fprintln(stdout, "Goodbye!")
			return
		case "h", "help":
			showHelp = !showHelp
		case "u", "unit":
			if unit == "celsius" {
				unit = "fahrenheit"
			} else {
				unit = "celsius"
			}
		case "f", "forecast":
			if forecastDays == 0 {
				forecastDays = 3
			} else {
				forecastDays = 0
			}
		case "c", "city":
			_, _ = fmt.Fprint(stdout, "Enter city name (leave empty for IP detection): ")
			if scanner.Scan() {
				city = strings.TrimSpace(scanner.Text())
			}
		default:
			errorMessage = fmt.Sprintf("Unknown command '%s'. Press 'h' for help.", cmd)
		}
	}
}

func renderHelpScreen(w io.Writer) {
	_, _ = fmt.Fprintln(w, "╭────────────────────────────────────────────────────╮")
	_, _ = fmt.Fprintln(w, "│  Weather CLI - Help & Guidance                     │")
	_, _ = fmt.Fprintln(w, "├────────────────────────────────────────────────────┤")
	_, _ = fmt.Fprintln(w, "│  Keyboard Commands:                                │")
	_, _ = fmt.Fprintln(w, "│    c - Change target city (prompts for input)      │")
	_, _ = fmt.Fprintln(w, "│    u - Toggle temperature unit (Celsius/Fahrenheit)│")
	_, _ = fmt.Fprintln(w, "│    f - Toggle 3-day forecast display on/off        │")
	_, _ = fmt.Fprintln(w, "│    h - Toggle this help page                       │")
	_, _ = fmt.Fprintln(w, "│    q - Quit the interactive TUI                    │")
	_, _ = fmt.Fprintln(w, "├────────────────────────────────────────────────────┤")
	_, _ = fmt.Fprintln(w, "│  Storage Directory:                                │")
	_, _ = fmt.Fprintln(w, "│    ~/.weather-cli/                                 │")
	_, _ = fmt.Fprintln(w, "│      - config.json (Saved default city & unit)     │")
	_, _ = fmt.Fprintln(w, "│      - cache.json  (Weather queries local cache)   │")
	_, _ = fmt.Fprintln(w, "╰────────────────────────────────────────────────────╯")
	_, _ = fmt.Fprintln(w, "")
}
