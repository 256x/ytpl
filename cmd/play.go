// cmd/play.go
package cmd

import (
	"fmt"
	"log"
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
			log.Fatalf("failed to initialize track manager: %v", err)
		}

		// Get all tracks from the track manager and sort by title
		trackList := trackManager.ListTracks()
		if len(trackList) == 0 {
			log.Fatal("no tracks found in .tracks file. Please run 'rebuild' command first.")
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

		for i, track := range trackList {
			trackPath := filepath.Join(cfg.DownloadDir, track.ID+".mp3")

			// Check if the file exists
			if _, err := os.Stat(trackPath); os.IsNotExist(err) {
				log.Printf("warning: track file not found: %s", trackPath)
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
			displayText := fmt.Sprintf("%02d:[%s] - %s", i+1, durationStr, displayTitle)

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
				fmt.Printf("\n- no local songs found matching \"%s\".\n", filterQuery)
			} else {
				fmt.Println("\n- no local songs found. use 'ytpl search' to download some.\n")
			}
			return
		}

		// Initialize fzf
		f, err := fuzzyfinder.New(fuzzyfinder.WithPrompt("[ play ] > "))
		if err != nil {
			log.Fatalf("Error initializing fzf: %v", err)
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
			log.Fatalf("Error running fzf: %v", err)
		}

		// Get the selected track
		if len(idxs) == 0 {
			log.Fatalf("no track selected")
		}
		selectedItem := displayItems[idxs[0]]
		if err != nil {
			if strings.Contains(err.Error(), "cancelled") {
				fmt.Println("\n- selection cancelled.\n")
				return
			}
			log.Fatalf("error selecting track: %v", err)
		}

		if err := player.StartPlayer(cfg, appState, selectedItem.Path); err != nil {
			log.Fatalf("error starting player: %v", err)
		}

		appState.CurrentTrackID = selectedItem.Info.ID
		appState.CurrentTrackTitle = selectedItem.DisplayTitle // Store the selected display title
		appState.CurrentTrackDuration = selectedItem.Info.Duration
		appState.DownloadedFilePath = selectedItem.Path
		appState.IsPlaying = true
		appState.CurrentPlaylist = ""
		if err := state.SaveState(); err != nil {
			log.Printf("error saving state: %v", err)
		}

		// Show the status in the same format as the status command
		ShowStatus()
	},
}

// init initializes the play command
func init() {
	rootCmd.AddCommand(playCmd)
}
