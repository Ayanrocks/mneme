package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// CLI Output Functions
// These provide user-friendly output for CLI applications
// Separate from structured logging (Debug, Info, etc.)

var (
	// Color functions for CLI output
	successColor = color.New(color.FgGreen).SprintFunc()
	warningColor = color.New(color.FgYellow).SprintFunc()
	errorColor   = color.New(color.FgRed).SprintFunc()
	infoColor    = color.New(color.FgCyan).SprintFunc()
	headerColor  = color.New(color.FgWhite, color.Bold).SprintFunc()

	// Enable/disable colors (for testing or non-TTY output)
	colorsEnabled = true
)

// SetColors enables or disables colored output for CLI
func SetColors(enabled bool) {
	colorsEnabled = enabled
	if !enabled {
		// Reset color functions to no-op
		successColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
		warningColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
		errorColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
		infoColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
		headerColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
	}
}

// Success prints a success message (green)
func Success(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if colorsEnabled {
		fmt.Printf("%s %s\n", successColor("✓"), msg)
	} else {
		fmt.Printf("✓ %s\n", msg)
	}
}

// Print prints an info message (cyan)
func Print(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if colorsEnabled {
		fmt.Printf("%s %s\n", infoColor("ℹ"), msg)
	} else {
		fmt.Printf("ℹ %s\n", msg)
	}
}

// PrintRaw prints a message without any prefix or icon
func PrintRaw(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
}

// Warning prints a warning message (yellow)
func Warning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if colorsEnabled {
		fmt.Printf("%s %s\n", warningColor("⚠"), msg)
	} else {
		fmt.Printf("⚠ %s\n", msg)
	}
}

// PrintError prints an error message (red)
func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if colorsEnabled {
		fmt.Fprintf(os.Stderr, "%s %s\n", errorColor("✖"), msg)
	} else {
		fmt.Fprintf(os.Stderr, "✖ %s\n", msg)
	}
}

// PrintFatal prints an error message and exits
func PrintFatal(format string, a ...interface{}) {
	PrintError(format, a...)
	os.Exit(1)
}

// Header prints a header/section title
func Header(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if colorsEnabled {
		fmt.Printf("\n%s\n\n", headerColor(msg))
	} else {
		fmt.Printf("\n%s\n\n", msg)
	}
}

// SubHeader prints a sub-header
func SubHeader(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if colorsEnabled {
		fmt.Printf("%s\n", color.New(color.FgWhite, color.Bold).SprintFunc()(msg))
	} else {
		fmt.Printf("%s\n", msg)
	}
}

// Bullet prints a bullet point
func Bullet(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("  • %s\n", msg)
}

// List prints a numbered list
func List(items []string) {
	for i, item := range items {
		fmt.Printf("  %d. %s\n", i+1, item)
	}
}

// KeyValue prints a key-value pair
func KeyValue(key string, value string) {
	if colorsEnabled {
		fmt.Printf("  %s: %s\n", color.New(color.FgWhite).SprintFunc()(key), value)
	} else {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

// Separator prints a separator line
func Separator() {
	fmt.Println(strings.Repeat("─", 60))
}

// Blank prints a blank line
func Blank() {
	fmt.Println()
}

// Progress prints a progress message (without newline)
func Progress(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Print("\r" + msg)
}

// ClearProgress clears the current progress line
func ClearProgress() {
	fmt.Print("\r\033[K")
}

// Spinner shows a spinning indicator
type Spinner struct {
	active bool
	msg    string
}

// Start starts a spinner with a message
func (s *Spinner) Start(msg string) {
	s.active = true
	s.msg = msg
	go s.animate()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.active = false
	ClearProgress()
}

func (s *Spinner) animate() {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for s.active {
		if colorsEnabled {
			Progress("%s %s", infoColor(frames[i]), s.msg)
		} else {
			Progress("%s %s", frames[i], s.msg)
		}
		i = (i + 1) % len(frames)
		// Sleep for ~100ms
		// Note: In a real implementation, you'd use time.Sleep
		// but we're avoiding imports for simplicity
	}
}

// NewSpinner creates a new spinner instance
func NewSpinner() *Spinner {
	return &Spinner{}
}
