<!-- generated-by: gsd-doc-writer -->
## DEVELOPMENT.md

**Local setup:**
1. Clone the repository.
2. Run `go mod download` (the project is standard-library-only, so this resolves nothing external).
3. Use `go run ./cmd/weather` for local testing.

**Build commands:**
| Command | Description |
|---------|-------------|
| `go build -o weather-cli ./cmd/weather` | Builds the CLI executable |
| `go build -ldflags "-X main.version=v1.2.0" -o weather-cli ./cmd/weather` | Builds with an injected version string, reported by `--version` |
| `go run ./cmd/weather` | Runs the CLI without compiling an executable |
| `gofmt -w .` | Formats all Go source files |
| `gofmt -l .` | Lists misformatted files; prints nothing when everything is clean |

**Code style:**
- Standard library only. No 3rd-party dependencies.
- `gofmt` is mandatory. `gofmt -l .` must print nothing before you commit.
- Keep `.go` files LF-terminated (`.gitattributes` enforces `eol=lf`).
- Colors are never gated on `runtime.GOOS`; they are controlled by the `NO_COLOR` environment variable only.
- Mock HTTP integrations in tests with `net/http/httptest` by overriding the package-level `*BaseURL`/`apiURL` vars. Tests must not touch the network.

**Branch conventions:**
No convention documented.

**PR process:**
- Ensure `gofmt -l .` is empty and `go test -race ./... -count=1` passes locally before submitting.
- CI (`.github/workflows/ci.yml`) re-runs `go build`, `go vet`, `gofmt -l` and `go test -race` on Ubuntu, Windows and macOS. All four gates must be green.
- Follow the guidelines in [CONTRIBUTING.md](CONTRIBUTING.md).
