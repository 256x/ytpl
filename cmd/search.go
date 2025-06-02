// cmd/search.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ytpl/internal/player"
	"ytpl/internal/state"
	trackpkg "ytpl/internal/tracks"
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
		sanitizedQuery := strings.ReplaceAll(query, "\n", " ")
		// Use line style for search
		searchSpinner := util.NewSpinnerWithStyle(
			fmt.Sprintf("searching '%s'...", sanitizedQuery),
			util.StyleLine,
		)
		tracks, err := yt.SearchYouTube(cfg, query)
		searchSpinner.Stop("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching YouTube: %v\n", err)
			os.Exit(1)
		}

		if len(tracks) == 0 {
			fmt.Println("\n- no results found.\n")
			return
		}

		// Initialize fzf
		f, err := fuzzyfinder.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing fzf: %v\n", err)
			os.Exit(1)
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
				fmt.Println("\n- search cancelled.\n")
				return
			}
			fmt.Fprintf(os.Stderr, "Error running fzf: %v\n", err)
			os.Exit(1)
		}

		// Get the selected track
		if len(idxs) == 0 {
			fmt.Fprintln(os.Stderr, "No track selected")
			os.Exit(1)
		}
		selectedTrack := tracks[idxs[0]]

		// Check if track already exists locally
		localTrackInfo, err := yt.GetLocalTrackInfo(cfg, selectedTrack.ID)
		var downloadedFilePath string
		var finalTrackInfo *yt.TrackInfo

		if err == nil {
			// Use existing local file
			downloadedFilePath = filepath.Join(cfg.DownloadDir, selectedTrack.ID+".mp3")
			// Use the local track info but preserve the title from search results
			// as it might be more up-to-date
			finalTrackInfo = localTrackInfo
			finalTrackInfo.Title = selectedTrack.Title
		} else {
			// Download the track if not found locally
			sanitizedTitle := strings.ReplaceAll(selectedTrack.Title, "\n", " ")
			// Use line style for download
			downloadSpinner := util.NewSpinnerWithStyle(
				fmt.Sprintf("downloading '%s'...", sanitizedTitle),
				util.StyleLine,
			)
			downloadedFilePath, finalTrackInfo, err = yt.DownloadTrack(cfg, selectedTrack.ID)
			if err != nil {
				downloadSpinner.StopWithError(fmt.Sprintf("failed to download '%s'", sanitizedTitle))
			} else {
				downloadSpinner.Stop("")
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading track: %v\n", err)
				os.Exit(1)
			}
		}

		// Initialize track manager
		tracksDir := filepath.Dir(cfg.DownloadDir)
		trackManager, err := trackpkg.NewManager("", tracksDir)
		if err == nil {
			// Add downloaded track to the library
			_ = trackManager.AddTrack(*finalTrackInfo)
		}

		if err := player.StartPlayer(cfg, appState, downloadedFilePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting player: %v\n", err)
			os.Exit(1)
		}

		appState.CurrentTrackID = finalTrackInfo.ID
		appState.CurrentTrackTitle = finalTrackInfo.Title
		appState.DownloadedFilePath = downloadedFilePath
		appState.IsPlaying = true
		appState.CurrentPlaylist = ""

		// Ignore error when saving state
		_ = state.SaveState()

		// Show status after starting player
		ShowStatus()
	},
}
