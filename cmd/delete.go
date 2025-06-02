// cmd/delete.go
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"ytpl/internal/playertags"
	"ytpl/internal/playlist"
	"ytpl/internal/util"
	"ytpl/internal/yt"

	fuzzyfinder "github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
)

type trackItem struct {
	*yt.TrackInfo
	Audio *playertags.AudioInfo
	Path  string
}

var delCmd = &cobra.Command{
	Use:   "del [query]",
	Short: "Delete a downloaded track",
	Args:  cobra.MaximumNArgs(1), // Optional query for filtering
	Run: func(cmd *cobra.Command, args []string) {
		// List all local tracks
		tracks, err := yt.ListLocalTracks(cfg)
		if err != nil {
			log.Fatalf("Error listing local tracks: %v", err)
		}

		var selectableTracks []trackItem
		for _, track := range tracks {
			// Get audio info
			audioInfo, _ := playertags.ReadTagsFromMP3(
				filepath.Join(cfg.DownloadDir, track.ID+".mp3"),
				"", "",
			)

			selectableTracks = append(selectableTracks, trackItem{
				TrackInfo: track,
				Path:      filepath.Join(cfg.DownloadDir, track.ID+".mp3"),
				Audio:     audioInfo,
			})
		}

		// Show warning message
		fmt.Println(util.Red("⚠️  warning: this will permanently delete the selected track."))
		fmt.Println(util.Yellow("   select a track to delete (ctrl+c to cancel):"))
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
				track := selectableTracks[i]
				durationStr := ""
				if track.Duration > 0 {
					durationStr = strings.Trim(util.FormatDuration(track.Duration), "[]")
				}

				artist := "Unknown Artist"
				if track.Audio != nil && track.Audio.Artist != "" {
					artist = track.Audio.Artist
				} else if track.Uploader != "" {
					artist = track.Uploader
				}

				title := "Unknown Title"
				if track.Title != "" {
					title = track.Title
				}

				display := fmt.Sprintf("%s - %s [%s]",
					title,
					artist,
					durationStr,
				)
				return display
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
		confirmMsg := fmt.Sprintf("- are you sure you want to delete '%s'? (y/N) ", selected.Title)
		confirmed, err := util.Confirm(confirmMsg)
		if err != nil {
			log.Fatalf("Error getting confirmation: %v", err)
		}
		if !confirmed {
			fmt.Println("\n- deletion cancelled\n")
			return
		}

		// Remove from playlists first
		var removedFromPlaylists []string
		playlists, err := playlist.ListAllPlaylists()
		if err != nil {
			log.Printf("warning: failed to list playlists: %v", err)
		} else {
			for _, plName := range playlists {
				err := playlist.RemoveTrack(plName, selected.ID)
				if err != nil && !strings.Contains(err.Error(), "not found") {
					log.Printf("warning: failed to remove track from playlist %s: %v", plName, err)
				} else if err == nil {
					removedFromPlaylists = append(removedFromPlaylists, plName)
				}
			}

			if len(removedFromPlaylists) > 0 {
				fmt.Println(util.Green("\n- removed from playlists:"))
				for _, plName := range removedFromPlaylists {
					fmt.Printf("  - %s\n", plName)
				}
			}
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

		// Remove from playlists
		// TODO: Implement playlist removal

		// Show result
		if len(deletedFiles) > 0 {
			fmt.Println(util.Green("\n- deleted:"))
			for _, file := range deletedFiles {
				fmt.Printf("  - %s\n", filepath.Base(file))
			}
		}
	},
}
