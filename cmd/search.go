// cmd/search.go
package cmd

import (
	"fmt"
	"log"
	"path/filepath"
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
	Short: "search youtube for music",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		// Show searching spinner with query
		searchSpinner := util.StartSpinner(fmt.Sprintf("Searching '%s'...", query))
		tracks, err := yt.SearchYouTube(cfg, query)
		util.StopSpinner(searchSpinner)
		if err != nil {
			log.Fatalf("error searching youtube: %v", err)
		}

		if len(tracks) == 0 {
			fmt.Println("no results found.")
			return
		}

		// Initialize fzf
		f, err := fuzzyfinder.New()
		if err != nil {
			log.Fatalf("error initializing fzf: %v", err)
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
				fmt.Println("search cancelled.")
				return
			}
			log.Fatalf("error running fzf: %v", err)
		}

		// Get the selected track
		if len(idxs) == 0 {
			log.Fatalf("no track selected")
		}
		selectedTrack := tracks[idxs[0]]

		// Check if track already exists locally
		localTrackInfo, err := yt.GetLocalTrackInfo(cfg, selectedTrack.ID)
		var downloadedFilePath string
		var finalTrackInfo *yt.TrackInfo

		if err == nil {
			// Use existing local file
			log.Printf("using existing local track: %s", selectedTrack.Title)
			downloadedFilePath = filepath.Join(cfg.DownloadDir, selectedTrack.ID+".mp3")
			// Use the local track info but preserve the title from search results
			// as it might be more up-to-date
			finalTrackInfo = localTrackInfo
			finalTrackInfo.Title = selectedTrack.Title
		} else {
			// Download the track if not found locally
			downloadSpinner := util.StartSpinner(fmt.Sprintf("Downloading '%s'...", selectedTrack.Title))
			downloadedFilePath, finalTrackInfo, err = yt.DownloadTrack(cfg, selectedTrack.ID)
			util.StopSpinner(downloadSpinner)

			if err != nil {
				log.Fatalf("error downloading track: %v", err)
			}
		}

		if err := player.StartPlayer(cfg, appState, downloadedFilePath); err != nil {
			log.Fatalf("error starting player: %v", err)
		}

		appState.CurrentTrackID = finalTrackInfo.ID
		appState.CurrentTrackTitle = finalTrackInfo.Title
		appState.DownloadedFilePath = downloadedFilePath
		appState.IsPlaying = true
		appState.CurrentPlaylist = ""

		if err := state.SaveState(); err != nil {
			log.Printf("error saving state: %v", err)
		}

		// Show status after starting player
		ShowStatus()
	},
}