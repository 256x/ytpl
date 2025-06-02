// cmd/list.go
package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	fuzzyfinder "github.com/koki-develop/go-fzf"

	"ytpl/internal/player"
	"ytpl/internal/playlist"
	"ytpl/internal/state"
	"ytpl/internal/util"
	"ytpl/internal/yt"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Manage and play music playlists",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// if no arguments, show playlist selection prompt
			playlistNames, err := playlist.ListAllPlaylists()
			if err != nil {
				log.Fatalf("error loading playlist: %v", err)
			}

			if len(playlistNames) == 0 {
				fmt.Println("\n- no playlists found. use 'ytpl list create <name>' to create one.\n")
				return
			}

			selectedPlaylist, err := util.SelectFromList(
				"[ list ] ",
				playlistNames,
				func(name string) string { return name },
			)
			if err != nil {
				if strings.Contains(err.Error(), "cancelled") {
					fmt.Println("\n- playlist selection cancelled.\n")
					return
				}
				log.Fatalf("error selecting playlist: %v", err)
			}

			action, err := util.SelectFromList(
				"[ list action for '"+selectedPlaylist+"' ] ",
				[]string{"play (ordered)", "play (shuffled)", "show contents"},
				func(s string) string { return s },
			)
			if err != nil {
				if strings.Contains(err.Error(), "cancelled") {
					fmt.Println("\n- action cancelled.\n")
					return
				}
				log.Fatalf("error selecting action: %v", err)
			}

			switch action {
			case "play (ordered)":
				listPlayCmd.Run(listPlayCmd, []string{selectedPlaylist})
			case "play (shuffled)":
				listShuffleCmd.Run(listShuffleCmd, []string{selectedPlaylist})
			case "show contents":
				listShowCmd.Run(listShowCmd, []string{selectedPlaylist})
			default:
				fmt.Println("invalid action selected.")
			}
			// status will be shown by the respective command handlers
			return
		}
		// if arguments are provided, cobra's default behavior handles subcommand execution.
		cmd.Help()
	},
}

var listCreateCmd = &cobra.Command{
	Use:   "create <playlist_name>",
	Short: "create a new playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := playlist.MakePlaylist(name); err != nil {
			log.Fatalf("error creating playlist '%s': %v", name, err)
		}
		fmt.Printf("\n- playlist '%s' created successfully.\n\n", name)
	},
}

var listAddCmd = &cobra.Command{
	Use:   "add <playlist_name>",
	Short: "add the currently playing song to a playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		playlistName := args[0]

		if appState.CurrentTrackID == "" {
			fmt.Println("\n- no song is currently playing to add to a playlist.\n")
			return
		}

		trackToAdd := playlist.TrackInfo{
			ID:       appState.CurrentTrackID,
			Title:    appState.CurrentTrackTitle,
			Duration: appState.CurrentTrackDuration,
		}

		err := playlist.AddTrack(playlistName, trackToAdd)
		if err != nil {
			if strings.Contains(err.Error(), "already exists in playlist") {
				fmt.Printf("\n- track '%s' is already in playlist '%s'\n", appState.CurrentTrackTitle, playlistName)
			} else {
				log.Fatalf("error adding song to playlist '%s': %v", playlistName, err)
			}
		} else {
			fmt.Printf("\n- added \"%s\" to playlist '%s'\n", appState.CurrentTrackTitle, playlistName)
		}
	},
}

var listRemoveCmd = &cobra.Command{
	Use:   "remove <playlist_name>",
	Short: "remove the currently playing song from a playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		playlistName := args[0]

		if appState.CurrentTrackID == "" {
			fmt.Println("\n- no song is currently playing to remove from a playlist.\n")
			return
		}

		if err := playlist.RemoveTrack(playlistName, appState.CurrentTrackID); err != nil {
			log.Fatalf("error removing track from playlist: %v", err)
		}
		fmt.Printf("\n- removed \"%s\" from playlist '%s'.\n", appState.CurrentTrackTitle, playlistName)
	},
}

var listDelCmd = &cobra.Command{
	Use:   "del <playlist_name>",
	Short: "Delete a playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		confirm, err := util.Confirm(fmt.Sprintf("\n- are you sure you want to delete playlist '%s'? (y/N) ", name))
		if err != nil || !confirm {
			fmt.Println("\n- playlist deletion cancelled.\n")
			return
		}

		if err := playlist.DeletePlaylist(name); err != nil {
			log.Fatalf("Error deleting playlist '%s': %v", name, err)
		}
		fmt.Printf("\n- playlist '%s' deleted successfully.\n", name)
	},
}

