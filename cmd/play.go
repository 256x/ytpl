// cmd/play.go
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"ytpl/internal/player"
	"ytpl/internal/playertags"
	"ytpl/internal/state"
	"ytpl/internal/util"
	"ytpl/internal/yt"

	fuzzyfinder "github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play [query]",
	Short: "play a locally stocked song",
	Args:  cobra.MaximumNArgs(1), // Optional query for filtering
	Run: func(cmd *cobra.Command, args []string) {
		filterQuery := ""
		if len(args) > 0 {
			filterQuery = strings.ToLower(args[0])
		}

		files, err := os.ReadDir(cfg.DownloadDir)
		if err != nil {
			log.Fatalf("error reading download directory %s: %v", cfg.DownloadDir, err)
		}

		var selectablePaths []struct { // Struct to hold info and path for selection
			Info         *yt.TrackInfo
			Path         string
			Audio        *playertags.AudioInfo
			DisplayTitle string // The best title for selection prompt and final display
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if strings.HasSuffix(file.Name(), ".mp3") {
				id := strings.TrimSuffix(file.Name(), ".mp3")
				trackPath := filepath.Join(cfg.DownloadDir, file.Name())

				var currentDisplayTitle string // The best title to use for display

				// Fetch ytTrackInfo for its title as a fallback
				ytTrackInfo, ytInfoErr := yt.GetLocalTrackInfo(cfg, id)

				// Fetch MP3 tags
				audioInfo, readTagErr := playertags.ReadTagsFromMP3(trackPath, "", "") // No initial fallbacks for tags
				if readTagErr != nil {
					log.Printf("warning: failed to read mp3 tags for %s: %v.\n", file.Name(), readTagErr)
				}

				// --- NEW: Determine the best display title based on priority ---
				// Priority 1: MP3 tag Title (audioInfo.Title)
				if audioInfo != nil && audioInfo.Title != "" && !strings.Contains(audioInfo.Title, "Unknown Title") {
					currentDisplayTitle = audioInfo.Title
				} else if ytInfoErr == nil && ytTrackInfo != nil && ytTrackInfo.Title != "" {
					// Priority 2: YouTube JSON Title (ytTrackInfo.Title)
					currentDisplayTitle = ytTrackInfo.Title
				} else {
					// Priority 3: Fallback to filename
					currentDisplayTitle = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				}
				// --- END NEW ---

				// Ensure ytTrackInfoForID is valid for .Info.ID
				var ytTrackInfoForID *yt.TrackInfo
				if ytInfoErr == nil && ytTrackInfo != nil {
					ytTrackInfoForID = ytTrackInfo
				} else {
					ytTrackInfoForID = &yt.TrackInfo{ID: id, Title: currentDisplayTitle}
				}

				// For filtering, use the determined currentDisplayTitle and audioInfo.Artist (if available)
				var filterArtist string
				if audioInfo != nil && audioInfo.Artist != "" && audioInfo.Artist != "Unknown Artist" {
					filterArtist = audioInfo.Artist
				} else if ytTrackInfo != nil && ytTrackInfo.Uploader != "" {
					filterArtist = ytTrackInfo.Uploader
				}

				if filterQuery == "" ||
					strings.Contains(strings.ToLower(currentDisplayTitle), filterQuery) || // Filter by best display title
					strings.Contains(strings.ToLower(filterArtist), filterQuery) || // Filter by artist (best available)
					strings.Contains(strings.ToLower(ytTrackInfoForID.ID), filterQuery) { // Filter by YouTube ID

					selectablePaths = append(selectablePaths, struct {
						Info         *yt.TrackInfo
						Path         string
						Audio        *playertags.AudioInfo
						DisplayTitle string
					}{Info: ytTrackInfoForID, Path: trackPath, Audio: audioInfo, DisplayTitle: currentDisplayTitle})
				}
			}
		}

		if len(selectablePaths) == 0 {
			if filterQuery != "" {
				fmt.Printf("no local songs found matching \"%s\".\n", filterQuery)
			} else {
				fmt.Println("no local songs found. use 'ytpl search' to download some.")
			}
			return
		}

		// Format tracks for display
		displayItems := make([]struct {
			Info         *yt.TrackInfo
			Path         string
			Audio        *playertags.AudioInfo
			DisplayTitle string
			DisplayText  string // New field for formatted display text
		}, len(selectablePaths))

		for i, item := range selectablePaths {
			duration := 0.0
			if item.Info != nil {
				duration = item.Info.Duration
			}

			durationStr := strings.Trim(util.FormatDuration(duration), "[]")
			displayText := fmt.Sprintf("%02d:[%s] - %s", i+1, durationStr, item.DisplayTitle)

			displayItems[i] = struct {
				Info         *yt.TrackInfo
				Path         string
				Audio        *playertags.AudioInfo
				DisplayTitle string
				DisplayText  string
			}{
				Info:         item.Info,
				Path:         item.Path,
				Audio:        item.Audio,
				DisplayTitle: item.DisplayTitle,
				DisplayText:  displayText,
			}
		}

		// Initialize fzf
		f, err := fuzzyfinder.New(fuzzyfinder.WithPrompt("[ play from stock ] > "))
		if err != nil {
			log.Fatalf("error initializing fzf: %v", err)
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
				fmt.Println("selection cancelled.")
				return
			}
			log.Fatalf("error running fzf: %v", err)
		}

		// Get the selected track
		if len(idxs) == 0 {
			log.Fatalf("no track selected")
		}
		selectedItem := displayItems[idxs[0]]
		if err != nil {
			if strings.Contains(err.Error(), "cancelled") {
				fmt.Println("selection cancelled.")
				return
			}
			log.Fatalf("error selecting track: %v", err)
		}

		if err := player.StartPlayer(cfg, appState, selectedItem.Path); err != nil {
			log.Fatalf("error starting player: %v", err)
		}

		appState.CurrentTrackID = selectedItem.Info.ID
		appState.CurrentTrackTitle = selectedItem.DisplayTitle // Store the selected display title
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
