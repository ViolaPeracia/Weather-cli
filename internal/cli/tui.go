package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ViolaPeracia/Weather-cli/internal/config"
	"github.com/ViolaPeracia/Weather-cli/internal/display"
	"github.com/ViolaPeracia/Weather-cli/internal/location"
	"github.com/ViolaPeracia/Weather-cli/internal/weather"
)

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
			// Resolve and fetch weather
			var lat, lon float64
			var resolvedName string
			var data weather.WeatherData
			var loadSuccess bool

			if city == "" {
				loc, err := location.DetectLocation()
				if err != nil {
					errorMessage = fmt.Sprintf("Failed to detect location: %v", err)
				} else {
					lat = loc.Lat
					lon = loc.Lon
					resolvedName = fmt.Sprintf("%s, %s", loc.City, loc.Country)
					loadSuccess = true
				}
			} else {
				var err error
				lat, lon, resolvedName, err = weather.GeocodeCity(city)
				if err != nil {
					errorMessage = fmt.Sprintf("City '%s' not found. Reverting...", city)
					city = cfg.DefaultCity // Revert to config default
				} else {
					loadSuccess = true
				}
			}

			if loadSuccess {
				// Attempt to load from cache
				var cacheHit bool
				if cacheEntry, ok := config.LoadCache(lat, lon, unit, forecastDays); ok {
					data = cacheEntry.WeatherData
					cacheHit = true
				}

				if !cacheHit {
					var err error
					data, err = weather.FetchWeather(lat, lon, unit, forecastDays)
					if err != nil {
						errorMessage = fmt.Sprintf("Failed to fetch weather: %v", err)
						loadSuccess = false
					} else {
						_ = config.SaveCache(lat, lon, resolvedName, unit, forecastDays, data)
					}
				}
			}

			if loadSuccess && errorMessage == "" {
				display.RenderWeather(stdout, resolvedName, data)
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
