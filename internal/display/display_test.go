package display

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ViolaPeracia/Weather-cli/internal/weather"
)

func TestFormatTempCelsius(t *testing.T) {
	tests := []struct {
		temp        float64
		expectedRaw string
	}{
		{10.0, "10.0°C"},
		{20.0, "20.0°C"},
		{30.0, "30.0°C"},
	}

	for _, tt := range tests {
		raw, col := formatTemp(tt.temp, "C")
		if raw != tt.expectedRaw {
			t.Errorf("Expected raw '%s', got '%s'", tt.expectedRaw, raw)
		}
		if col == "" {
			t.Errorf("Expected colored string, got empty")
		}
	}
}

func TestFormatTempFahrenheit(t *testing.T) {
	raw, col := formatTemp(50.0, "F")
	if raw != "50.0°F" {
		t.Errorf("Expected raw '50.0°F', got '%s'", raw)
	}
	if col == "" {
		t.Error("Expected colored fahrenheit string, got empty")
	}
}

func TestFormatHumidity(t *testing.T) {
	raw, col := formatHumidity(85)
	if raw != "85%" {
		t.Errorf("Expected raw '85%%', got '%s'", raw)
	}
	if col == "" {
		t.Error("Expected colored humidity string, got empty")
	}
}

func TestFormatConditions(t *testing.T) {
	raw, col := formatConditions("Clear sky", "󰖙")
	if raw != "󰖙 Clear sky" {
		t.Errorf("Expected raw '󰖙 Clear sky', got '%s'", raw)
	}
	if col == "" {
		t.Error("Expected colored conditions string, got empty")
	}
}

// visibleRunes counts runes of s ignoring ANSI escape sequences.
func visibleRunes(s string) int {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return utf8.RuneCountInString(b.String())
}

// sampleData returns a representative payload for render tests.
func sampleData() weather.WeatherData {
	return weather.WeatherData{
		Temperature: 25.0,
		Humidity:    80,
		Conditions:  "Clear sky",
		Icon:        "󰖙",
		Unit:        "C",
	}
}

func TestRenderWeatherDynamicBoxWidth(t *testing.T) {
	longCity := "Taumatawhakatangihangakoauauotamateaturipukakapikimaungahoronukupokaiwhenuakitanatahu, New Zealand"
	data := sampleData()

	var buf bytes.Buffer
	RenderWeather(&buf, longCity, data)
	output := buf.String()

	expectedHeaderLen := utf8.RuneCountInString(" Weather for: ") + utf8.RuneCountInString(longCity)

	// The top border must be wide enough to hold the header row.
	var topBorder string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "╭") {
			topBorder = line
			break
		}
	}
	if topBorder == "" {
		t.Fatalf("Expected a top border in output, got:\n%s", output)
	}
	borderLen := utf8.RuneCountInString(topBorder)
	if borderLen < expectedHeaderLen+2 {
		t.Errorf("Expected top border of at least %d runes, got %d", expectedHeaderLen+2, borderLen)
	}

	// The header row itself must not be truncated or wrapped.
	if !strings.Contains(output, longCity) {
		t.Errorf("Expected output to contain the full location name, got:\n%s", output)
	}
	headerLine := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, longCity) {
			headerLine = line
			break
		}
	}
	if got := visibleRunes(headerLine); got != borderLen {
		t.Errorf("Expected header row to be %d runes wide like the border, got %d", borderLen, got)
	}
}

func TestRenderWeatherWritesToProvidedWriter(t *testing.T) {
	// Capture the real stdout to prove RenderWeather does not write to it.
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	var buf bytes.Buffer
	RenderWeather(&buf, "Hanoi", sampleData())

	_ = w.Close()
	os.Stdout = oldStdout

	var stdoutBuf bytes.Buffer
	if _, err := io.Copy(&stdoutBuf, r); err != nil {
		t.Fatalf("Failed to read captured stdout: %v", err)
	}
	_ = r.Close()

	if buf.Len() == 0 {
		t.Error("Expected RenderWeather to write into the provided writer")
	}
	if !strings.Contains(buf.String(), "Hanoi") {
		t.Errorf("Expected provided writer to contain the location name, got:\n%s", buf.String())
	}
	if stdoutBuf.Len() != 0 {
		t.Errorf("Expected nothing on stdout, got:\n%s", stdoutBuf.String())
	}
}

func TestRenderWeatherRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var buf bytes.Buffer
	RenderWeather(&buf, "Hanoi", sampleData())
	output := buf.String()

	if strings.Contains(output, "\033") {
		t.Errorf("Expected no ANSI escape sequences when NO_COLOR is set, got:\n%q", output)
	}
	if !strings.Contains(output, "Hanoi") {
		t.Errorf("Expected plain output to still contain the location name, got:\n%s", output)
	}
}

func TestRenderWeatherEmitsColorByDefault(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	var buf bytes.Buffer
	RenderWeather(&buf, "Hanoi", sampleData())

	if !strings.Contains(buf.String(), "\033[") {
		t.Errorf("Expected ANSI escape sequences by default, got:\n%q", buf.String())
	}
}

func TestFormatForecastRow(t *testing.T) {
	f := weather.DailyForecast{
		Date:       "2023-10-02",
		MinTemp:    15.2,
		MaxTemp:    26.5,
		Conditions: "Clear sky",
		Icon:       "󰖙",
	}

	raw, colored := formatForecastRow(f, "C")

	if !strings.HasPrefix(raw, "10-02") {
		t.Errorf("Expected short date '10-02' at the start of %q", raw)
	}
	if !strings.Contains(raw, "L:15.2°C") || !strings.Contains(raw, "H:26.5°C") {
		t.Errorf("Expected low/high temperatures in %q", raw)
	}
	if colored == "" {
		t.Error("Expected a colored forecast row")
	}
	// The colored variant must carry the same visible content as the raw one.
	if !strings.Contains(colored, "10-02") || !strings.Contains(colored, "L:") || !strings.Contains(colored, "H:") {
		t.Errorf("Expected colored row to keep the visible content, got %q", colored)
	}
}
