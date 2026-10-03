# Weather CLI - Development Roadmap

This document outlines the phased development plan for the Weather CLI project. It provides a strategic view of how the project will evolve from scaffolding to a feature-rich application.

---

## Current Status

All five planned phases are complete. The shipped CLI covers:

- IP-based location detection via `https://ipwho.is/`, geocoding and forecasts via Open-Meteo. No API key needed.
- `--city`, `--unit`, `--forecast`, `--force`, `--tui`, `--save-config`, `--config-show`, `--config-set` and `--version` flags.
- A 10-minute local response cache at `~/.weather-cli/cache.json`, bypassable with `--force`.
- Color-coded ASCII output with `NO_COLOR` support, sized by rune count so long city names do not break the box border.
- Cross-platform CI (`.github/workflows/ci.yml`) gating every push and pull request with `go build`, `go vet`, `gofmt -l` and `go test -race` on Ubuntu, Windows and macOS.

Ongoing work now comes from the [features tracker](FEATURES.md) rather than new roadmap phases. See [ROADMAP.md status notes](../.planning/ROADMAP.md) for milestone-level planning.

---

## 🟢 Phase 1: Foundation (MVP) [COMPLETE]
**Goal:** Establish the project structure and basic CLI interactions.

- [x] **Project Scaffolding:** Create the standard Go directory structure (`cmd/weather`, `internal/`).
- [x] **CLI Argument Parsing:** Implement the `flag` package to handle basic inputs like `--city`, `--help`, and `--version`.
- [x] **Basic Output & Error Handling:** Ensure the CLI can run, print dummy text, and handle basic execution errors gracefully (no panics).

---

## 🟢 Phase 2: Core Data Integration [COMPLETE]
**Goal:** Connect to external services to fetch real data.

- [x] **IP Geolocation Detection:** Implement logic in `internal/location` to auto-detect the user's city/coordinates if `--city` is not provided (using `https://ipwho.is/` over HTTPS).
- [x] **Weather API Client:** Implement the HTTP client in `internal/weather` to call a weather service (e.g., OpenWeatherMap).
- [x] **Data Parsing:** Unmarshal the JSON API responses into strongly typed Go structs.

---

## 🟢 Phase 3: Presentation & User Experience [COMPLETE]
**Goal:** Make the output beautiful and cross-platform compatible.

- [x] **Terminal Formatting:** Create clean, aligned text output for temperature, humidity, and weather conditions.
- [x] **Cross-Platform Colors:** Implement color-coded output (e.g., Blue for cold, Red for hot, Yellow for sunny) ensuring it works on both Linux terminals and Windows (CMD/PowerShell).
- [x] **Robust Error Messages:** Format network errors or "city not found" errors into user-friendly terminal messages.

---

## 🟢 Phase 4: Advanced Features & Polish [COMPLETE]
**Goal:** Add convenience features that make the tool a daily driver.

- [x] **Local Configuration:** Allow users to save a default location to a config file (`~/.weather-cli/config.json`) to bypass IP detection and speed up execution.
- [x] **Unit Support:** Add flags to switch between Metric and Imperial units (e.g., `--unit fahrenheit`).
- [x] **Extended Forecasts:** Add a flag (e.g., `--forecast 3`) to show a multi-day weather forecast.
- [x] **ASCII Art / Icons:** Optionally integrate simple terminal icons (e.g., ⛅, 🌧️) if the terminal supports UTF-8.

---

## 🟢 Phase 5: Interactive Terminal User Interface (TUI) [COMPLETE]
**Goal:** Create a real-time console dashboard with interactive settings.

- [x] **Interactive Input Processing:** Read console inputs in raw or line-buffered mode to execute dynamic key commands.
- [x] **ANSI Dashboard Redrawing:** Use escape codes to clear screen and redeliver beautiful layouts on city, unit, and forecast changes.
- [x] **Console Help Screen:** Implement an interactive page mapping commands and explaining caching and config storage.
- [x] **Graceful Program Exit:** Support clean termination on quit key presses.

---

> *Note: For a granular list of individual tasks and their current status, please refer to [FEATURES.md](FEATURES.md).*
