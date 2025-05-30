// internal/util/color.go
package util

// Color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"

	StyleBold      = "\033[1m"
	StyleUnderline = "\033[4m"
	StyleInvert    = "\033[7m"
)

// Red returns a red string
func Red(s string) string {
	return ColorRed + s + ColorReset
}

// Green returns a green string
func Green(s string) string {
	return ColorGreen + s + ColorReset
}

// Yellow returns a yellow string
func Yellow(s string) string {
	return ColorYellow + s + ColorReset
}

// Bold returns a bold string
func Bold(s string) string {
	return StyleBold + s + ColorReset
}

// Invert returns an inverted color string
func Invert(s string) string {
	return StyleInvert + s + ColorReset
}

// Warning returns a formatted warning message in yellow
func Warning(s string) string {
	return Yellow("⚠️  " + s)
}

// Error returns a formatted error message in red
func Error(s string) string {
	return Red("✗ " + s)
}

// Success returns a formatted success message in green
func Success(s string) string {
	return Green("✓ " + s)
}

// FormatDeleteItem formats a string to indicate it will be deleted
func FormatDeleteItem(s string) string {
	return Red(StyleInvert + "[DELETE] " + s + ColorReset)
}

// FormatConfirm formats a confirmation prompt
func FormatConfirm(question string) string {
	return Bold(question) + " " + Yellow("[y/N]")
}
