// internal/util/spinner.go
package util

import (
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

type Spinner struct {
	s *spinner.Spinner
	mu sync.Mutex
}

// Spinner styles - these are just a few examples, see spinner.CharSets for all available styles
const (
	StyleLine     = 14 // Default line style (|/-\)
	StyleDots     = 9  // Simple dot animation
	StyleBouncing = 11 // Bouncing bar
	StyleGrow     = 26 // Growing bar
	StyleMoon     = 30 // Moon phases
	StyleArrow    = 35 // Rotating arrows
)

// NewSpinner creates a new spinner with the given message and default style
func NewSpinner(message string) *Spinner {
	return NewSpinnerWithStyle(message, StyleLine)
}

// NewSpinnerWithStyle creates a new spinner with the given message and style
func NewSpinnerWithStyle(message string, style int) *Spinner {
	// Ensure style is within bounds
	if style < 0 || style >= len(spinner.CharSets) {
		style = StyleLine // Fallback to default style
	}

	s := spinner.New(spinner.CharSets[style], 100*time.Millisecond)
	s.Prefix = "- "
	s.Suffix = " " + message
	s.Start()
	return &Spinner{s: s}
}

// UpdateMessage updates the spinner's message
func (sp *Spinner) UpdateMessage(message string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.s.Suffix = " " + message
}

// Stop stops the spinner with a final message
func (sp *Spinner) Stop(finalMessage string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	
	if finalMessage != "" {
		sp.s.FinalMSG = "- " + finalMessage + "\n"
	} else {
		sp.s.FinalMSG = ""
	}
	sp.s.Stop()
}

// StopWithSuccess stops the spinner with a success message
func (sp *Spinner) StopWithSuccess(message string) {
	sp.Stop("✓ " + message)
}

// StopWithError stops the spinner with an error message
func (sp *Spinner) StopWithError(message string) {
	sp.Stop("✗ " + message)
}
