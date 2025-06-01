// cmd/shuffle.go
package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ytpl/internal/player"
	"ytpl/internal/state"
	"ytpl/internal/yt"

	"github.com/spf13/cobra"
)

var shuffleCmd = &cobra.Command{
	Use:   "shuffle",
	Short: "Shuffle and play all local stocked songs",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().UnixNano()) // Initialize random seed for different results each run

		// Shuffle operation starts, no need to show a message
		files, err := os.ReadDir(cfg.DownloadDir)
		if err != nil {
			log.Fatalf("error reading download directory %s: %v", cfg.DownloadDir, err)
		}

		var tracksToShuffle []yt.TrackInfo // Use yt.TrackInfo for consistency
		var filePaths []string

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if strings.HasSuffix(file.Name(), ".mp3") {
				id := strings.TrimSuffix(file.Name(), ".mp3")
				trackPath := filepath.Join(cfg.DownloadDir, file.Name())

				trackInfo, err := yt.GetLocalTrackInfo(cfg, id)
				if err != nil {
					// Fallback if info.json is missing or corrupted
					trackInfo = &yt.TrackInfo{ID: id, Title: strings.TrimSuffix(file.Name(), ".mp3")}
					fmt.Fprintf(os.Stderr, "warning: could not read info.json for %s: %v. using filename as title.\n", file.Name(), err)
				}
				tracksToShuffle = append(tracksToShuffle, *trackInfo)
				filePaths = append(filePaths, trackPath)
			}
		}

		if len(filePaths) == 0 {
			fmt.Println("\n- no local songs to shuffle. use 'ytpl search' to download some.\n")
			return
		}

		// Shuffle the file paths and corresponding track infos to keep them in sync
		rand.Shuffle(len(filePaths), func(i, j int) {
			filePaths[i], filePaths[j] = filePaths[j], filePaths[i]
			tracksToShuffle[i], tracksToShuffle[j] = tracksToShuffle[j], tracksToShuffle[i]
		})

		// Load the shuffled all-songs playlist into mpv
		if err := player.LoadPlaylistIntoPlayer(cfg, appState, filePaths, 0); err != nil { // Start from index 0
			log.Fatalf("error loading shuffled global playlist into player: %v", err)
		}

		// Update state for the first track in the shuffled global playlist
		firstTrack := tracksToShuffle[0] // Get info for the actual first track after shuffle
		appState.CurrentTrackID = firstTrack.ID
		appState.CurrentTrackTitle = firstTrack.Title // Will be formatted by status/next/prev display
		appState.DownloadedFilePath = filePaths[0]
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