var listShowCmd = &cobra.Command{
	Use:   "show <playlist_name>",
	Short: "Show contents of a playlist with fzf interface",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		p, err := playlist.LoadPlaylist(name)
		if err != nil {
			if os.IsNotExist(err) || strings.Contains(err.Error(), "does not exist") {
				fmt.Printf("\n- playlist '%s' does not exist.\n\n", name)
				return
			}
			log.Fatalf("error loading playlist '%s': %v", name, err)
		}

		if len(p.Tracks) == 0 {
			fmt.Printf("\n- playlist '%s' is empty.\n\n", name)
			return
		}

		// Format tracks for display with track numbers and duration
		displayItems := make([]struct {
			Track       playlist.TrackInfo
			DisplayText string
		}, len(p.Tracks))

		for i, track := range p.Tracks {
		displayText := fmt.Sprintf("%02d: %s", i+1, track.Title)

		displayItems[i] = struct {
			Track       playlist.TrackInfo
			DisplayText string
		}{
			Track:       track,
			DisplayText: displayText,
		}
	}


		// Create fzf instance with custom prompt
		prompt := fmt.Sprintf("[ play from %s ]", name)
		f, err := fuzzyfinder.New(fuzzyfinder.WithPrompt(prompt + " > "))
		if err != nil {
			log.Fatalf("error initializing fzf: %v", err)
		}

		// Get user selection
		idxs, err := f.Find(
			displayItems,
			func(i int) string {
				return displayItems[i].DisplayText
			},
		)

		if err != nil {
			if strings.Contains(err.Error(), "cancelled") {
				fmt.Println("\n- selection cancelled.\n")
				return
			}
			log.Fatalf("error selecting track: %v", err)
		}

		if len(idxs) == 0 {
			fmt.Println("\n- no track selected.\n")
			return
		}

		// Play the selected track
		selected := displayItems[idxs[0]]
		trackPath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", selected.Track.ID))

		// Start playing the selected track
		if err := player.StartPlayer(cfg, appState, trackPath); err != nil {
			log.Fatalf("error playing track: %v", err)
		}

		// Update the current track info in the app state
		appState.CurrentTrackID = selected.Track.ID
		appState.CurrentTrackTitle = selected.Track.Title
		appState.IsPlaying = true
		if err := state.SaveState(); err != nil {
			log.Printf("warning: failed to save state: %v", err)
		}

		// Show status after starting player
		statusCmd.Run(statusCmd, []string{})
	},
}

// listPlayCmd plays a playlist in order.
var listPlayCmd = &cobra.Command{
	Use:   "play <playlist_name>",
	Short: "Play songs from a playlist (ordered)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		playlistName := args[0]
		p, err := playlist.LoadPlaylist(playlistName)
		if err != nil {
			log.Fatalf("Error loading playlist '%s': %v", playlistName, err)
		}

		if len(p.Tracks) == 0 {
			fmt.Printf("\n- playlist '%s' is empty. nothing to play.\n", playlistName)
			return
		}

		// No extra empty line here, as per your request.

		var tracksToPlay []playlist.TrackInfo // Keep track info for state update
		var playlistFilePaths []string        // Paths to pass to mpv

		// Collect all file paths and ensure they are downloaded
		for i, track := range p.Tracks {
			downloadedFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", track.ID))
			if _, err := os.Stat(downloadedFilePath); os.IsNotExist(err) {
				fmt.Printf("\n- track %d/%d \"%s\" is not stocked locally. downloading...\n", i+1, len(p.Tracks), track.Title)
				stopSpinner := util.StartSpinner(fmt.Sprintf("\n- downloading \"%s\"", track.Title))
				_, _, downloadErr := yt.DownloadTrack(cfg, track.ID) // yt.DownloadTrack returns (filePath, TrackInfo, error)
				util.StopSpinner(stopSpinner)
				if downloadErr != nil {
					fmt.Printf("\n- warning: failed to download track \"%s\": %v. skipping from playlist.\n", track.Title, downloadErr)
					continue // Skip this track if download fails
				}
				fmt.Printf("\n- downloaded \"%s\" to %s.\n", track.Title, downloadedFilePath)
			}
			tracksToPlay = append(tracksToPlay, track)
			playlistFilePaths = append(playlistFilePaths, downloadedFilePath)
		}

		if len(playlistFilePaths) == 0 {
			fmt.Printf("\n- no playable songs found in playlist '%s' after checking local stock.\n", playlistName)
			return
		}

		// Load the entire playlist into mpv using LoadPlaylistIntoPlayer
		if err := player.LoadPlaylistIntoPlayer(cfg, appState, playlistFilePaths, 0); err != nil { // Start from index 0
			log.Fatalf("error loading playlist into player: %v", err)
		}

		// Update state for the first track in the playlist
		firstTrack := tracksToPlay[0]
		appState.CurrentTrackID = firstTrack.ID
		appState.CurrentTrackTitle = firstTrack.Title
		appState.DownloadedFilePath = playlistFilePaths[0]
		appState.IsPlaying = true
		appState.CurrentPlaylist = playlistName
		appState.LastPlayedTrackIndex = 0

		if err := state.SaveState(); err != nil {
			log.Printf("Error saving state: %v", err)
		}

		// Show status instead of custom message
		ShowStatus()
	},
}

