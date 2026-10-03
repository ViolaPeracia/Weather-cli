package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ViolaPeracia/Weather-cli/internal/cli"
	"github.com/ViolaPeracia/Weather-cli/internal/config"
	"github.com/ViolaPeracia/Weather-cli/internal/display"
	"github.com/ViolaPeracia/Weather-cli/internal/location"
	"github.com/ViolaPeracia/Weather-cli/internal/weather"
)

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
		fmt.Println("Weather CLI v1.0.0")
		os.Exit(0)
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
		parts := strings.SplitN(*configSetFlag, "=", 2)
		if len(parts) != 2 {
			display.PrintError(fmt.Errorf("invalid config-set format. Expected key=value (e.g. city=Hanoi)"))
			os.Exit(1)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

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

	var targetLat, targetLon float64
	var targetName string

	if finalCity == "" {
		// IP Detection
		loc, err := location.DetectLocation()
		if err != nil {
			display.PrintError(err)
			os.Exit(1)
		}
		targetLat = loc.Lat
		targetLon = loc.Lon
		targetName = fmt.Sprintf("%s, %s", loc.City, loc.Country)
	} else {
		// Geocoding
		lat, lon, name, err := weather.GeocodeCity(finalCity)
		if err != nil {
			display.PrintError(err)
			os.Exit(1)
		}
		targetLat = lat
		targetLon = lon
		targetName = name
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

	// 5. Fetch weather data (check cache first)
	var data weather.WeatherData
	var cacheHit bool

	if !*forceFlag {
		if cacheEntry, ok := config.LoadCache(targetLat, targetLon, finalUnit, *forecastFlag); ok {
			data = cacheEntry.WeatherData
			cacheHit = true
		}
	}

	if !cacheHit {
		var err error
		data, err = weather.FetchWeather(targetLat, targetLon, finalUnit, *forecastFlag)
		if err != nil {
			display.PrintError(err)
			os.Exit(1)
		}
		// Save to cache (non-fatal if it fails)
		_ = config.SaveCache(targetLat, targetLon, targetName, finalUnit, *forecastFlag, data)
	}

	// 6. Render beautiful ASCII widget
	display.RenderWeather(os.Stdout, targetName, data)
}
