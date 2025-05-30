// internal/yt/yt.go
package yt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	config "ytpl/internal/config" // Alias for internal/config
)

// TrackInfo represents metadata for a YouTube video or downloaded track.
// Fields correspond to yt-dlp's --dump-json output.
type TrackInfo struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	WebpageURL string  `json:"webpage_url"`
	Duration   float64 `json:"duration"` // In seconds

	// Additional fields for richer local metadata, corresponding to yt-dlp's JSON output
	Uploader    string `json:"uploader"`      // The channel name (often the artist)
	Creator     string `json:"creator"`       // Sometimes more specific artist info
	Album       string `json:"album"`         // Album name from metadata
	ReleaseYear int    `json:"release_year"`  // Year of release from metadata
	ViewCount   int64  `json:"view_count"`    // Number of views
	UploadDate  string `json:"upload_date"`   // Upload date in YYYYMMDD format
	// Add more fields from yt-dlp's --dump-json output as needed, e.g.,
	// Channel        string `json:"channel"`
	// ChannelURL     string `json:"channel_url"`
	// LikeCount      int    `json:"like_count"`
	// Thumbnail      string `json:"thumbnail"`
}

// SearchYouTube searches YouTube using yt-dlp and returns a list of TrackInfo.
// This function is optimized for speed and only retrieves essential metadata.
func SearchYouTube(cfg *config.Config, query string) ([]TrackInfo, error) {
	// Use cfg.MaxSearchResults for the number of results.
	// config.LoadConfig ensures it's at least 1 and defaults to 10 if not set.
	numResults := cfg.MaxSearchResults

	cmdArgs := []string{
		"--dump-json",
		"--flat-playlist", // Fast search with minimal metadata
		"--print-json",    // Print one JSON object per line
		"--no-warnings",   // Ignore warnings
		"--ignore-errors", // Continue on download errors, e.g. skip unavailable videos
		"--match-filter", "!is_live & !is_upcoming", // Filter out live and upcoming videos
		fmt.Sprintf("ytsearch%d:%s", numResults, query),
	}

	cmd := exec.Command(cfg.YtDlpPath, cmdArgs...)

	// Redirect stderr to a buffer to prevent yt-dlp messages from interfering with TUI.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute yt-dlp search: %w\nStderr: %s", err, stderr.String())
	}

	var tracks []TrackInfo
	// yt-dlp --print-json outputs one JSON object per line for multiple results.
	// Process each line as a separate JSON object.
	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var track TrackInfo
		if err := json.Unmarshal([]byte(line), &track); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to unmarshal yt-dlp JSON line: %v\nLine: %s\n", err, line)
			continue
		}
		// Skip channels: channels typically have 0 duration and specific URL pattern.
		// This helps filter out non-playable items from search results.
		if track.Duration == 0 && strings.Contains(track.WebpageURL, "/channel/") {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// DownloadTrack downloads a YouTube video as an audio file.
// It returns the path to the downloaded file and its full metadata.
func DownloadTrack(cfg *config.Config, trackID string) (string, *TrackInfo, error) {
	outputTemplate := filepath.Join(cfg.DownloadDir, "%(id)s.%(ext)s")

	cmdArgs := []string{
		"-o", outputTemplate,         // Output file template
		"--extract-audio",            // Extract audio
		"--audio-format", "mp3",      // Convert to MP3
		"--audio-quality", "0",       // Best audio quality
		"--restrict-filenames",       // Restrict filenames to ASCII and numeric
		"--no-simulate",              // Ensure actual download occurs
		"--print-json",               // Print final info JSON to stdout after download
		"--no-progress",              // Suppress download progress bar for cleaner output
		"--write-info-json",          // Save metadata to a .info.json file alongside the audio
		"--embed-metadata",           // Explicitly embed metadata into the audio file
		"--embed-thumbnail",          // Embed thumbnail into the audio file (optional, makes it larger)
	}

	// Add cookie options if configured, for age-restricted content
	if cfg.CookieBrowser != "" {
		cookieArg := fmt.Sprintf("--cookies-from-browser=%s", cfg.CookieBrowser)
		if cfg.CookieProfile != "" {
			cookieArg = fmt.Sprintf("--cookies-from-browser=%s:%s", cfg.CookieBrowser, cfg.CookieProfile)
		}
		cmdArgs = append(cmdArgs, cookieArg)
	}

	// URL argument must be passed after '--' for safety
	cmdArgs = append(cmdArgs, "--", fmt.Sprintf("https://www.youtube.com/watch?v=%s", trackID))

	cmd := exec.Command(cfg.YtDlpPath, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr // Redirect yt-dlp's stderr to a buffer

	output, err := cmd.Output() // Execute the command and capture stdout
	if err != nil {
		return "", nil, fmt.Errorf("failed to execute yt-dlp download for ID %s: %w\nStderr: %s", trackID, err, stderr.String())
	}

	var downloadedTrackInfo TrackInfo
	if err := json.Unmarshal(output, &downloadedTrackInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to unmarshal yt-dlp download info JSON from stdout: %v\nOutput: %s\n", err, string(output))
		// Fallback: try to load from local .info.json if JSON output from yt-dlp was malformed
		infoPath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.info.json", trackID))
		if data, readErr := os.ReadFile(infoPath); readErr == nil {
			if json.Unmarshal(data, &downloadedTrackInfo) == nil {
				fmt.Fprintf(os.Stderr, "Successfully loaded info from local file: %s\n", infoPath)
			}
		}
	}

	// Ensure basic info is populated even if JSON parsing fails or is incomplete
	if downloadedTrackInfo.ID == "" {
		downloadedTrackInfo.ID = trackID
		downloadedTrackInfo.Title = fmt.Sprintf("Unknown Title (ID: %s)", trackID)
	}

	// Construct the likely downloaded file path (assuming mp3 extension based on --audio-format)
	downloadedFilePath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.mp3", trackID))

	// Verify the downloaded file exists on disk
	if _, err := os.Stat(downloadedFilePath); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("downloaded file not found at expected path: %s", downloadedFilePath)
	}

	return downloadedFilePath, &downloadedTrackInfo, nil
}

// GetLocalTrackInfo reads metadata for a local track from its .info.json file.
// It tries to populate as much information as possible from the .info.json.
func GetLocalTrackInfo(cfg *config.Config, trackID string) (*TrackInfo, error) {
	infoPath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.info.json", trackID))
	data, err := os.ReadFile(infoPath)
	if err != nil {
		// Return error but indicate it's a "soft" error, allowing caller to fallback
		return nil, fmt.Errorf("failed to read info JSON for track %s: %w", trackID, err)
	}

	var track TrackInfo
	if err := json.Unmarshal(data, &track); err != nil {
		return nil, fmt.Errorf("failed to unmarshal info JSON for track %s: %w", trackID, err)
	}

	// Populate missing fields if they are empty
	if track.Title == "" {
		track.Title = "Unknown Title from JSON"
	}
	if track.Uploader == "" && track.Creator == "" {
		track.Uploader = "Unknown Artist from JSON" // Default fallback for artist
	}
	// Add more fallbacks for other fields if needed

	return &track, nil
}
