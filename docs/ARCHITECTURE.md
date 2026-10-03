# Software Architecture

This document describes the high-level architecture and planned Go project structure for the Weather CLI.

## Directory Structure

We adhere to standard Go project layouts to keep the codebase clean and maintainable.

```text
weather-cli/
├── cmd/
│   └── weather/
│       └── main.go         # Entry point: Parses flags, initializes app, and handles fatal errors
├── internal/
│   ├── cli/                # Interactive TUI dashboard (alternate screen, key handling)
│   ├── config/             # Config file (~/.weather-cli/config.json) and response cache (~/.weather-cli/cache.json)
│   ├── location/           # Logic for IP-based location detection
│   ├── weather/            # API client for geocoding and fetching weather data
│   └── display/            # Terminal output rendering and color coding
├── docs/                   # Documentation and feature tracking
├── AGENTS.md               # AI Agent Mission & Directives
└── AI_POLICY_RULES.md      # AI Coding Rules
```

## ⚙️ Core Components

1. **Config Manager:** Reads and writes `~/.weather-cli/config.json` (default city, default unit) and manages the short-lived response cache at `~/.weather-cli/cache.json`. It uses `os.UserHomeDir` and `path/filepath`, so paths resolve correctly on Linux and Windows. No API key or environment variable is required - both upstream providers (Open-Meteo, ipwho.is) are free and unauthenticated.
2. **Location Service:** Reaches out to `https://ipwho.is/` over HTTPS to determine coordinates/city if the user doesn't pass a `--city` flag. HTTPS is mandatory here: plain HTTP would leak the user's IP address in cleartext and let a network attacker forge the response.
3. **Weather Client:** Sends HTTPS GET requests to the Open-Meteo geocoding and forecast endpoints. Uses `net/http` and `encoding/json` to decode the response.
4. **Cache Layer:** `internal/config/cache.go` stores the last weather response keyed by coordinates, unit and forecast length, with a 10-minute TTL. `--force` bypasses it. Cache misses and write failures are non-fatal, so a broken cache degrades to a plain API call rather than an error.
5. **Display Engine:** Takes the structured weather data and formats it into an ASCII widget, writing to any `io.Writer` so tests can capture output. Box width is derived from `utf8.RuneCountInString` of the content rather than `len()`, so multi-byte location names do not break the right border.

   Color is **not** gated on `runtime.GOOS`. Colors are on by default and disabled only when the `NO_COLOR` environment variable is set to a non-empty value (https://no-color.org), which keeps output readable in pipes, logs and pagers. This works identically on Windows CMD/PowerShell and Linux/macOS terminals.

6. **Resolve Pipeline:** `weather.Resolve(city, unit, days, force, cache)` is the single entry point both `cmd/weather` and the TUI use to turn a location request into weather data (IP detection → geocoding → cache → fetch). Callers pass a `weather.CacheStore` adapter (`configCache`) instead of the config package directly, which keeps `internal/weather` free of an import cycle back into `internal/config`. Geocoding returns no match, `weather.ErrCityNotFound` is reported, letting the TUI fall back to the configured default city while genuinely failing calls are surfaced as errors.
