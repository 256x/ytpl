// cmd/shuffle.go
package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	"ytpl/internal/player"
	"ytpl/internal/state"
	"ytpl/internal/tracks"

	"github.com/spf13/cobra"
)

// trackInfo represents minimal track information needed for shuffling
type trackInfo struct {
	ID    string
	Title string
	Path  string
}

var shuffleCmd = &cobra.Command{
	Use:   "shuffle",
	Short: "Shuffle and play all local stocked songs",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().UnixNano()) // Initialize random seed for different results each run

		// Initialize track manager
		trackManager, err := tracks.NewManager("", cfg.DownloadDir)
		if err != nil {
			log.Fatalf("error initializing track manager: %v", err)
		}

		// Get all tracks from the manager
		allTracks := trackManager.ListTracks()
		if len(allTracks) == 0 {
			fmt.Println("\n- no local songs to shuffle. use 'ytpl search' to download some.\n")
			return
		}

		// Convert to our minimal trackInfo format
		tracksToShuffle := make([]trackInfo, 0, len(allTracks))
		for _, track := range allTracks {
			tracksToShuffle = append(tracksToShuffle, trackInfo{
				ID:    track.ID,
				Title: track.Title,
				Path:  filepath.Join(cfg.DownloadDir, track.ID+".mp3"),
			})
		}

		// Shuffle the tracks
		rand.Shuffle(len(tracksToShuffle), func(i, j int) {
			tracksToShuffle[i], tracksToShuffle[j] = tracksToShuffle[j], tracksToShuffle[i]
		})

		// Extract file paths for the player
		filePaths := make([]string, len(tracksToShuffle))
		for i, track := range tracksToShuffle {
			filePaths[i] = track.Path
		}

		// Load the shuffled all-songs playlist into mpv
		if err := player.LoadPlaylistIntoPlayer(cfg, appState, filePaths, 0); err != nil { // Start from index 0
			log.Fatalf("error loading shuffled global playlist into player: %v", err)
		}

		// Update state for the first track in the shuffled global playlist
		firstTrack := tracksToShuffle[0]
		appState.CurrentTrackID = firstTrack.ID
		appState.CurrentTrackTitle = firstTrack.Title
		appState.DownloadedFilePath = firstTrack.Path
		appState.IsPlaying = true
		appState.CurrentPlaylist = "all songs (shuffled)" // Special name for global shuffled playlist
		appState.LastPlayedTrackIndex = 0

		if err := state.SaveState(); err != nil {
			log.Printf("error saving state: %v", err)
		}

		// Show status without extra messages
		statusCmd.Run(statusCmd, []string{})
	},
}
