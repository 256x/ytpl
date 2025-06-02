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
			log.Printf("warning: failed to initialize track manager: %v", err)
			// Fallback to directory scanning if track manager fails
			log.Println("falling back to directory scanning")
			fallbackToDirectoryScanning(filterQuery)
			return
		}

		// Get all tracks from the track manager
		trackList := trackManager.ListTracks()
		if len(trackList) == 0 {
			log.Println("no tracks found in .tracks file, falling back to directory scanning")
			fallbackToDirectoryScanning(filterQuery)
			return
		}

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

			// Get MP3 tags for additional metadata
			audioInfo, _ := playertags.ReadTagsFromMP3(trackPath, "", "")

			// Determine the best display title
			displayTitle := track.Title
			if audioInfo != nil && audioInfo.Title != "" && !strings.Contains(audioInfo.Title, "Unknown Title") {
				displayTitle = audioInfo.Title
			}

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
				Audio:        audioInfo,
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

func fallbackToDirectoryScanning(filterQuery string) {
	// This is the original directory scanning code
	files, err := os.ReadDir(cfg.DownloadDir)
	if err != nil {
		log.Fatalf("error reading download directory %s: %v", cfg.DownloadDir, err)
	}

	var selectablePaths []struct {
		Info         *yt.TrackInfo
		Path         string
		Audio        *playertags.AudioInfo
		DisplayTitle string
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".mp3") {
			id := strings.TrimSuffix(file.Name(), ".mp3")
			trackPath := filepath.Join(cfg.DownloadDir, file.Name())

			var currentDisplayTitle string

			ytTrackInfo, ytInfoErr := yt.GetLocalTrackInfo(cfg, id)

			audioInfo, readTagErr := playertags.ReadTagsFromMP3(trackPath, "", "")
			if readTagErr != nil {
				log.Printf("warning: failed to read MP3 tags for %s: %v.\n", file.Name(), readTagErr)
			}

			if audioInfo != nil && audioInfo.Title != "" && !strings.Contains(audioInfo.Title, "Unknown Title") {
				currentDisplayTitle = audioInfo.Title
			} else if ytInfoErr == nil && ytTrackInfo != nil && ytTrackInfo.Title != "" {
				currentDisplayTitle = ytTrackInfo.Title
			} else {
				currentDisplayTitle = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			}

			var ytTrackInfoForID *yt.TrackInfo
			if ytInfoErr == nil && ytTrackInfo != nil {
				ytTrackInfoForID = ytTrackInfo
			} else {
				ytTrackInfoForID = &yt.TrackInfo{ID: id, Title: currentDisplayTitle}
			}

			var filterArtist string
			if audioInfo != nil && audioInfo.Artist != "" && audioInfo.Artist != "Unknown Artist" {
				filterArtist = audioInfo.Artist
			} else if ytTrackInfo != nil && ytTrackInfo.Uploader != "" {
				filterArtist = ytTrackInfo.Uploader
			}

			if filterQuery == "" ||
				strings.Contains(strings.ToLower(currentDisplayTitle), filterQuery) ||
				strings.Contains(strings.ToLower(filterArtist), filterQuery) ||
				strings.Contains(strings.ToLower(ytTrackInfoForID.ID), filterQuery) {

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
			fmt.Printf("\n- no local songs found matching \"%s\".\n", filterQuery)
		} else {
			fmt.Println("\n- no local songs found. use 'ytpl search' to download some.\n")
		}
		return
	}

	displayItems := make([]struct {
		Info         *yt.TrackInfo
		Path         string
		Audio        *playertags.AudioInfo
		DisplayTitle string
		DisplayText  string
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
	f, err := fuzzyfinder.New(fuzzyfinder.WithPrompt("[ play ] "))
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

	if err := player.StartPlayer(cfg, appState, selectedItem.Path); err != nil {
		log.Fatalf("error starting player: %v", err)
	}

	appState.CurrentTrackID = selectedItem.Info.ID
	appState.CurrentTrackTitle = selectedItem.DisplayTitle
	appState.CurrentTrackDuration = selectedItem.Info.Duration
	appState.DownloadedFilePath = selectedItem.Path
	appState.IsPlaying = true
	appState.CurrentPlaylist = ""
	if err := state.SaveState(); err != nil {
		log.Printf("error saving state: %v", err)
	}

	// Show the status in the same format as the status command
	ShowStatus()
}

func init() {
	rootCmd.AddCommand(playCmd)
}
