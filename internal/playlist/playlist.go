// internal/playlist/playlist.go
package playlist

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// TrackInfo represents a single entry in a playlist.
// Must be consistent with yt.TrackInfo for ID, Title, and Duration.
type TrackInfo struct {
	ID       string
	Title    string
	Duration float64 // Duration in seconds
}

// Playlist holds a list of tracks.
type Playlist struct {
	Name   string
	Tracks []TrackInfo
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
	return filepath.Join(playlistsDir, sanitizedName+".txt")
}

// LoadPlaylist loads a playlist from its file.
func LoadPlaylist(name string) (*Playlist, error) {
	mu.Lock()
	defer mu.Unlock()

	filePath := getPlaylistFilePath(name)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("playlist '%s' does not exist", name)
		}
		return nil, fmt.Errorf("failed to open playlist file %s: %w", filePath, err)
	}
	defer file.Close()

	p := &Playlist{Name: name}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: ID|Title|Duration
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			track := TrackInfo{
				ID:    parts[0],
				Title: parts[1],
			}
			// If Duration is available, parse it
			if len(parts) >= 3 && parts[2] != "" {
				var duration float64
				n, err := fmt.Sscanf(parts[2], "%f", &duration)
				if err == nil && n == 1 {
					track.Duration = duration
				} else {
					fmt.Fprintf(os.Stderr, "warning: failed to parse duration '%s' in playlist %s: %v\n", parts[2], name, err)
				}
			}
			p.Tracks = append(p.Tracks, track)
		} else {
			fmt.Fprintf(os.Stderr, "warning: malformed line in playlist %s: %s\n", name, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read playlist file %s: %w", filePath, err)
	}
	return p, nil
}

// SavePlaylist saves a playlist to its file.
func SavePlaylist(p *Playlist) error {
	mu.Lock()
	defer mu.Unlock()

	filePath := getPlaylistFilePath(p.Name)
	var sb strings.Builder
	for _, track := range p.Tracks {
		// Format: ID|Title|Duration
		sb.WriteString(fmt.Sprintf("%s|%s|%f\n", track.ID, track.Title, track.Duration))
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

	for _, t := range p.Tracks {
		if t.ID == track.ID {
			return fmt.Errorf("track '%s' (ID: %s) already exists in playlist '%s'", track.Title, track.ID, playlistName)
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
func DeletePlaylist(name string) error {
	filePath := getPlaylistFilePath(name)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("playlist '%s' does not exist", name)
	}
	return os.Remove(filePath)
}

// ListAllPlaylists returns a list of all available playlist names.
func ListAllPlaylists() ([]string, error) {
	files, err := os.ReadDir(playlistsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read playlist directory %s: %w", playlistsDir, err)
	}

	var names []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			names = append(names, strings.TrimSuffix(file.Name(), ".txt"))
		}
	}
	return names, nil
}
