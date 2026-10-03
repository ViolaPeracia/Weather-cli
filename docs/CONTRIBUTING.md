# Contributing Guidelines

Thank you for investing your time in contributing to the Weather CLI project!

## AI-Assisted Development
This project is actively developed with the help of AI coding agents (like Antigravity). All contributors (human or AI) must strictly follow the rules defined in `AI_POLICY_RULES.md` and the instructions in `AGENTS.md`.

## Code Standards
- **Standard Library Only:** Avoid introducing 3rd-party dependencies unless absolutely necessary and previously discussed.
- **Formatting:** All Go code must be formatted using `gofmt` before committing.
- **Line Endings:** Keep `.go` files LF-terminated. `.gitattributes` pins `eol=lf`, and CI fails if `gofmt -l .` lists anything.
- **Error Handling:** Avoid panics. Return errors explicitly and format them nicely for the CLI user.

## Continuous Integration

`.github/workflows/ci.yml` runs on every push and pull request across Ubuntu, Windows and macOS, and enforces four gates: `go build ./...`, `go vet ./...`, `gofmt -l .` (fails if it prints anything), and `go test -race ./... -count=1`. All four must be green before a PR can be merged. Details in [TESTING.md](TESTING.md).

## Process for Adding Features
1. Check `docs/FEATURES.md` to pick up a task or add a new planned feature.
2. Discuss the architectural approach if it requires a new package or major refactoring.
3. Write clean, cross-platform code.
4. Update `docs/FEATURES.md` to reflect your progress (move from 'Planned' to 'Work in Progress' to 'Implemented').
