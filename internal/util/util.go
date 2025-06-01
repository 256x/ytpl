// internal/util/util.go
package util

import (
	"fmt"
	"os"
	"strings"
	"time"

	fuzzyfinder "github.com/koki-develop/go-fzf"
)

// SelectFromList prompts the user to select an item from a list using fzf.
func SelectFromList[T any](label string, items []T, displayFunc func(T) string) (T, error) {
	if len(items) == 0 {
		var zero T
		return zero, fmt.Errorf("no items to select from")
	}

	f, err := fuzzyfinder.New(fuzzyfinder.WithPrompt(label + "> "))
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to initialize fzf: %w", err)
	}

	idxs, err := f.Find(
		items,
		func(i int) string {
			return displayFunc(items[i])
		},
	)

	if err != nil {
		var zero T
		if err == fuzzyfinder.ErrAbort {
			return zero, fmt.Errorf("selection cancelled by user")
		}
		return zero, fmt.Errorf("failed to select item: %w", err)
	}

	if len(idxs) == 0 {
		var zero T
		return zero, fmt.Errorf("no item selected")
	}

	return items[idxs[0]], nil
}

// PromptForInput asks the user for a string input.
func PromptForInput(label string, defaultValue string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s [%s]: ", label, defaultValue)

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		if err.Error() == "unexpected newline" {
			return defaultValue, nil
		}
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	if input == "" {
		return defaultValue, nil
	}

	return input, nil
}

// Confirm asks for a yes/no confirmation from the user.
func Confirm(question string) (bool, error) {
	for {
		fmt.Fprintf(os.Stderr, "%s [y/N]: ", question)

		var response string
		_, err := fmt.Scanln(&response)
		if err != nil && err.Error() != "unexpected newline" {
			return false, fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response == "" {
			return false, nil
		}
		if response == "y" || response == "yes" {
			return true, nil
		}
		if response == "n" || response == "no" {
			return false, nil
		}

		fmt.Fprintln(os.Stderr, "please type 'y' or 'n'")
	}
}

// --- Spinner functions ---

var spinnerChars = []string{"-", "\\", "|", "/"}

// StartSpinner starts a new goroutine to display a spinner animation.
// It returns a channel that can be used to stop the spinner.
func StartSpinner(message string) chan struct{} {
	stop := make(chan struct{})
	go func() {
		for i := 0; ; i++ {
			select {
			case <-stop:
				// Clear the spinner line by moving to the beginning of the line
				// and overwriting with spaces, then return to line start.
				fmt.Print("\r" + strings.Repeat(" ", len(message)+len(spinnerChars[0])+2) + "\r")
				return
			default:
				// Move to the beginning of the line and print the current spinner frame.
				fmt.Printf("\r%s %s ", message, spinnerChars[i%len(spinnerChars)])
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return stop
}

// StopSpinner stops the spinner animation.
// It waits a short moment to ensure the spinner goroutine has cleared the line.
func StopSpinner(stopCh chan struct{}) {
	close(stopCh)
	time.Sleep(100 * time.Millisecond) // Give the spinner goroutine a moment to finish
	fmt.Print("\r\033[K")               // Clear the line after spinner
}

// FormatDuration converts seconds to a clean MM:SS or HH:MM:SS format
func FormatDuration(seconds float64) string {
	total := int(seconds)
	hours := total / 3600
	minutes := (total % 3600) / 60
	secs := total % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}
