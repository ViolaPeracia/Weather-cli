# Weather CLI - Features Roadmap

This document tracks the capabilities of the CLI application. The AI agent (Antigravity) MUST update this file as new features are implemented.

## Implemented Features
- [x] Setup initial Go project structure (`cmd/weather/main.go`).
- [x] Basic CLI argument parsing using the `flag` standard library (e.g., `--help`, `--city`).
- [x] Auto-detect user location via IP address API (`https://ipwho.is/` over HTTPS).
- [x] Fetch current weather data from an external API (Open-Meteo forecast + geocoding).
- [x] Parse JSON response and map it to Go structs.
- [x] Display clean, formatted weather output (Temperature, Humidity, Conditions) in the terminal.
- [x] Implement color-coded terminal output (e.g., blue for cold, red for hot) compatible with Windows Command Prompt and Linux, with `NO_COLOR` support to disable all ANSI escapes.
- [x] Allow saving a default location to a local config file (e.g., `~/.weather-cli/config.json`) so the IP detection isn't needed every run.
- [x] Add forecasting capabilities (e.g., `--forecast 3` for a 3-day forecast).
- [x] Support multiple units (Celsius/Fahrenheit) via flag (e.g., `--unit C`).
- [x] Add offline-first unit tests for core packages (`weather`, `display`, `location`, `config`) mocking HTTP integrations.
- [x] Dynamically adjust console box-width to prevent layout border wrapping for long city names.
- [x] **Interactive TUI Dashboard:** Real-time console dashboard loop featuring alternate screen rendering, geocoding lookups, and toggles for temperature format & forecasts (triggered via the `--tui` flag; Go's `flag` package also accepts `-tui`).
- [x] **Interactive Help Screen:** Guide users on command shortcuts, local config, caching directory paths, and exit parameters.
- [x] **Local Response Cache:** Cache weather responses at `~/.weather-cli/cache.json` for 10 minutes, keyed by coordinates, unit and forecast length. Bypass with `--force`.

## Continuous Integration
- [x] **Cross-platform CI:** GitHub Actions workflow (`.github/workflows/ci.yml`) runs `go build`, `go vet`, `gofmt -l` and `go test -race` on Ubuntu, Windows and macOS for every push and pull request.
- [x] **Line-ending enforcement:** `.gitattributes` pins `.go` files to LF so the `gofmt` gate cannot be broken by Windows CRLF checkouts.
- [x] **Injectable version string:** `go build -ldflags "-X main.version=v1.2.0"` stamps the version reported by `--version`.

## Work in Progress
*(None currently. Ready for new features!)*

## Planned Features
*(None currently. Ready for new features!)*

## Backlog / Ideas
- [x] ~~Add caching for weather API responses to avoid rate limits.~~ *(Shipped - see Implemented Features above.)*
- [ ] Add hourly forecast display.
- [ ] Add weather alerts (e.g., severe weather warnings).
- [ ] Support custom API keys via config/ENV. *(Niche: current Open-Meteo / ipwho.is integration needs no key.)*
