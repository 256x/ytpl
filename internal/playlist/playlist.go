// internal/playlist/playlist.go
package playlist

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// TrackInfo represents a single track in a playlist.
// Only the ID is stored in the playlist file.
type TrackInfo struct {
	ID string // YouTube video ID
}

// Playlist holds a list of track IDs.
type Playlist struct {
	Name   string      // Name of the playlist
	Tracks []TrackInfo // List of track IDs in the playlist
}

var (
	playlistsDir string // This will be set by Init function
	mu           sync.Mutex
)

// Init initializes the playlist package with the base directory for playlists.
func Init(dir string) {
	playlistsDir = dir
	if err := os.MkdirAll(playlistsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating playlist directory %s: %v\n", playlistsDir, err)
	}
}

// getPlaylistFilePath returns the full path for a given playlist name.
func getPlaylistFilePath(name string) string {
	sanitizedName := strings.ReplaceAll(name, string(filepath.Separator), "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, "..", "__")
	// Remove .ytpl extension if already present to avoid double extension
	sanitizedName = strings.TrimSuffix(sanitizedName, ".ytpl")
	return filepath.Join(playlistsDir, sanitizedName+".ytpl")
}

// LoadPlaylist loads a playlist from its file.
// Each line in the file should contain a single track ID.
// It tries to load .ytpl file first, and falls back to .txt if not found.
func LoadPlaylist(name string) (*Playlist, error) {
	// First try with .ytpl extension
	path := getPlaylistFilePath(name)
	_, err := os.Stat(path)

	// If .ytpl doesn't exist, try with .txt extension for backward compatibility
	if os.IsNotExist(err) {
		txtPath := strings.TrimSuffix(path, ".ytpl") + ".txt"
		if _, txtErr := os.Stat(txtPath); txtErr == nil {
			path = txtPath
		}
	} else if err != nil {
		return nil, fmt.Errorf("error checking playlist file %s: %w", path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening playlist file %s: %w", path, err)
	}
	defer file.Close()

	playlist := &Playlist{
		Name:   name,
		Tracks: []TrackInfo{},
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		trackID := strings.TrimSpace(scanner.Text())
		if trackID != "" {
			playlist.Tracks = append(playlist.Tracks, TrackInfo{ID: trackID})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading playlist file %s: %w", path, err)
	}

	return playlist, nil
}

// SavePlaylist saves a playlist to its file.
// Each line in the file will contain a single track ID.
// It always saves with .ytpl extension and removes any old .txt version.
func SavePlaylist(p *Playlist) error {
	var sb strings.Builder
	for _, track := range p.Tracks {
		sb.WriteString(track.ID)
		sb.WriteString("\n")
	}

	filePath := getPlaylistFilePath(p.Name)

	// Remove old .txt version if it exists
	txtPath := strings.TrimSuffix(filePath, ".ytpl") + ".txt"
	if _, err := os.Stat(txtPath); err == nil {
		if err := os.Remove(txtPath); err != nil {
			log.Printf("warning: failed to remove old .txt playlist file %s: %v", txtPath, err)
		}
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create playlist directory: %w", err)
	}

	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

// AddTrack adds a track to the specified playlist.
// If the playlist doesn't exist, it will be created.
func AddTrack(playlistName string, track TrackInfo) error {
	p, err := LoadPlaylist(playlistName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			p = &Playlist{Name: playlistName}
		} else {
			return err
		}
	}

	// Check if track already exists in the playlist
	for _, t := range p.Tracks {
		if t.ID == track.ID {
			return fmt.Errorf("track with ID %s already exists in playlist '%s'", track.ID, playlistName)
		}
	}

	p.Tracks = append(p.Tracks, track)
	return SavePlaylist(p)
}

// RemoveTrack removes a track from the specified playlist by ID.
func RemoveTrack(playlistName string, trackID string) error {
	p, err := LoadPlaylist(playlistName)
	if err != nil {
		return err
	}

	var newTracks []TrackInfo
	found := false
	for _, t := range p.Tracks {
		if t.ID == trackID {
			found = true
		} else {
			newTracks = append(newTracks, t)
		}
	}

	if !found {
		return fmt.Errorf("track with ID '%s' not found in playlist '%s'", trackID, playlistName)
	}

	p.Tracks = newTracks
	return SavePlaylist(p)
}

// MakePlaylist creates an empty playlist file.
func MakePlaylist(name string) error {
	filePath := getPlaylistFilePath(name)
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("playlist '%s' already exists", name)
	}
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create playlist file %s: %w", filePath, err)
	}
	return file.Close()
}

// DeletePlaylist deletes a playlist file.
// It will try to delete both .ytpl and .txt versions if they exist.
func DeletePlaylist(name string) error {
	path := getPlaylistFilePath(name)
	err := os.Remove(path)

	// If .ytpl file doesn't exist or was successfully deleted, try .txt
	if os.IsNotExist(err) || err == nil {
		txtPath := strings.TrimSuffix(path, ".ytpl") + ".txt"
		if _, txtErr := os.Stat(txtPath); txtErr == nil {
			if txtErr := os.Remove(txtPath); txtErr != nil {
				log.Printf("warning: failed to remove .txt playlist file %s: %v", txtPath, txtErr)
				// Only return the original error if it exists
				if err == nil {
					err = txtErr
				}
			}
		}
	}
	return err
}

// ListAllPlaylists returns a list of all available playlist names.
// Playlist names are derived from the filenames in the playlist directory.
// It returns each playlist name only once, preferring .ytpl over .txt if both exist.
func ListAllPlaylists() ([]string, error) {
	files, err := os.ReadDir(playlistsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("error reading playlists directory: %w", err)
	}

	// Track seen playlists and their extensions
	playlistSet := make(map[string]string) // name -> best extension found ("ytpl" or "txt")

	var playlists []string

	// First pass: find all playlists and track the best extension
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		var baseName, ext string

		switch {
		case strings.HasSuffix(name, ".ytpl"):
			ext = "ytpl"
			baseName = name[:len(name)-5]
		case strings.HasSuffix(name, ".txt"):
			ext = "txt"
			baseName = name[:len(name)-4]
		default:
			continue
		}

		// If we haven't seen this playlist yet, or we found a better extension
		if currentExt, exists := playlistSet[baseName]; !exists || currentExt == "txt" && ext == "ytpl" {
			playlistSet[baseName] = ext
		}
	}

	// Convert the map to a slice of playlist names
	for name := range playlistSet {
		playlists = append(playlists, name)
	}

	// Sort the playlists alphabetically
	sort.Strings(playlists)

	return playlists, nil
}