// listShuffleCmd shuffles and plays a specific playlist.
var listShuffleCmd = &cobra.Command{
	Use:   "shuffle <playlist_name>",
	Short: "Shuffle and play songs from a playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Remove the "Preparing playlist..." message
		playlistName := args[0]
		p, err := playlist.LoadPlaylist(playlistName)
		if err != nil {
			log.Fatalf("Error loading playlist '%s': %v", playlistName, err)
		}

		if len(p.Tracks) == 0 {
			fmt.Printf("i\n- playlist '%s' is empty. nothing to shuffle and play.\n", playlistName)
			return
		}

		// No extra empty line here, as per your request.

		var tracksToPlay []playlist.TrackInfo
		var playlistFilePaths []string

		// Collect all file paths and ensure they are downloaded
		for i, track := range p.Tracks {
			downloadedFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", track.ID))
			if _, err := os.Stat(downloadedFilePath); os.IsNotExist(err) {
				fmt.Printf("\n- track %d/%d \"%s\" is not stocked locally. downloading...\n", i+1, len(p.Tracks), track.Title)
				stopSpinner := util.StartSpinner(fmt.Sprintf("\n- downloading \"%s\"", track.Title))
				_, _, downloadErr := yt.DownloadTrack(cfg, track.ID)
				util.StopSpinner(stopSpinner)
				if downloadErr != nil {
					fmt.Printf("\n- warning: failed to download track \"%s\": %v. skipping from playlist.\n", track.Title, downloadErr)
					continue
				}
				fmt.Printf("\n- downloaded \"%s\" to %s.\n", track.Title, downloadedFilePath)
			}
			tracksToPlay = append(tracksToPlay, track)
			playlistFilePaths = append(playlistFilePaths, downloadedFilePath)
		}

		if len(playlistFilePaths) == 0 {
			fmt.Printf("\n- no playable songs found in playlist '%s' after checking local stock.\n", playlistName)
			return
		}

		// Shuffle the collected file paths and corresponding track infos
		rand.Seed(time.Now().UnixNano()) // Seed for randomness
		rand.Shuffle(len(playlistFilePaths), func(i, j int) {
			playlistFilePaths[i], playlistFilePaths[j] = playlistFilePaths[j], playlistFilePaths[i]
			tracksToPlay[i], tracksToPlay[j] = tracksToPlay[j], tracksToPlay[i] // Keep tracksToPlay in sync
		})

		// Load the shuffled playlist into mpv
		if err := player.LoadPlaylistIntoPlayer(cfg, appState, playlistFilePaths, 0); err != nil {
			log.Fatalf("error loading shuffled playlist into player: %v", err)
		}

		// Update state for the first track in the shuffled playlist
		firstTrack := tracksToPlay[0]
		appState.CurrentTrackID = firstTrack.ID
		appState.CurrentTrackTitle = firstTrack.Title
		appState.DownloadedFilePath = playlistFilePaths[0]
		appState.IsPlaying = true
		appState.CurrentPlaylist = playlistName
		appState.LastPlayedTrackIndex = 0

		if err := state.SaveState(); err != nil {
			log.Printf("Error saving state: %v", err)
		}

		// Show status instead of custom message
		ShowStatus()
	},
}
