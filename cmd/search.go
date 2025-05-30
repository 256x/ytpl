// cmd/search.go
package cmd

import (
	"fmt"
	"log"
	"strings"

	"ytpl/internal/player"
	"ytpl/internal/state"
	"ytpl/internal/util"
	"ytpl/internal/yt"

	fuzzyfinder "github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search YouTube for music",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		// Show searching spinner with query
		searchSpinner := util.StartSpinner(fmt.Sprintf("Searching '%s'...", query))
		tracks, err := yt.SearchYouTube(cfg, query)
		util.StopSpinner(searchSpinner)
		if err != nil {
			log.Fatalf("Error searching YouTube: %v", err)
		}

		if len(tracks) == 0 {
			fmt.Println("No results found.")
			return
		}

		// Initialize fzf
		f, err := fuzzyfinder.New()
		if err != nil {
			log.Fatalf("Error initializing fzf: %v", err)
		}

		// Show fzf prompt with basic info
		idxs, err := f.Find(
			tracks,
			func(i int) string {
				// Format index with leading zeros (e.g., 01, 02, ..., 10, 11, ...)
				indexStr := fmt.Sprintf("%02d", i+1)
				durationStr := strings.Trim(util.FormatDuration(tracks[i].Duration), "[]")
				return fmt.Sprintf("%s:[%s] - %s", indexStr, durationStr, tracks[i].Title)
			},
		)
		
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Search cancelled.")
				return
			}
			log.Fatalf("Error running fzf: %v", err)
		}

		// Get the selected track
		if len(idxs) == 0 {
			log.Fatalf("No track selected")
		}
		selectedTrack := tracks[idxs[0]]

		downloadSpinner := util.StartSpinner(fmt.Sprintf("Downloading '%s'...", selectedTrack.Title))
		downloadedFilePath, finalTrackInfo, err := yt.DownloadTrack(cfg, selectedTrack.ID)
		util.StopSpinner(downloadSpinner)

		if err != nil {
			log.Fatalf("Error downloading track: %v", err)
		}

		if err := player.StartPlayer(cfg, appState, downloadedFilePath); err != nil {
			log.Fatalf("Error starting player: %v", err)
		}

		appState.CurrentTrackID = finalTrackInfo.ID
		appState.CurrentTrackTitle = finalTrackInfo.Title
		appState.DownloadedFilePath = downloadedFilePath
		appState.IsPlaying = true
		appState.CurrentPlaylist = ""

		if err := state.SaveState(); err != nil {
			log.Printf("Error saving state: %v", err)
		}

		// Show status after starting player
		ShowStatus()
	},
}