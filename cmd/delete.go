// cmd/delete.go
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ytpl/internal/playlist"
	"ytpl/internal/tracks"
	"ytpl/internal/util"
	"ytpl/internal/yt"

	fuzzyfinder "github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
)

type trackItem struct {
	Info         *yt.TrackInfo
	Path         string
	DisplayTitle string
	DisplayText  string
}

var delCmd = &cobra.Command{
	Use:   "del [query]",
	Short: "Delete a downloaded track",
	Args:  cobra.MaximumNArgs(1), // Optional query for filtering
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize track manager
		trackManager, err := tracks.NewManager(filepath.Dir(cfg.DownloadDir), cfg.DownloadDir)
		if err != nil {
			log.Fatalf("failed to initialize track manager: %v", err)
		}

		// Get all tracks from the track manager and sort by title
		trackList := trackManager.ListTracks()
		if len(trackList) == 0 {
			log.Fatal("no tracks found in .tracks file. Please run 'rebuild' command first.")
		}

		// Sort by title (case-insensitive)
		sort.Slice(trackList, func(i, j int) bool {
			return strings.ToLower(trackList[i].Title) < strings.ToLower(trackList[j].Title)
		})

		var selectableTracks []trackItem
		for i, track := range trackList {
			trackPath := filepath.Join(cfg.DownloadDir, track.ID+".mp3")

			// Skip if the file doesn't exist
			if _, err := os.Stat(trackPath); os.IsNotExist(err) {
				log.Printf("warning: track file not found: %s", trackPath)
				continue
			}

			durationStr := strings.Trim(util.FormatDuration(track.Duration), "[]")
			displayText := fmt.Sprintf("%02d:[%s] - %s", i+1, durationStr, track.Title)

			// Create a copy of the track to avoid referencing loop variable
			trackCopy := track
			selectableTracks = append(selectableTracks, trackItem{
				Info:         &trackCopy,
				Path:         trackPath,
				DisplayTitle: track.Title,
				DisplayText:  displayText,
			})
		}

		// Show warning message
		fmt.Println(util.Red("⚠️  warning: this will permanently delete the selected track."))
		fmt.Println()

		// Initialize fzf
		f, err := fuzzyfinder.New(
			fuzzyfinder.WithPrompt("[ delete ] > "),
		)
		if err != nil {
			log.Fatalf("Error initializing fzf: %v", err)
		}

		// Show fzf prompt
		idxs, err := f.Find(
			selectableTracks,
			func(i int) string {
				return selectableTracks[i].DisplayText
			},
		)

		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println(util.Yellow("\n- deletion cancelled\n"))
				return
			}
			log.Fatalf(util.Error("Error selecting track: %v"), err)
		}

		if len(idxs) == 0 {
			fmt.Println(util.Yellow("\n- no track selected\n"))
			return
		}

		selected := selectableTracks[idxs[0]]

		// Show confirmation
		confirmMsg := fmt.Sprintf("- are you sure you want to delete '%s'? [y/N] ", selected.Info.Title)
		confirmed, err := util.Confirm(confirmMsg)
		if err != nil {
			log.Fatalf("Error getting confirmation: %v", err)
		}
		if !confirmed {
			fmt.Println("\n- deletion cancelled\n")
			return
		}

		// Re-initialize track manager for removal
		trackManager, err = tracks.NewManager(filepath.Dir(cfg.DownloadDir), cfg.DownloadDir)
		if err != nil {
			log.Printf("warning: failed to initialize track manager: %v", err)
		}

		// Remove from .tracks first
		var removeTrackErr error
		if trackManager != nil {
			if err := trackManager.RemoveTrack(selected.Info.ID); err != nil {
				log.Printf("warning: failed to remove track from .tracks: %v", err)
				removeTrackErr = err
			} else {
				fmt.Println(util.Green("\n- removed from track library"))
			}
		}

		// Release trackManager lock before removing from playlists
		trackManager = nil

		// Remove from all playlists
		removedFromPlaylists, err := playlist.RemoveTrackFromAllPlaylists(selected.Info.ID)
		if err != nil {
			log.Printf("warning: failed to remove track from playlists: %v", err)
		} else if len(removedFromPlaylists) > 0 {
			fmt.Println(util.Green("\n- removed from playlists:"))
			for _, plName := range removedFromPlaylists {
				fmt.Printf("  - %s\n", plName)
			}
		}

		// If there was an error removing from .tracks, return early
		if removeTrackErr != nil {
			log.Fatalf("failed to remove track: %v", removeTrackErr)
		}

		// Delete files
		filesToDelete := []string{
			selected.Path,
			strings.TrimSuffix(selected.Path, ".mp3") + ".info.json",
			strings.TrimSuffix(selected.Path, ".mp3") + ".jpg",
		}

		var deletedFiles []string
		for _, file := range filesToDelete {
			if _, err := os.Stat(file); err == nil {
				if err := os.Remove(file); err != nil {
					log.Printf("error deleting %s: %v", file, err)
				} else {
					deletedFiles = append(deletedFiles, file)
				}
			}
		}

		// Playlist removal is already handled above

		// Show result
		if len(deletedFiles) > 0 {
			fmt.Println(util.Green("\n- deleted:"))
			for _, file := range deletedFiles {
				fmt.Printf("  - %s\n", filepath.Base(file))
			}
		}
	},
}
