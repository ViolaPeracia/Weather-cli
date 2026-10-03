package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ViolaPeracia/Weather-cli/internal/cli"
	"github.com/ViolaPeracia/Weather-cli/internal/config"
	"github.com/ViolaPeracia/Weather-cli/internal/display"
	"github.com/ViolaPeracia/Weather-cli/internal/weather"
)

// version is overridden at build time via -ldflags "-X main.version=1.2.0".
var version = "dev"

// maxForecastDays mirrors the help text: --forecast accepts 1-7, with 0 meaning
// "current conditions only".
const maxForecastDays = 7

// configCache adapts the config package's package-level cache functions to the
// weather.CacheStore interface, so internal/weather stays free of an import
// cycle back into internal/config.
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

func main() {
	// Parse CLI flags
	cityFlag := flag.String("city", "", "Specify a city to fetch the weather for")
	unitFlag := flag.String("unit", "", "Unit of temperature: 'celsius' or 'fahrenheit'")
	forecastFlag := flag.Int("forecast", 0, "Number of upcoming days to forecast (1-7)")
	saveConfigFlag := flag.Bool("save-config", false, "Save current settings as default")
	versionFlag := flag.Bool("version", false, "Print the version of Weather CLI")
	configShowFlag := flag.Bool("config-show", false, "Show current default configuration settings")
	configSetFlag := flag.String("config-set", "", "Set a default preference key=value (e.g. city=Hanoi, unit=celsius)")
	forceFlag := flag.Bool("force", false, "Force fetch weather data, bypassing local cache")
	tuiFlag := flag.Bool("tui", false, "Start Weather CLI in interactive TUI mode")

	flag.Parse()

	if *versionFlag {
		fmt.Println("Weather CLI v" + version)
		os.Exit(0)
	}

	// Validate flag values at the boundary so bad input never reaches the
	// lower layers (which would silently clamp it).
	if err := validateForecastDays(*forecastFlag); err != nil {
		display.PrintError(err)
		os.Exit(1)
	}

	// 1. Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		display.PrintError(fmt.Errorf("failed to load config: %v", err))
		// Non-fatal, continue with defaults
	}

	if *tuiFlag {
		cli.StartTUI(cfg)
		os.Exit(0)
	}

	if *configShowFlag {
		fmt.Println("Current Default Configuration Settings:")
		fmt.Printf("  Default City: %s\n", cfg.DefaultCity)
		fmt.Printf("  Default Unit: %s\n", cfg.Unit)
		os.Exit(0)
	}

	if *configSetFlag != "" {
		key, value, err := parseConfigSet(*configSetFlag)
		if err != nil {
			display.PrintError(err)
			os.Exit(1)
		}

		switch key {
		case "city":
			cfg.DefaultCity = value
		case "unit":
			if value != "celsius" && value != "fahrenheit" {
				display.PrintError(fmt.Errorf("invalid unit '%s'. Must be 'celsius' or 'fahrenheit'", value))
				os.Exit(1)
			}
			cfg.Unit = value
		default:
			display.PrintError(fmt.Errorf("unknown config key '%s'. Supported keys: city, unit", key))
			os.Exit(1)
		}

		if err := config.SaveConfig(cfg); err != nil {
			display.PrintError(fmt.Errorf("failed to save config: %v", err))
			os.Exit(1)
		}
		fmt.Printf("Successfully updated configuration: %s=%s\n", key, value)
		os.Exit(0)
	}

	// 2. Cascade logic for Unit
	finalUnit := "celsius"
	if *unitFlag != "" {
		finalUnit = *unitFlag
	} else if cfg.Unit != "" {
		finalUnit = cfg.Unit
	}

	if finalUnit != "celsius" && finalUnit != "fahrenheit" {
		display.PrintError(fmt.Errorf("invalid unit '%s'. Must be 'celsius' or 'fahrenheit'", finalUnit))
		os.Exit(1)
	}

	// 3. Cascade logic for City
	finalCity := *cityFlag
	if finalCity == "" {
		finalCity = cfg.DefaultCity
	}

	// 4. Save Config if requested
	if *saveConfigFlag {
		newCfg := config.Config{
			DefaultCity: finalCity,
			Unit:        finalUnit,
		}
		if err := config.SaveConfig(newCfg); err != nil {
			display.PrintError(fmt.Errorf("failed to save config: %v", err))
		} else {
			fmt.Println("Config saved successfully.")
		}
	}

	// 5. Resolve location, then fetch weather (via cache when possible)
	result, err := weather.Resolve(finalCity, finalUnit, *forecastFlag, *forceFlag, configCache{})
	if err != nil {
		display.PrintError(err)
		os.Exit(1)
	}

	// 6. Render beautiful ASCII widget
	display.RenderWeather(os.Stdout, result.Name, result.Data)
}

// validateForecastDays rejects --forecast values outside 0-7, naming both the
// offending value and the accepted range.
func validateForecastDays(days int) error {
	if days < 0 || days > maxForecastDays {
		return fmt.Errorf("invalid forecast '%d'. Must be between 0 and %d (0 = current conditions only)", days, maxForecastDays)
	}
	return nil
}

// parseConfigSet splits a "key=value" argument and rejects malformed input or
// an empty value, which would otherwise be swallowed silently.
func parseConfigSet(arg string) (key string, value string, err error) {
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid config-set format. Expected key=value (e.g. city=Hanoi)")
	}
	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])

	if key == "" {
		return "", "", fmt.Errorf("invalid config-set: missing key. Expected key=value (e.g. city=Hanoi)")
	}
	if value == "" {
		return "", "", fmt.Errorf("invalid config-set: value for '%s' cannot be empty", key)
	}
	return key, value, nil
}
