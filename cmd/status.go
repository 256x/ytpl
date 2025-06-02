// cmd/status.go
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"

	"ytpl/internal/player"
	"ytpl/internal/playertags"
	"ytpl/internal/state"
	"ytpl/internal/tracks"
	"ytpl/internal/yt"

	"github.com/spf13/cobra"
)

// isFirstOutput tracks if this is the first output to the console
var isFirstOutput int32 = 1 // 1 means true (first output), 0 means false

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current playback status",
	Run: func(cmd *cobra.Command, args []string) {
		ShowStatus()
	},
}

// formatDuration converts seconds to HH:MM:SS format.
func formatDuration(s int) string {
	h := s / 3600
	s %= 3600
	m := s / 60
	s %= 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// ShowStatus displays the current playback status in a single line format
// with consistent newlines around it
func ShowStatus() {
	// Add a newline before status if this is not the first output
	if atomic.LoadInt32(&isFirstOutput) == 0 {
		fmt.Println()
	}

	if appState.PID == 0 {
		fmt.Println("No track playing")
		atomic.StoreInt32(&isFirstOutput, 0)
		return
	}
	atomic.StoreInt32(&isFirstOutput, 0)

	// Update appState with real-time info from MPV for accurate display
	updateAppStateFromMpvStatus()

	// Re-check appState.PID after update, as it might have become 0 if mpv exited
	if appState.PID == 0 {
		fmt.Println("No track playing")
		return
	}

	// Get the best available display title
	currentDisplayTitle := getBestAvailableTitle()

	// Get volume
	volume, _ := player.GetProperty(appState, "volume")
	volumePercent := 100
	if vol, ok := volume.(float64); ok {
		volumePercent = int(vol)
	}

	// Build the status line
	statusLine := fmt.Sprintf("\n♪  %s", currentDisplayTitle)

	// Add playlist info if available
	if appState.CurrentPlaylist != "" {
		statusLine += fmt.Sprintf(" @ %s", appState.CurrentPlaylist)
	}

	// Add time and volume info
	statusLine += fmt.Sprintf(" | 🔊 %d%%\n", volumePercent)

	// Print the status line with a newline after it
	fmt.Printf("%s\n", statusLine)
}

// getBestAvailableTitle returns the best available title for the current track
func getBestAvailableTitle() string {
	if appState.DownloadedFilePath == "" {
		return "Unknown Track"
	}

	// Get track ID from file path
	base := filepath.Base(appState.DownloadedFilePath)
	trackID := strings.TrimSuffix(base, filepath.Ext(base))

	// Initialize track manager
	trackManager, err := tracks.NewManager(filepath.Dir(cfg.DownloadDir), cfg.DownloadDir)
	if err != nil {
		return strings.TrimSuffix(base, filepath.Ext(base))
	}

	// Get track info from .tracks file
	track, exists := trackManager.GetTrack(trackID)
	if exists && track != nil {
		return track.Title
	}

	// Fallback to filename
	return strings.TrimSuffix(base, filepath.Ext(base))
}


// updateAppStateFromMpvStatus fetches current playing info from mpv and updates appState.
func updateAppStateFromMpvStatus() {
    currentFilePath, currentPlaylistPos, err := player.GetCurrentlyPlayingTrackInfo(appState)
    if err != nil {
        if appState.PID != 0 && strings.Contains(err.Error(), "player not reachable") {
            player.StopPlayer(appState)
        }
        return
    }

    if currentFilePath != "" {
        _, fileName := filepath.Split(currentFilePath)
        currentTrackID := strings.TrimSuffix(fileName, filepath.Ext(fileName))

        // Get ytTrackInfo for its title as a primary fallback
        ytTrackInfo, err := yt.GetLocalTrackInfo(cfg, currentTrackID)
        var ytTitle, ytArtist string
        if err == nil && ytTrackInfo != nil {
            ytTitle = ytTrackInfo.Title
            ytArtist = ytTrackInfo.Uploader
            if ytArtist == "" {
                ytArtist = ytTrackInfo.Creator
            }
        } else {
            ytTitle = strings.TrimSuffix(fileName, filepath.Ext(fileName)) // Fallback to filename
            ytArtist = ""
        }

        // Initialize track manager
        trackManager, err := tracks.NewManager(filepath.Dir(cfg.DownloadDir), cfg.DownloadDir)
        if err != nil {
            return
        }

        // Get track info from .tracks file
        var audioInfo *playertags.AudioInfo
        if track, exists := trackManager.GetTrack(currentTrackID); exists && track != nil {
            audioInfo = &playertags.AudioInfo{
                Title:  track.Title,
                Artist: track.Uploader,
            }
            ytTitle = track.Title
            ytArtist = track.Uploader
        } else {
            audioInfo = &playertags.AudioInfo{
                Title:  ytTitle,
                Artist: ytArtist,
            }
        }

        // Determine the best display title based on priority:
        // 1. MP3 tag Title (if available and not generic)
        // 2. YouTube JSON Title (ytTrackInfo.Title)
        // 3. Filename
        var bestDisplayTitle string
        if audioInfo != nil && audioInfo.Title != "" && !strings.Contains(audioInfo.Title, "Unknown Title") {
            bestDisplayTitle = audioInfo.Title // Use MP3 tag Title directly
        } else if ytTrackInfo != nil && ytTrackInfo.Title != "" && !strings.Contains(ytTrackInfo.Title, "Unknown Title") {
            bestDisplayTitle = ytTrackInfo.Title // Use YouTube's original title
        } else {
            bestDisplayTitle = strings.TrimSuffix(fileName, filepath.Ext(fileName)) // Fallback to filename
        }

        appState.CurrentTrackTitle = bestDisplayTitle // Update appState with the chosen display title
        appState.CurrentTrackID = currentTrackID
        appState.DownloadedFilePath = currentFilePath
        appState.LastPlayedTrackIndex = currentPlaylistPos
        state.SaveState()
    } else {
        appState.CurrentTrackID = ""
        appState.CurrentTrackTitle = ""
        appState.DownloadedFilePath = ""
        appState.LastPlayedTrackIndex = -1
        state.SaveState()
    }
}
