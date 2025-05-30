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

	"ytpl/internal/player"
	"ytpl/internal/playlist"
	"ytpl/internal/state"
	"ytpl/internal/util"
	"ytpl/internal/yt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Manage and play music playlists",
	Long:  `Subcommands for managing playlists: make, add, remove, del, show, play, shuffle.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// If no arguments, show playlist selection prompt
			playlistNames, err := playlist.ListAllPlaylists()
			if err != nil {
				log.Fatalf("Error listing playlists: %v", err)
			}

			if len(playlistNames) == 0 {
				fmt.Println("No playlists found. Use 'ytpl list make <name>' to create one.")
				return
			}

			selectedPlaylist, err := util.SelectFromList(
				"Select a playlist:",
				playlistNames,
				func(name string) string { return name },
			)
			if err != nil {
				if strings.Contains(err.Error(), "cancelled") {
					fmt.Println("Playlist selection cancelled.")
					return
				}
				log.Fatalf("Error selecting playlist: %v", err)
			}

			action, err := util.SelectFromList(
				"Select action for playlist '"+selectedPlaylist+"':",
				[]string{"Play (ordered)", "Play (shuffled)", "Show contents"},
				func(s string) string { return s },
			)
			if err != nil {
				if strings.Contains(err.Error(), "cancelled") {
					fmt.Println("Action cancelled.")
					return
				}
				log.Fatalf("Error selecting action: %v", err)
			}

			switch action {
			case "Play (ordered)":
				listPlayCmd.Run(listPlayCmd, []string{selectedPlaylist})
			case "Play (shuffled)":
				listShuffleCmd.Run(listShuffleCmd, []string{selectedPlaylist})
			case "Show contents":
				listShowCmd.Run(listShowCmd, []string{selectedPlaylist})
			default:
				fmt.Println("Invalid action selected.")
			}
			// Status will be shown by the respective command handlers
			return
		}
		// If arguments are provided, Cobra's default behavior handles subcommand execution.
		cmd.Help()
	},
}

var listMakeCmd = &cobra.Command{
	Use:   "make <playlist_name>",
	Short: "Create a new playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := playlist.MakePlaylist(name); err != nil {
			log.Fatalf("Error creating playlist '%s': %v", name, err)
		}
		fmt.Printf("Playlist '%s' created successfully.\n", name)
	},
}

var listAddCmd = &cobra.Command{
	Use:   "add <playlist_name>",
	Short: "Add the currently playing song to a playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		playlistName := args[0]

		if appState.CurrentTrackID == "" {
			fmt.Println("No song is currently playing to add to a playlist.")
			return
		}

		trackToAdd := playlist.TrackInfo{
			ID:    appState.CurrentTrackID,
			Title: appState.CurrentTrackTitle,
		}

		err := playlist.AddTrack(playlistName, trackToAdd)
		if err != nil {
			if strings.Contains(err.Error(), "already exists in playlist") {
				fmt.Printf("Track '%s' is already in playlist '%s'\n", appState.CurrentTrackTitle, playlistName)
			} else {
				log.Fatalf("Error adding song to playlist '%s': %v", playlistName, err)
			}
		} else {
			fmt.Printf("Added \"%s\" to playlist '%s'\n", appState.CurrentTrackTitle, playlistName)
		}
	},
}

var listRemoveCmd = &cobra.Command{
	Use:   "remove <playlist_name>",
	Short: "Remove the currently playing song from a playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		playlistName := args[0]

		if appState.CurrentTrackID == "" {
			fmt.Println("No song is currently playing to remove from a playlist.")
			return
		}

		if err := playlist.RemoveTrack(playlistName, appState.CurrentTrackID); err != nil {
			log.Fatalf("Error removing song from playlist '%s': %v", playlistName, err)
		}
		fmt.Printf("Removed \"%s\" from playlist '%s'.\n", appState.CurrentTrackTitle, playlistName)
	},
}

var listDelCmd = &cobra.Command{
	Use:   "del <playlist_name>",
	Short: "Delete a playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		confirm, err := util.Confirm(fmt.Sprintf("Are you sure you want to delete playlist '%s'?", name))
		if err != nil || !confirm {
			fmt.Println("Playlist deletion cancelled.")
			return
		}

		if err := playlist.DeletePlaylist(name); err != nil {
			log.Fatalf("Error deleting playlist '%s': %v", name, err)
		}
		fmt.Printf("Playlist '%s' deleted successfully.\n", name)
	},
}

var listShowCmd = &cobra.Command{
	Use:   "show <playlist_name>",
	Short: "Show contents of a playlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		p, err := playlist.LoadPlaylist(name)
		if err != nil {
			log.Fatalf("Error loading playlist '%s': %v", name, err)
		}

		fmt.Printf("Playlist '%s' (%d songs):\n", p.Name, len(p.Tracks))
		if len(p.Tracks) == 0 {
			fmt.Println("  (empty)")
			return
		}
		for i, track := range p.Tracks {
			fmt.Printf("  %d. %s (ID: %s)\n", i+1, track.Title, track.ID)
		}
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
			fmt.Printf("Playlist '%s' is empty. Nothing to play.\n", playlistName)
			return
		}

		// No extra empty line here, as per your request.

		var tracksToPlay []playlist.TrackInfo // Keep track info for state update
		var playlistFilePaths []string        // Paths to pass to mpv

		// Collect all file paths and ensure they are downloaded
		for i, track := range p.Tracks {
			downloadedFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", track.ID))
			if _, err := os.Stat(downloadedFilePath); os.IsNotExist(err) {
				fmt.Printf("Track %d/%d \"%s\" is not stocked locally. Downloading...\n", i+1, len(p.Tracks), track.Title)
				stopSpinner := util.StartSpinner(fmt.Sprintf("Downloading \"%s\"", track.Title))
				_, _, downloadErr := yt.DownloadTrack(cfg, track.ID) // yt.DownloadTrack returns (filePath, TrackInfo, error)
				util.StopSpinner(stopSpinner)
				if downloadErr != nil {
					fmt.Printf("Warning: Failed to download track \"%s\": %v. Skipping from playlist.\n", track.Title, downloadErr)
					continue // Skip this track if download fails
				}
				fmt.Printf("Downloaded \"%s\" to %s.\n", track.Title, downloadedFilePath)
			}
			tracksToPlay = append(tracksToPlay, track)
			playlistFilePaths = append(playlistFilePaths, downloadedFilePath)
		}

		if len(playlistFilePaths) == 0 {
			fmt.Printf("No playable songs found in playlist '%s' after checking local stock.\n", playlistName)
			return
		}

		// Load the entire playlist into mpv using LoadPlaylistIntoPlayer
		if err := player.LoadPlaylistIntoPlayer(cfg, appState, playlistFilePaths, 0); err != nil { // Start from index 0
			log.Fatalf("Error loading playlist into player: %v", err)
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
			fmt.Printf("Playlist '%s' is empty. Nothing to shuffle and play.\n", playlistName)
			return
		}

		// No extra empty line here, as per your request.

		var tracksToPlay []playlist.TrackInfo
		var playlistFilePaths []string

		// Collect all file paths and ensure they are downloaded
		for i, track := range p.Tracks {
			downloadedFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", track.ID))
			if _, err := os.Stat(downloadedFilePath); os.IsNotExist(err) {
				fmt.Printf("Track %d/%d \"%s\" is not stocked locally. Downloading...\n", i+1, len(p.Tracks), track.Title)
				stopSpinner := util.StartSpinner(fmt.Sprintf("Downloading \"%s\"", track.Title))
				_, _, downloadErr := yt.DownloadTrack(cfg, track.ID)
				util.StopSpinner(stopSpinner)
				if downloadErr != nil {
					fmt.Printf("Warning: Failed to download track \"%s\": %v. Skipping from playlist.\n", track.Title, downloadErr)
					continue
				}
				fmt.Printf("Downloaded \"%s\" to %s.\n", track.Title, downloadedFilePath)
			}
			tracksToPlay = append(tracksToPlay, track)
			playlistFilePaths = append(playlistFilePaths, downloadedFilePath)
		}

		if len(playlistFilePaths) == 0 {
			fmt.Printf("No playable songs found in playlist '%s' after checking local stock.\n", playlistName)
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
			log.Fatalf("Error loading shuffled playlist into player: %v", err)
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
