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
			fmt.Println("\n- no song is currently playing.\n")
			return
		}
		if appState.IsPlaying {
			if err := player.Pause(appState); err != nil {
				log.Fatalf("error pausing player: %v", err)
			}
			fmt.Println("\n- paused.\n")
		} else {
			fmt.Println("\n- song is already paused.\n")
		}
	},
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume the paused song",
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("\n- no song is currently playing or paused.\n")
			return
		}
		if !appState.IsPlaying {
			if err := player.Resume(appState); err != nil {
				log.Fatalf("error resuming player: %v", err)
			}
			fmt.Println("\n- resumed.\n")
		} else {
			fmt.Println("\n- song is already playing.\n")
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
		if err := player.StopPlayer(appState); err != nil {
			log.Fatalf("error stopping player: %v", err)
		}
		fmt.Println("\n- stopped.\n")
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
			log.Printf("Error: Invalid volume percentage: %v", err)
			return
		}

		// Adjust volume if out of range
		adjustedVolume := volume
		if volume > 100 {
			adjustedVolume = 100
			log.Printf("volume adjusted to maximum (100%%)")
		} else if volume < 0 {
			adjustedVolume = 0
			log.Printf("volume set to mute (0%%)")
		}

		if err := player.SetVolume(appState, adjustedVolume); err != nil {
			log.Printf("error setting volume: %v", err)
			return
		}
		
		if adjustedVolume != volume {
			fmt.Printf("\n- volume set to %d%% (adjusted from %d%%)\n\n", adjustedVolume, volume)
		} else {
			fmt.Printf("\n- vol: %d%%\n\n", adjustedVolume)
		}
	},
}
