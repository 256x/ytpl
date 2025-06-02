// cmd/next_prev.go
package cmd

import (
	"fmt"
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
			fmt.Println("\n- no active playlist or shuffle queue to advance.\n")
			return
		}

		if appState.CurrentPlaylist != "" {
			err := player.Next(appState)
			if err != nil {
				// Error sending next command to player
			}
			time.Sleep(200 * time.Millisecond) // Short delay to allow mpv to update state
			// No direct display here. statusCmd.Run() will handle it.
		} else if len(appState.ShuffleQueue) > 0 {
			if appState.LastPlayedTrackIndex+1 < len(appState.ShuffleQueue) {
				nextIndex := appState.LastPlayedTrackIndex + 1
				nextTrackID := appState.ShuffleQueue[nextIndex]

				nextFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", nextTrackID))
				if _, err := os.Stat(nextFilePath); os.IsNotExist(err) {
					return
				}

				if err := player.LoadFile(appState, nextFilePath); err != nil {
					fmt.Fprintf(os.Stderr, "Error loading next shuffled track: %v\n", err)
					os.Exit(1)
				}

				// Update appState based on new track info for display
				appState.CurrentTrackID = nextTrackID
				appState.DownloadedFilePath = nextFilePath
				appState.LastPlayedTrackIndex = nextIndex
				appState.IsPlaying = true // Assuming playback starts
				if err := state.SaveState(); err != nil {
					fmt.Fprintf(os.Stderr, "Error saving state: %v\n", err)
					os.Exit(1)
				}
				time.Sleep(200 * time.Millisecond) // Short delay to allow mpv to update state
				// No direct display here. statusCmd.Run() will handle it.
			} else {
				fmt.Println("\n- end of shuffle queue. no more songs.")
				player.StopPlayer(appState) // Stop player at end of queue
				return
			}
		} else {
			fmt.Println("\n- no next song available.\n")
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
			fmt.Println("\n- player is not running.\n")
			return
		}
		if appState.CurrentPlaylist == "" && len(appState.ShuffleQueue) == 0 {
			fmt.Println("\n- no active playlist or shuffle queue to go back.\n")
			return
		}

		if appState.CurrentPlaylist != "" {
			err := player.Prev(appState)
			if err != nil {
				// Error sending previous command to player
			}
			time.Sleep(200 * time.Millisecond) // Short delay
			// No direct display here. statusCmd.Run() will handle it.
		} else if len(appState.ShuffleQueue) > 0 {
			if appState.LastPlayedTrackIndex-1 >= 0 {
				prevIndex := appState.LastPlayedTrackIndex - 1
				prevTrackID := appState.ShuffleQueue[prevIndex]

				prevFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", prevTrackID))
				if _, err := os.Stat(prevFilePath); os.IsNotExist(err) {
					return
				}

				if err := player.LoadFile(appState, prevFilePath); err != nil {
					fmt.Fprintf(os.Stderr, "Error loading previous shuffled track: %v\n", err)
					os.Exit(1)
				}

				appState.CurrentTrackID = prevTrackID
				appState.DownloadedFilePath = prevFilePath
				appState.LastPlayedTrackIndex = prevIndex
				appState.IsPlaying = true
				_ = state.SaveState()
				time.Sleep(200 * time.Millisecond) // Short delay
				// No direct display here. statusCmd.Run() will handle it.
			} else {
				fmt.Println("\n- beginning of shuffle queue. no previous songs.\n")
				return
			}
		} else {
			fmt.Println("\n- no previous song available.\n")
		}
		statusCmd.Run(statusCmd, []string{}) // Call status command
	},
}

// updateAppStateFromMpvStatusAndDisplay is no longer defined here.
// Its logic has been moved to updateAppStateFromMpvStatus in status.go.
// The commented-out block below should be completely removed from the file.
// (Ensure there are no trailing curly braces or comments that create syntax errors after this)///
