# AI Policy & Coding Rules

> [!WARNING]
> These rules are absolute. Any AI agent (Antigravity) operating in this workspace MUST follow them without exception.

## 1. Core Directives
- **Think Before Writing:** Always plan the architecture before writing code. Propose the approach using an implementation plan if the change is significant.
- **Minimal Dependencies:** Rely exclusively on the Go standard library (`os`, `net/http`, `fmt`, `flag`, etc.). If a third-party package is unavoidable, justify its use first.
- **Cross-Platform Compatibility:** Code MUST work natively on Linux and Windows. Use OS-detection logic (`runtime.GOOS`) if OS-specific shell behavior or terminal formatting is needed.

## 2. Code Quality & Style
- **Error Handling:** Never let the app crash with a raw stack trace (`panic`). Catch errors gracefully and return formatted, user-friendly messages to the terminal.
- **Clean Output:** Terminal output must be visually appealing. Format output nicely with alignments, and consider utilizing standard ANSI color codes (ensure Windows compatibility via `golang.org/x/sys/windows` or standard library equivalents if possible).
- **Secrets Management:** NEVER hardcode API keys or sensitive data. Instruct the user to use environment variables (e.g., `WEATHER_API_KEY`) or a local configuration file.

## 3. Workflow & Interaction
- **Respect Existing Code:** Do not remove existing features unless explicitly instructed by the user.
- **Keep Docs Updated:** Automatically update `docs/FEATURES.md` whenever a feature is successfully implemented and verified. Update `docs/ARCHITECTURE.md` if package structure changes.
