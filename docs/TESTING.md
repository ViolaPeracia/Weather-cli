<!-- generated-by: gsd-doc-writer -->
## TESTING.md

**Test framework and setup:**
- Framework: `go test` (standard library). No third-party test libraries.
- Setup: Ensure `go mod tidy` has been run. The project has no external dependencies, so `go mod download` is a no-op in practice.

**Running tests:**
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Match what CI runs (no cached results, race detector on)
go test -race ./... -count=1
```

**Writing new tests:**
- Append `_test.go` to the file name you are testing (e.g., `main_test.go`).
- Use the standard `testing` package.
- Mock HTTP integrations with `net/http/httptest` by overriding the package-level `apiURL` / `*BaseURL` vars, never by hitting the real network.

**Coverage requirements:**
No coverage threshold configured.

**CI integration:**

A GitHub Actions pipeline lives at `.github/workflows/ci.yml` and runs on every push and pull request. Each job runs on Ubuntu, Windows and macOS, because the CLI mixes ANSI output, `path/filepath` path handling and raw stdin reads - exactly the places where Windows and Linux disagree.

Four gates must all pass:

| Gate | Command | Catches |
| --- | --- | --- |
| Build | `go build ./...` | Compile errors on every supported platform |
| Vet | `go vet ./...` | Suspicious constructs (bad printf verbs, unreachable code, shadowed mistakes) |
| Format | `gofmt -l .` | Formatting drift **and CRLF reintroduction** |
| Test | `go test -race ./... -count=1` | Test failures and data races, with no cached results |

Notes on the gates:

- There is **no golangci-lint**. The project has no external dependencies and no linter config; adding a third-party linter would violate its standard-library-only and YAGNI rules. `go vet` plus `gofmt` cover every class of issue found in this codebase so far.
- The `gofmt -l .` gate is written in CI as "fail if the command produces any output", because `gofmt -l` exits `0` whether or not it lists files.
- `gofmt` rewrites CRLF to LF, so a CRLF checkout on Windows would fail the format gate. `.gitattributes` pins `*.go text eol=lf` to prevent that.
