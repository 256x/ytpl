// cmd/player_controls.go
package cmd

import (
	"fmt"
	"strconv"

	player "ytpl/internal/player" // Alias for internal/player
	// "ytpl/internal/state" // Not directly used, removed

	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the currently playing song",
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("\n- no song is currently playing.\n")
			return
		}
		if appState.IsPlaying {
			// Ignore error when pausing player
			_ = player.Pause(appState)
			fmt.Println("Paused")
		} else {
			fmt.Println("Already paused")
		}
	},
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume the paused song",
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("\n- no song is currently playing or paused\n")
			return
		}
		if !appState.IsPlaying {
			// Ignore error when resuming player
			_ = player.Resume(appState)
			fmt.Println("\n- resumed\n")
		} else {
			fmt.Println("\nalready playing\n")
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the currently playing song",
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("\n- no song is currently playing.\n")
			return
		}
		// Ignore error when stopping player
		_ = player.StopPlayer(appState)
		fmt.Println("\n- stopped\n")
	},
}

var volCmd = &cobra.Command{
	Use:   "vol <percentage>",
	Short: "Set playback volume (0-100)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("\n- no song is currently playing to set volume.\n")
			return
		}

		volume, err := strconv.Atoi(args[0])
		if err != nil {
			// Invalid volume percentage
			return
		}

		// Adjust volume if out of range
		adjustedVolume := volume
		if volume > 100 {
			adjustedVolume = 100
			// Volume adjusted to maximum (100%%)
		} else if volume < 0 {
			adjustedVolume = 0
			// Volume set to mute (0%%)
		}

		if err := player.SetVolume(appState, adjustedVolume); err != nil {
			// Error setting volume
			return
		}
		
		if adjustedVolume != volume {
			fmt.Printf("\n- volume set to %d%%.\n\n", adjustedVolume)
		} else {
			fmt.Printf("\n- volume: %d%%.\n\n", adjustedVolume)
		}
	},
}
