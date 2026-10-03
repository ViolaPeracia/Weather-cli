package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ViolaPeracia/Weather-cli/internal/config"
	"github.com/ViolaPeracia/Weather-cli/internal/weather"
)

func TestTuiQuitCommand(t *testing.T) {
	input := "q\n"
	stdin := bytes.NewBufferString(input)
	var stdout bytes.Buffer

	cfg := config.Config{
		DefaultCity: "Hanoi",
		Unit:        "celsius",
	}

	runTUI(stdin, &stdout, cfg)

	output := stdout.String()
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("Expected output to contain 'Goodbye!', got:\n%s", output)
	}
}

func TestTuiHelpToggle(t *testing.T) {
	// Toggle help on ("h") and then quit ("q")
	input := "h\nq\n"
	stdin := bytes.NewBufferString(input)
	var stdout bytes.Buffer

	cfg := config.Config{
		DefaultCity: "Hanoi",
		Unit:        "celsius",
	}

	runTUI(stdin, &stdout, cfg)

	output := stdout.String()
	if !strings.Contains(output, "Help & Guidance") {
		t.Errorf("Expected output to contain Help page, got:\n%s", output)
	}
}

func TestTuiUnknownCommandDoesNotPanic(t *testing.T) {
	// "z" is not a valid command; the default branch must report it and loop.
	input := "z\nq\n"
	stdin := bytes.NewBufferString(input)
	var stdout bytes.Buffer

	cfg := config.Config{
		DefaultCity: "Hanoi",
		Unit:        "celsius",
	}

	runTUI(stdin, &stdout, cfg)

	output := stdout.String()
	if !strings.Contains(output, "Unknown command 'z'") {
		t.Errorf("Expected output to report the unknown command, got:\n%s", output)
	}
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("Expected TUI to keep looping and quit on 'q', got:\n%s", output)
	}
}

// TestTuiErrorForUnknownCityReverts covers the revert branch without touching
// the network: geocodeBaseURL is unexported to internal/weather, so an
// end-to-end TUI test would depend on a live Open-Meteo round trip and become
// flaky when offline. Driving the error mapping directly keeps the contract -
// ErrCityNotFound means "fall back to the configured default" - pinned.
func TestTuiErrorForUnknownCityReverts(t *testing.T) {
	// Mirrors what Resolve returns for an unmatched city, plus an extra
	// wrapping layer, to prove errors.Is still sees the sentinel.
	resolveErr := fmt.Errorf("resolving weather: %w",
		fmt.Errorf("%w: %s", weather.ErrCityNotFound, "Atlantis"))

	message, revert := tuiErrorFor("Atlantis", resolveErr)
	if !revert {
		t.Error("Expected an unknown city to revert to the default city")
	}
	want := "City 'Atlantis' not found. Reverting..."
	if message != want {
		t.Errorf("Expected message %q, got %q", want, message)
	}
}

// TestTuiErrorForOtherFailures checks the non-reverting branches keep the
// original wording: IP detection failures are reported as such, everything
// else as a fetch failure.
func TestTuiErrorForOtherFailures(t *testing.T) {
	// Resolve prefixes IP detection failures; the banner must not repeat it.
	detectErr := fmt.Errorf("failed to detect location: %w", errors.New("IP geolocation API returned status: 503"))
	message, revert := tuiErrorFor("", detectErr)
	if revert {
		t.Error("Expected no revert when location detection fails")
	}
	want := "Failed to detect location: IP geolocation API returned status: 503"
	if message != want {
		t.Errorf("Expected message %q, got %q", want, message)
	}

	fetchErr := fmt.Errorf("failed to fetch weather data: %w", errors.New("connection refused"))
	message, revert = tuiErrorFor("Hanoi", fetchErr)
	if revert {
		t.Error("Expected no revert on a fetch failure")
	}
	want = "Failed to fetch weather: failed to fetch weather data: connection refused"
	if message != want {
		t.Errorf("Expected message %q, got %q", want, message)
	}
}
