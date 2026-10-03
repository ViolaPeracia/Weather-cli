// Package display renders weather data as ASCII widgets for the terminal.
//
// Color is enabled by default. Setting the NO_COLOR environment variable to a
// non-empty value (https://no-color.org) disables every ANSI escape sequence,
// which keeps the output readable in pipes, logs and pagers.
package display

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/ViolaPeracia/Weather-cli/internal/weather"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

const (
	// baseBoxWidth is the minimum inner width of the ASCII widget; it grows to
	// fit the widest row when location names or conditions are longer.
	baseBoxWidth = 50
	// condWidth pads the condition column of a forecast row so the low/high
	// temperatures line up across rows.
	condWidth = 16
)

// colorEnabled reports whether ANSI colors should be emitted. Colors are on by
// default and turned off whenever NO_COLOR is set to a non-empty value.
func colorEnabled() bool {
	return os.Getenv("NO_COLOR") == ""
}

// colorize wraps s in an ANSI color sequence, or returns s unchanged when
// colors are disabled.
func colorize(color, s string) string {
	if !colorEnabled() {
		return s
	}
	return color + s + ColorReset
}

// PrintError formats and prints error messages to Stderr.
func PrintError(err error) {
	if !colorEnabled() {
		fmt.Fprintf(os.Stderr, "✖ Error: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "%s%s✖ Error:%s %v\n", ColorBold, ColorRed, ColorReset, err)
}

// formatTemp colorizes the temperature based on its value (Celsius).
// For Fahrenheit, it roughly maps equivalent ranges.
func formatTemp(temp float64, unit string) (string, string) {
	raw := fmt.Sprintf("%.1f°%s", temp, unit)

	celsiusTemp := temp
	if unit == "F" {
		celsiusTemp = (temp - 32) * 5 / 9
	}

	if celsiusTemp <= 15 {
		return raw, colorize(ColorBlue, raw)
	} else if celsiusTemp <= 27 {
		return raw, colorize(ColorGreen, raw)
	}
	return raw, colorize(ColorRed, raw)
}

// formatHumidity colorizes the humidity.
func formatHumidity(hum int) (string, string) {
	raw := fmt.Sprintf("%d%%", hum)
	return raw, colorize(ColorCyan, raw)
}

// formatConditions colorizes the conditions.
func formatConditions(cond string, icon string) (string, string) {
	raw := fmt.Sprintf("%s %s", icon, cond)
	return raw, colorize(ColorYellow, raw)
}

// formatForecastRow builds a single forecast row.
// It returns the plain string (used to measure the box width) and the
// colorized string (used to render), so both always stay in sync.
func formatForecastRow(f weather.DailyForecast, unit string) (string, string) {
	// e.g., "2023-10-02" -> "10-02"
	shortDate := f.Date
	if len(shortDate) >= 5 {
		shortDate = shortDate[5:]
	}

	rawMin, colMin := formatTemp(f.MinTemp, unit)
	rawMax, colMax := formatTemp(f.MaxTemp, unit)

	// Format: "10-02  󰖙 Clear sky" and "L:15.2°C  H:26.5°C"
	// To ensure clean alignment, we pad the condition section
	condRaw := fmt.Sprintf("%s %s", f.Icon, f.Conditions)
	padCond := condWidth - utf8.RuneCountInString(condRaw)
	if padCond < 0 {
		padCond = 0
	}
	condPadded := condRaw + strings.Repeat(" ", padCond)

	raw := fmt.Sprintf("%s  %s  L:%s  H:%s", shortDate, condPadded, rawMin, rawMax)
	colored := fmt.Sprintf("%s%s%s  %s  L:%s  H:%s",
		ColorCyan, shortDate, ColorReset, colorize(ColorYellow, condPadded), colMin, colMax)
	return raw, colored
}

// printRow prints a padded row of width (excluding the borders) to w.
func printRow(w io.Writer, width int, label, rawValue, coloredValue string) {
	contentLen := utf8.RuneCountInString(label) + utf8.RuneCountInString(rawValue)
	padding := width - contentLen
	if padding < 0 {
		padding = 0
	}
	fmt.Fprintf(w, "│ %s%s%s │\n", label, coloredValue, strings.Repeat(" ", padding))
}

// RenderWeather writes the weather data and forecast inside a beautiful ASCII
// widget to w. The box width is derived from the content, so nothing is shared
// between calls and the function is safe for concurrent use.
func RenderWeather(w io.Writer, locationName string, data weather.WeatherData) {
	width := boxContentWidth(locationName, data)

	border := strings.Repeat("─", width+2)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "╭%s╮\n", border)

	// Header row
	headerPrefix := " Weather for: "
	printRow(w, width, headerPrefix, locationName, colorize(ColorBold, locationName))

	fmt.Fprintf(w, "├%s┤\n", border)

	// Data rows
	rawT, colT := formatTemp(data.Temperature, data.Unit)
	printRow(w, width, " Temperature: ", rawT, colT)

	rawH, colH := formatHumidity(data.Humidity)
	printRow(w, width, " Humidity:    ", rawH, colH)

	rawC, colC := formatConditions(data.Conditions, data.Icon)
	printRow(w, width, " Conditions:  ", rawC, colC)

	// Forecast rows
	if len(data.Forecast) > 0 {
		fmt.Fprintf(w, "├%s┤\n", border)
		printRow(w, width, " Forecast:    ", "", "")
		for _, f := range data.Forecast {
			rawF, colF := formatForecastRow(f, data.Unit)
			printRow(w, width, "  ", rawF, colF)
		}
	}

	fmt.Fprintf(w, "╰%s╯\n", border)
	fmt.Fprintln(w)
}

// boxContentWidth returns the inner width needed to fit the header and every
// row without breaking the right border.
func boxContentWidth(locationName string, data weather.WeatherData) int {
	width := baseBoxWidth

	// Check location header length
	headerLen := utf8.RuneCountInString(" Weather for: ") + utf8.RuneCountInString(locationName)
	if headerLen > width {
		width = headerLen
	}

	// Check forecast rows if any
	for _, f := range data.Forecast {
		rawF, _ := formatForecastRow(f, data.Unit)
		rowLen := utf8.RuneCountInString("  ") + utf8.RuneCountInString(rawF)
		if rowLen > width {
			width = rowLen
		}
	}

	return width
}
