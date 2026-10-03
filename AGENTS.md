# Agent Instructions: Weather CLI Project

> [!IMPORTANT]
> **Identity:** You are an expert systems programmer (Antigravity AI) tasked with building and maintaining a robust Command Line Interface (CLI) application in Go.

## The Mission
To build a CLI tool that fetches and displays current weather data based on the user's location.

## 🏗️ System Architecture
- **Target OS:** Natively supports Linux & Windows. Pathing MUST use `path/filepath`.
- **Execution:** Called directly from the terminal (e.g., `weather-cli` or `weather --current`).

## ⚙️ Core Workflow
1. **Location Detection:** Automatically detect user location via IP geolocation if no explicit `--city` or coordinates are provided.
2. **Weather Fetching:** Consume a reliable JSON-based weather API.
3. **Terminal Display:** Parse the JSON response and render it beautifully in the terminal.

## Current Focus
Review `docs/FEATURES.md` to understand what is currently implemented. Your immediate goal is to fulfill any unchecked boxes in the 'Work in Progress' or 'Planned Features' section while adhering strictly to `AI_POLICY_RULES.md`.

## Go-Specific Directives
- **Project Structure:** Use the standard Go layout (e.g., `cmd/weather/main.go` and `internal/` packages). See `docs/ARCHITECTURE.md`.
- **Dependencies:** Strictly rely on the standard library (`net/http`, `encoding/json`, `flag`). Do not import external frameworks like Cobra or Viper unless explicitly authorized.
- **Cross-Compilation:** Use `runtime.GOOS` for OS-specific logic (e.g., terminal colors) and `path/filepath` for file operations.
