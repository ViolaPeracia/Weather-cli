<!-- generated-by: gsd-doc-writer -->
## CONFIGURATION.md

**Environment variables:**

The CLI needs no secrets. Both upstream providers (Open-Meteo, ipwho.is) are free and unauthenticated, so there is no API key to set.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NO_COLOR` | Optional | unset | Set to any non-empty value to disable all ANSI color escape sequences. See https://no-color.org. Colors are enabled by default. |

**Config file format:**

Local config is stored at `~/.weather-cli/config.json`.

```json
{
  "default_city": "New York",
  "unit": "celsius"
}
```

- `default_city` (string, optional): city passed to the Open-Meteo geocoding API. If empty, the CLI falls back to IP-based location detection.
- `unit` (string, optional): `celsius` or `fahrenheit`. Defaults to `celsius` when omitted or empty.

**Cache file:**

Weather responses are cached at `~/.weather-cli/cache.json` for 10 minutes, keyed by coordinates, unit and forecast length. Entries older than the TTL, or recorded for different coordinates/unit/forecast, are ignored. Pass `--force` to skip the cache and always hit the network. Cache read/write failures are non-fatal: the CLI falls back to a live API call.

**Required vs optional settings:**
Nothing is required. Without a saved city, Weather CLI attempts IP-based location detection; without a saved unit, it uses Celsius.

**Inspecting and editing settings from the CLI:**

```bash
./weather-cli --config-show                    # print the active defaults
./weather-cli --config-set city=Hanoi          # set default_city
./weather-cli --config-set unit=fahrenheit     # set unit
./weather-cli --city Paris --unit celsius --save-config   # save a one-off run as the new default
```

**Per-environment overrides:**
Not applicable for this CLI tool.
