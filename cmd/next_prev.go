// cmd/next_prev.go
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"ytpl/internal/player"
	"ytpl/internal/state"

	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Play the next song in the current playlist or shuffled queue",
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("Player is not running.")
			return
		}
		if appState.CurrentPlaylist == "" && len(appState.ShuffleQueue) == 0 {
			fmt.Println("No active playlist or shuffle queue to advance.")
			return
		}

		if appState.CurrentPlaylist != "" {
			err := player.Next(appState)
			if err != nil {
				log.Printf("Error sending next command to player: %v", err)
			}
			time.Sleep(200 * time.Millisecond) // Short delay to allow mpv to update state
			// No direct display here. statusCmd.Run() will handle it.
		} else if len(appState.ShuffleQueue) > 0 {
			if appState.LastPlayedTrackIndex+1 < len(appState.ShuffleQueue) {
				nextIndex := appState.LastPlayedTrackIndex + 1
				nextTrackID := appState.ShuffleQueue[nextIndex]

				nextFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", nextTrackID))
				if _, err := os.Stat(nextFilePath); os.IsNotExist(err) {
					fmt.Printf("Warning: Next shuffled track \"%s\" is not locally stocked. Skipping.\n", nextTrackID) // Use ID as fallback here
					return
				}

				if err := player.LoadFile(appState, nextFilePath); err != nil {
					log.Fatalf("Error loading next shuffled track: %v", err)
				}

				// Update appState based on new track info for display
				appState.CurrentTrackID = nextTrackID
				appState.DownloadedFilePath = nextFilePath
				appState.LastPlayedTrackIndex = nextIndex
				appState.IsPlaying = true // Assuming playback starts
				if err := state.SaveState(); err != nil {
					log.Printf("Error saving state: %v", err)
				}
				time.Sleep(200 * time.Millisecond) // Short delay to allow mpv to update state
				// No direct display here. statusCmd.Run() will handle it.
			} else {
				fmt.Println("End of shuffle queue. No more songs.")
				player.StopPlayer(appState) // Stop player at end of queue
				return
			}
		} else {
			fmt.Println("No next song available.")
			return
		}
		statusCmd.Run(statusCmd, []string{}) // Call status command
	},
}

var prevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Play the previous song in the current playlist or shuffled queue",
	Run: func(cmd *cobra.Command, args []string) {
		if appState.PID == 0 {
			fmt.Println("Player is not running.")
			return
		}
		if appState.CurrentPlaylist == "" && len(appState.ShuffleQueue) == 0 {
			fmt.Println("No active playlist or shuffle queue to go back.")
			return
		}

		if appState.CurrentPlaylist != "" {
			err := player.Prev(appState)
			if err != nil {
				log.Printf("Error sending prev command to player: %v", err)
			}
			time.Sleep(200 * time.Millisecond) // Short delay
			// No direct display here. statusCmd.Run() will handle it.
		} else if len(appState.ShuffleQueue) > 0 {
			if appState.LastPlayedTrackIndex-1 >= 0 {
				prevIndex := appState.LastPlayedTrackIndex - 1
				prevTrackID := appState.ShuffleQueue[prevIndex]

				prevFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", prevTrackID))
				if _, err := os.Stat(prevFilePath); os.IsNotExist(err) {
					fmt.Printf("Warning: Previous shuffled track \"%s\" is not locally stocked. Skipping.\n", prevTrackID) // Use ID as fallback
					return
				}

				if err := player.LoadFile(appState, prevFilePath); err != nil {
					log.Fatalf("Error loading previous shuffled track: %v", err)
				}

				appState.CurrentTrackID = prevTrackID
				appState.DownloadedFilePath = prevFilePath
				appState.LastPlayedTrackIndex = prevIndex
				appState.IsPlaying = true
				if err := state.SaveState(); err != nil {
					log.Printf("Error saving state: %v", err)
				}
				time.Sleep(200 * time.Millisecond) // Short delay
				// No direct display here. statusCmd.Run() will handle it.
			} else {
				fmt.Println("Beginning of shuffle queue. No previous songs.")
				return
			}
		} else {
			fmt.Println("No previous song available.")
		}
		statusCmd.Run(statusCmd, []string{}) // Call status command
	},
}

// updateAppStateFromMpvStatusAndDisplay is no longer defined here.
// Its logic has been moved to updateAppStateFromMpvStatus in status.go.
// The commented-out block below should be completely removed from the file.
// (Ensure there are no trailing curly braces or comments that create syntax errors after this)///
