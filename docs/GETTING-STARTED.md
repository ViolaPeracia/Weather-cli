<!-- generated-by: gsd-doc-writer -->
## GETTING-STARTED.md

**Prerequisites:**
- `Go >= 1.22` (see `go.mod`)
- Network access to `ipwho.is` and the Open-Meteo APIs

No API key is required.

**Installation steps:**
1. Clone the repository:
   ```bash
   git clone https://github.com/ViolaPeracia/Weather-cli.git
   cd Weather-cli
   ```
2. Build the project:
   ```bash
   go build -o weather-cli ./cmd/weather
   ```
   On Windows PowerShell:
   ```powershell
   go build -o weather-cli.exe .\cmd\weather
   ```
   To stamp a version into the binary, use linker flags (the `--version` flag reads it):
   ```bash
   go build -ldflags "-X main.version=v1.2.0" -o weather-cli ./cmd/weather
   ```

**First run:**
```bash
./weather-cli
```
With no `--city`, the CLI detects your location from your IP address. Pass `--city "London"` to target a specific city, or `--tui` for the interactive dashboard.

**Common setup issues:**
- `gofmt -l .` lists every `.go` file after cloning on Windows: your Git client is converting line endings to CRLF. `.gitattributes` pins `.go` files to LF; re-clone after pulling it, or run `git config core.autocrlf input`.
- Output is full of escape sequences when piping to a file or another program: set `NO_COLOR=1`.
- Reports `failed to fetch IP geolocation`: outbound HTTPS is blocked or `ipwho.is` is unreachable. Use `--city` to skip IP detection entirely.
- `invalid unit 'x'. Must be 'celsius' or 'fahrenheit'`: the `--unit` flag (and `unit` in the config file) only accepts those two values.

**Next steps:**
- See [CONFIGURATION.md](CONFIGURATION.md) for settings and the cache file.
- See [DEVELOPMENT.md](DEVELOPMENT.md) for local development setup.
- See [TESTING.md](TESTING.md) for running the test suite.
