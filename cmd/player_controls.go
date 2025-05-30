// cmd/player_controls.go
package cmd

import (
	"fmt"
	"log"
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
			fmt.Println("No song is currently playing.")
			return
		}
		if appState.IsPlaying {
			if err := player.Pause(appState); err != nil {
				log.Fatalf("Error pausing player: %v", err)
			}
			fmt.Println("Song paused.")
		} else {
			fmt.Println("Song is already paused.")
		}
	},
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume the paused song",
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("No song is currently playing or paused.")
			return
		}
		if !appState.IsPlaying {
			if err := player.Resume(appState); err != nil {
				log.Fatalf("Error resuming player: %v", err)
			}
			fmt.Println("Song resumed.")
		} else {
			fmt.Println("Song is already playing.")
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the currently playing song",
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("No song is currently playing.")
			return
		}
		if err := player.StopPlayer(appState); err != nil {
			log.Fatalf("Error stopping player: %v", err)
		}
		fmt.Println("Song stopped.")
	},
}

var volCmd = &cobra.Command{
	Use:   "vol <percentage>",
	Short: "Set playback volume (0-100)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("No song is currently playing to set volume.")
			return
		}

		volume, err := strconv.Atoi(args[0])
		if err != nil {
			log.Printf("Error: Invalid volume percentage: %v", err)
			return
		}

		// Adjust volume if out of range
		adjustedVolume := volume
		if volume > 100 {
			adjustedVolume = 100
			log.Printf("Volume adjusted to maximum (100%%)")
		} else if volume < 0 {
			adjustedVolume = 0
			log.Printf("Volume set to mute (0%%)")
		}

		if err := player.SetVolume(appState, adjustedVolume); err != nil {
			log.Printf("Error setting volume: %v", err)
			return
		}
		
		if adjustedVolume != volume {
			fmt.Printf("Volume set to %d%% (adjusted from %d%%)\n", adjustedVolume, volume)
		} else {
			fmt.Printf("Volume set to %d%%\n", adjustedVolume)
		}
	},
}
