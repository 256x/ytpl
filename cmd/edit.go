package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ytpl/internal/playertags"
	"ytpl/internal/tracks"
	"ytpl/internal/util"
	"ytpl/internal/yt"

	fuzzyfinder "github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [query]",
	Short: "Edit track metadata",
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

		for _, track := range trackList {
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
				fmt.Printf("\n- no local songs found matching \"%s\".\n", filterQuery)
			} else {
				fmt.Println("\n- no local songs found. use 'ytpl search' to download some.\n")
			}
			return
		}

		// Initialize fzf
		f, err := fuzzyfinder.New(fuzzyfinder.WithPrompt("[ edit ] > "))
		if err != nil {
			log.Fatalf("Error initializing fuzzy finder: %v", err)
		}

		// Show track selection
		idxs, err := f.Find(displayItems, func(i int) string {
			return displayItems[i].DisplayText
		})

		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("\n- selection cancelled.\n")
				return
			}
			log.Fatalf("Error selecting track: %v", err)
		}

		if len(idxs) == 0 {
			log.Fatalf("no track selected")
		}

		selectedItem := displayItems[idxs[0]]
		currentTitle := selectedItem.Info.Title

		// Show current title and get new title
		fmt.Println()
		fmt.Printf("- current: %s\n", currentTitle)
		fmt.Printf("      new: ")

		// Read new title from user input
		reader := bufio.NewReader(os.Stdin)
		newTitle, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}

		// Trim whitespace and check if empty
		newTitle = strings.TrimSpace(newTitle)
		if newTitle == "" {
			fmt.Println("\nNo changes made.")
			return
		}

		// If title hasn't changed, return early
		if newTitle == currentTitle {
			fmt.Println("\nNo changes made.")
			return
		}

		// Update title if changed
		if newTitle != "" && newTitle != currentTitle {
			selectedItem.Info.Title = newTitle
			if err := trackManager.UpdateTrack(selectedItem.Info); err != nil {
				log.Fatalf("Error updating track: %v", err)
			}
			if err := trackManager.SaveAll(); err != nil {
				log.Fatalf("Error saving tracks: %v", err)
			}
			fmt.Println("\nTitle updated.\n")
		} else if newTitle == "" {
			fmt.Println("\nEdit cancelled.\n")
		} else {
			fmt.Println("\nNo changes made.\n")
		}
	},
}

// Note: editCmd is added to rootCmd in root.go
