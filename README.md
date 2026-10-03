# Weather CLI

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Weather CLI is a small, cross-platform Go command line tool for checking current weather from the terminal. It can use your IP address to detect a location automatically, or it can fetch weather for a city you provide.

The project uses only the Go standard library and public JSON APIs.

## Features

- Detects your current city and coordinates with IP geolocation (`https://ipwho.is/`, HTTPS) when no city is provided.
- Looks up city coordinates with the Open-Meteo geocoding API.
- Fetches current temperature, humidity, and weather conditions from Open-Meteo.
- Supports Celsius and Fahrenheit output.
- Supports optional multi-day forecasts (1 to 7 days).
- Renders a formatted terminal weather panel with ANSI colors, honoring `NO_COLOR` for plain-text output.
- Interactive TUI dashboard with live settings toggles (`--tui`).
- Caches the last weather response locally for 10 minutes to avoid repeat API calls; bypass with `--force`.
- Saves default city and unit preferences in a local JSON config file.

## Requirements

- Go 1.22 or newer
- Network access for location, geocoding, and weather API calls

No API key is required: Open-Meteo and ipwho.is are both free and unauthenticated.

## Build

```bash
go build -o weather-cli ./cmd/weather
```

To inject a version string that `--version` will report:

```bash
go build -ldflags "-X main.version=v1.2.0" -o weather-cli ./cmd/weather
```

On Windows PowerShell:

```powershell
go build -o weather-cli.exe .\cmd\weather
```

## Usage

Run with automatic IP-based location detection:

```bash
./weather-cli
```

Run for a specific city:

```bash
./weather-cli --city "London"
```

Use Fahrenheit:

```bash
./weather-cli --city "New York" --unit fahrenheit
```

Show a forecast:

```bash
./weather-cli --city "Tokyo" --forecast 3
```

Save your current options as defaults:

```bash
./weather-cli --city "Paris" --unit celsius --save-config
```

Show the version:

```bash
./weather-cli --version
```

Launch the interactive TUI dashboard:

```bash
./weather-cli --tui
```

Bypass the local cache and always hit the network:

```bash
./weather-cli --city "Berlin" --force
```

Inspect or edit saved defaults:

```bash
./weather-cli --config-show
./weather-cli --config-set city=Hanoi
```

## Flags

| Flag | Description |
| --- | --- |
| `--city` | City name to fetch weather for. If omitted, the CLI uses config defaults or IP geolocation. |
| `--unit` | Temperature unit. Accepted values: `celsius`, `fahrenheit`. |
| `--forecast` | Number of forecast days to request, from 1 to 7. |
| `--save-config` | Save the selected city and unit to the local config file. |
| `--force` | Skip the local cache and always fetch fresh weather data. |
| `--tui` | Start the interactive terminal dashboard instead of printing a single report. |
| `--config-show` | Print the current default city and unit, then exit. |
| `--config-set` | Set a default preference as `key=value`. Supported keys: `city`, `unit`. Then exit. |
| `--version` | Print the CLI version and exit. |

## Configuration

Saved preferences are stored at:

```text
~/.weather-cli/config.json
```

Example:

```json
{
  "default_city": "Paris",
  "unit": "celsius"
}
```

Configuration is optional. Without a saved city, Weather CLI attempts IP-based location detection.

The last weather response is cached at `~/.weather-cli/cache.json` for 10 minutes, keyed by coordinates, unit and forecast length. Use `--force` to skip it.

To disable all ANSI colors (for pipes, logs or pagers), set the `NO_COLOR` environment variable to any non-empty value:

```bash
NO_COLOR=1 ./weather-cli --city "Oslo"
```

See [Configuration](docs/CONFIGURATION.md) for the full details.

## Development

Run all tests, matching CI:

```bash
go test -race ./... -count=1
```

Check formatting (must print nothing):

```bash
gofmt -l .
```

Continuous integration is defined in `.github/workflows/ci.yml`. It runs on every push and pull request across Ubuntu, Windows and macOS, and enforces `go build ./...`, `go vet ./...`, `gofmt -l .` and `go test -race ./... -count=1`.

## Documentation

- [Getting Started](docs/GETTING-STARTED.md)
- [Configuration](docs/CONFIGURATION.md)
- [Development Roadmap](docs/ROADMAP.md)
- [Features Tracker](docs/FEATURES.md)
- [Architecture & Design](docs/ARCHITECTURE.md)
- [Testing](docs/TESTING.md)
- [Contributing](docs/CONTRIBUTING.md)
- [Agent Directives](AGENTS.md)
- [AI Policy Rules](AI_POLICY_RULES.md)

## License

This project is licensed under the Apache License, Version 2.0 (Apache-2.0). See [`LICENSE`](LICENSE) for details.
