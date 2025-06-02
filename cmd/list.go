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
	"ytpl/internal/tracks"
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
			ID: appState.CurrentTrackID,
		}

		trackManager, err := tracks.NewManager("", cfg.DownloadDir)
		if err != nil {
			log.Fatalf("error initializing track manager: %v", err)
		}

		track, exists := trackManager.GetTrack(appState.CurrentTrackID)
		trackTitle := appState.CurrentTrackID // Fallback to ID if track not found
		if exists && track != nil {
			trackTitle = track.Title
		}

		err = playlist.AddTrack(playlistName, trackToAdd)
		if err != nil {
			if strings.Contains(err.Error(), "already exists in playlist") {
				fmt.Printf("\n- track '%s' is already in playlist '%s'\n\n", trackTitle, playlistName)
			} else if strings.Contains(err.Error(), "no such file or directory") || strings.Contains(err.Error(), "does not exist") {
				// プレイリストが存在しない場合は作成してから再試行
				if err := playlist.MakePlaylist(playlistName); err != nil {
					log.Fatalf("failed to create playlist '%s': %v\n", playlistName, err)
				}
				// 再度追加を試みる
				err = playlist.AddTrack(playlistName, trackToAdd)
				if err != nil {
					log.Fatalf("error adding song to playlist '%s': %v\n", playlistName, err)
				}
				fmt.Printf("\n- created playlist '%s' and added track '%s'\n\n", playlistName, trackTitle)
			} else {
				log.Fatalf("error adding song to playlist '%s': %v\n", playlistName, err)
			}
		} else {
			fmt.Printf("\n- added track '%s' to playlist '%s'\n\n", trackTitle, playlistName)
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
		fmt.Printf("\n- removed track with ID %s from playlist '%s'.\n", appState.CurrentTrackID, playlistName)
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

		// Initialize track manager to get metadata
		trackManager, err := tracks.NewManager("", cfg.DownloadDir)
		if err != nil {
			log.Fatalf("error initializing track manager: %v", err)
		}

		// Format tracks for display with track numbers, titles, and metadata
		displayItems := make([]struct {
			TrackID     string
			DisplayText string
		}, 0, len(p.Tracks))

		// Get tracks from track manager for metadata
		trackManager, err = tracks.NewManager("", cfg.DownloadDir)
		if err != nil {
			log.Fatalf("error initializing track manager: %v", err)
		}

		for i, track := range p.Tracks {
			trackInfo, found := trackManager.GetTrack(track.ID)
			
			var displayText string
			if found {
				// Format: "01:[3:45] - Track Title"
				durationStr := strings.Trim(util.FormatDuration(trackInfo.Duration), "[]")
				displayText = fmt.Sprintf("%02d:[%s] - %s", i+1, durationStr, trackInfo.Title)
			} else {
				displayText = fmt.Sprintf("%02d: [--:--] - Unknown Track (ID: %s)", i+1, track.ID)
			}

			displayItems = append(displayItems, struct {
				TrackID     string
				DisplayText string
			}{
				TrackID:     track.ID,
				DisplayText: displayText,
			})
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
		trackPath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", selected.TrackID))

		// Get track info for display
		trackInfo, found := trackManager.GetTrack(selected.TrackID)
		trackTitle := ""
		if found {
			trackTitle = trackInfo.Title
		}

		// Start playing the selected track
		if err := player.StartPlayer(cfg, appState, trackPath); err != nil {
			log.Fatalf("error playing track: %v", err)
		}

		// Update the current track info in the app state
		appState.CurrentTrackID = selected.TrackID
		appState.CurrentTrackTitle = trackTitle
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

		// Initialize track manager to get metadata
		trackManager, err := tracks.NewManager("", cfg.DownloadDir)
		if err != nil {
			log.Fatalf("error initializing track manager: %v", err)
		}

		var tracksToPlay []playlist.TrackInfo // Keep track info for state update
		var playlistFilePaths []string        // Paths to pass to mpv

		// Collect all file paths and ensure they are downloaded
		for i, track := range p.Tracks {
			trackInfo, found := trackManager.GetTrack(track.ID)
			trackTitle := fmt.Sprintf("ID: %s", track.ID)
			if found {
				trackTitle = trackInfo.Title
			}

			downloadedFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", track.ID))
			if _, err := os.Stat(downloadedFilePath); os.IsNotExist(err) {
				fmt.Printf("\n- track %d/%d \"%s\" is not stocked locally. downloading...\n", i+1, len(p.Tracks), trackTitle)
				stopSpinner := util.StartSpinner(fmt.Sprintf("\n- downloading \"%s\"", trackTitle))
				_, _, downloadErr := yt.DownloadTrack(cfg, track.ID) // yt.DownloadTrack returns (filePath, TrackInfo, error)
				util.StopSpinner(stopSpinner)
				if downloadErr != nil {
					fmt.Printf("\n- warning: failed to download track \"%s\": %v. skipping from playlist.\n", trackTitle, downloadErr)
					continue // Skip this track if download fails
				}
				fmt.Printf("\n- downloaded \"%s\" to %s.\n", trackTitle, downloadedFilePath)
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
		firstTrackInfo, found := trackManager.GetTrack(firstTrack.ID)
		firstTrackTitle := fmt.Sprintf("ID: %s", firstTrack.ID)
		if found {
			firstTrackTitle = firstTrackInfo.Title
		}

		appState.CurrentTrackID = firstTrack.ID
		appState.CurrentTrackTitle = firstTrackTitle
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

		// Initialize track manager to get metadata
		trackManager, err := tracks.NewManager("", cfg.DownloadDir)
		if err != nil {
			log.Fatalf("error initializing track manager: %v", err)
		}

		var tracksToPlay []playlist.TrackInfo
		var playlistFilePaths []string

		// Collect all file paths and ensure they are downloaded
		for i, track := range p.Tracks {
			trackInfo, found := trackManager.GetTrack(track.ID)
			trackTitle := fmt.Sprintf("ID: %s", track.ID)
			if found {
				trackTitle = trackInfo.Title
			}

			downloadedFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", track.ID))
			if _, err := os.Stat(downloadedFilePath); os.IsNotExist(err) {
				fmt.Printf("\n- track %d/%d \"%s\" is not stocked locally. downloading...\n", i+1, len(p.Tracks), trackTitle)
				stopSpinner := util.StartSpinner(fmt.Sprintf("\n- downloading \"%s\"", trackTitle))
				_, _, downloadErr := yt.DownloadTrack(cfg, track.ID)
				util.StopSpinner(stopSpinner)
				if downloadErr != nil {
					fmt.Printf("\n- warning: failed to download track \"%s\": %v. skipping from playlist.\n", trackTitle, downloadErr)
					continue
				}
				fmt.Printf("\n- downloaded \"%s\" to %s.\n", trackTitle, downloadedFilePath)
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
		firstTrackInfo, found := trackManager.GetTrack(firstTrack.ID)
		firstTrackTitle := fmt.Sprintf("ID: %s", firstTrack.ID)
		if found {
			firstTrackTitle = firstTrackInfo.Title
		}

		appState.CurrentTrackID = firstTrack.ID
		appState.CurrentTrackTitle = firstTrackTitle
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
