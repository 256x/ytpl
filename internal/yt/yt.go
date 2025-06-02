// internal/yt/yt.go
package yt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
			fmt.Fprintf(os.Stderr, "warning: failed to unmarshal yt-dlp json line: %v\nline: %s\n", err, line)
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
		fmt.Fprintf(os.Stderr, "warning: failed to unmarshal yt-dlp download info json from stdout: %v\noutput: %s\n", err, string(output))
		// Fallback: try to load from local .info.json if JSON output from yt-dlp was malformed
		infoPath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.info.json", trackID))
		if data, readErr := os.ReadFile(infoPath); readErr == nil {
			if json.Unmarshal(data, &downloadedTrackInfo) == nil {
				fmt.Fprintf(os.Stderr, "successfully loaded info from local file: %s\n", infoPath)
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

	// Optimize the info.json file to remove unnecessary fields
	if err := OptimizeInfoJSON(cfg, trackID); err != nil {
		// Non-fatal error, just log it
		log.Printf("warning: failed to optimize info.json for %s: %v", trackID, err)
	}

	return downloadedFilePath, &downloadedTrackInfo, nil
}

// ListLocalTracks returns a list of all locally downloaded tracks, sorted by title.
func ListLocalTracks(cfg *config.Config) ([]*TrackInfo, error) {
	files, err := filepath.Glob(filepath.Join(cfg.DownloadDir, "*.info.json"))
	if err != nil {
		return nil, fmt.Errorf("error listing local tracks: %w", err)
	}

	var tracks []*TrackInfo
	for _, file := range files {
		trackID := strings.TrimSuffix(filepath.Base(file), ".info.json")
		track, err := GetLocalTrackInfo(cfg, trackID)
		if err != nil {
			log.Printf("warning: failed to get info for %s: %v", trackID, err)
			continue
		}
		tracks = append(tracks, track)
	}

	// Sort tracks by title (case-insensitive)
	sort.Slice(tracks, func(i, j int) bool {
		return strings.ToLower(tracks[i].Title) < strings.ToLower(tracks[j].Title)
	})

	return tracks, nil
}

// OptimizeInfoJSON optimizes the info.json file by keeping only necessary fields
func OptimizeInfoJSON(cfg *config.Config, trackID string) error {
	infoPath := filepath.Join(cfg.DownloadDir, fmt.Sprintf("%s.info.json", trackID))
	
	// Read the existing info.json
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to read info.json for optimization: %w", err)
	}

	// Parse into a map to access all fields
	var infoMap map[string]interface{}
	if err := json.Unmarshal(data, &infoMap); err != nil {
		return fmt.Errorf("failed to parse info.json: %w", err)
	}

	// Define the fields we want to keep
	keepFields := []string{
		"id",
		"title",
		"uploader",
		"creator",
		"duration",
		"release_year",
		"upload_date",
		"webpage_url",
	}

	// Create a new map with only the fields we want to keep
	optimized := make(map[string]interface{})
	for _, field := range keepFields {
		if value, exists := infoMap[field]; exists {
			optimized[field] = value
		}
	}

	// Add some essential fields if they're missing but we can derive them
	if _, exists := optimized["id"]; !exists {
		optimized["id"] = trackID
	}

	// Write the optimized JSON back to the file
	optimizedJSON, err := json.MarshalIndent(optimized, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal optimized info.json: %w", err)
	}

	// Write to a temporary file first, then rename atomically
	tempPath := infoPath + ".tmp"
	if err := os.WriteFile(tempPath, optimizedJSON, 0644); err != nil {
		return fmt.Errorf("failed to write optimized info.json: %w", err)
	}

	// Atomic rename to replace the original file
	if err := os.Rename(tempPath, infoPath); err != nil {
		return fmt.Errorf("failed to replace original info.json: %w", err)
	}

	return nil
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
		return nil, fmt.Errorf("failed to unmarshal info json for track %s: %w", trackID, err)
	}

	// Populate missing fields if they are empty
	if track.Title == "" {
		track.Title = "unknown title from json"
	}
	if track.Uploader == "" && track.Creator == "" {
		track.Uploader = "unknown artist from json" // default fallback for artist
	}
	// Add more fallbacks for other fields if needed

	return &track, nil
}
