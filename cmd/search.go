// cmd/search.go
package cmd

import (
	"fmt"
	"log"
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

func init() {
	log.SetOutput(os.Stderr)
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search YouTube for music",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		// Show searching spinner with query
		sanitizedQuery := strings.ReplaceAll(query, "\n", " ")
		// Use line style for search
		fmt.Println()
		searchSpinner := util.NewSpinnerWithStyle(
			fmt.Sprintf("searching '%s'...", sanitizedQuery),
			util.StyleLine,
		)
		tracks, err := yt.SearchYouTube(cfg, query)
		searchSpinner.Stop("")
		fmt.Println()
		if err != nil {
			log.Fatalf("error searching youtube: %v", err)
		}

		if len(tracks) == 0 {
			fmt.Println("\n- no results found.\n")
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
				fmt.Println("\n- search cancelled.\n")
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
			sanitizedTitle := strings.ReplaceAll(selectedTrack.Title, "\n", " ")
			// Use line style for download
			fmt.Println()
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
				log.Fatalf("error downloading track: %v", err)
			}
		}

		// Debug log for DownloadDir
		log.Printf("DownloadDir: %s", cfg.DownloadDir)

		// Initialize track manager
		// Use the parent of DownloadDir as the base directory for tracks
		tracksDir := filepath.Dir(cfg.DownloadDir)
		trackManager, err := trackpkg.NewManager("", tracksDir)
		if err != nil {
			log.Printf("warning: failed to initialize track manager: %v", err)
		} else {
			// Add downloaded track to the library
			if err := trackManager.AddTrack(*finalTrackInfo); err != nil {
				log.Printf("warning: failed to add track to library: %v", err)
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
			log.Printf("\n- error saving state: %v\n", err)
		}

		// Show status after starting player
		ShowStatus()
	},
}
