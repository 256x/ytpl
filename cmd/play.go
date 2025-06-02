// cmd/play.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ytpl/internal/player"
	"ytpl/internal/playertags"
	"ytpl/internal/state"
	"ytpl/internal/tracks"
	"ytpl/internal/util"
	"ytpl/internal/yt"

	fuzzyfinder "github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play [query]",
	Short: "Play a locally stocked song",
	Args:  cobra.MaximumNArgs(1), // Optional query for filtering
	Run: func(cmd *cobra.Command, args []string) {
		filterQuery := ""
		if len(args) > 0 {
			filterQuery = strings.ToLower(args[0])
		}

		// Initialize track manager
		trackManager, err := tracks.NewManager("", cfg.DownloadDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing track manager: %v\n", err)
			os.Exit(1)
		}

		// Get all tracks from the track manager and sort by title
		trackList := trackManager.ListTracks()
		if len(trackList) == 0 {
			fmt.Println("No tracks found. Please run 'rebuild' command first.")
			os.Exit(1)
		}

		sort.Slice(trackList, func(i, j int) bool {
			return strings.ToLower(trackList[i].Title) < strings.ToLower(trackList[j].Title)
		})

		// Format tracks for display
		displayItems := make([]struct {
			Info         *yt.TrackInfo
			Path         string
			Audio        *playertags.AudioInfo
			DisplayTitle string
			DisplayText  string
		}, 0, len(trackList))

		for _, track := range trackList {
			trackPath := filepath.Join(cfg.DownloadDir, track.ID+".mp3")

			// Check if the file exists
			if _, err := os.Stat(trackPath); os.IsNotExist(err) {
				// Track file not found
				continue
			}

			// Use track title from .tracks file
			displayTitle := track.Title

			// Skip if filter query doesn't match
			if filterQuery != "" {
				if !strings.Contains(strings.ToLower(displayTitle), strings.ToLower(filterQuery)) &&
					!strings.Contains(strings.ToLower(track.ID), strings.ToLower(filterQuery)) {
					continue
				}
			}

			durationStr := strings.Trim(util.FormatDuration(track.Duration), "[]")
			displayText := fmt.Sprintf("[%s] - %s", durationStr, displayTitle)
			displayItems = append(displayItems, struct {
				Info         *yt.TrackInfo
				Path         string
				Audio        *playertags.AudioInfo
				DisplayTitle string
				DisplayText  string
			}{
				Info:         &track,
				Path:         trackPath,
				Audio:        nil, // MP3タグは読み取らない
				DisplayTitle: displayTitle,
				DisplayText:  displayText,
			})
		}

		if len(displayItems) == 0 {
			if filterQuery != "" {
				fmt.Printf("No local songs found matching '%s'\n", filterQuery)
			} else {
				fmt.Println("No local songs found. Use 'ytpl search' to download some")
			}
			return
		}

		// Initialize fzf
		f, err := fuzzyfinder.New(fuzzyfinder.WithPrompt("[ play ] > "))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing fzf: %v\n", err)
			os.Exit(1)
		}

		// Show fzf prompt
		idxs, err := f.Find(
			displayItems,
			func(i int) string {
				return displayItems[i].DisplayText
			},
		)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("\n- selection cancelled.\n")
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
		selectedItem := displayItems[idxs[0]]

		if err := player.StartPlayer(cfg, appState, selectedItem.Path); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting player: %v\n", err)
			os.Exit(1)
		}

		appState.CurrentTrackID = selectedItem.Info.ID
		appState.CurrentTrackTitle = selectedItem.DisplayTitle // Store the selected display title
		appState.CurrentTrackDuration = selectedItem.Info.Duration
		appState.DownloadedFilePath = selectedItem.Path
		appState.IsPlaying = true
		appState.CurrentPlaylist = ""
		// Ignore error when saving state
		_ = state.SaveState()

		// Show the status in the same format as the status command
		ShowStatus()
	},
}

// init initializes the play command
// Note: playCmd is added to rootCmd in root.go
