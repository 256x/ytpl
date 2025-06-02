package tracks

import (
	"fmt"
	"os"
	"path/filepath"

	"ytpl/internal/yt"
)

// Manager manages the tracks library
type Manager struct {
	tracks     *Tracks
	batchMode  bool
	configDir  string
	dataDir    string
}

// BatchMode enables or disables batch mode
// When enabled, tracks are not saved after each operation
// Call SaveAll() explicitly to save all changes
func (m *Manager) BatchMode(enabled bool) {
	m.batchMode = enabled
}

// NewManager creates a new track manager
func NewManager(configDir, dataDir string) (*Manager, error) {
	tracksPath := filepath.Join(dataDir, ".tracks")
	t := New(tracksPath)
	
	// Load existing tracks or create new file
	if err := t.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading tracks: %w", err)
	}
	return &Manager{
		tracks:     t,
		batchMode:  false,
		configDir:  configDir,
		dataDir:    dataDir,
	}, nil
}

// AddTrack adds a track to the library
func (m *Manager) AddTrack(track yt.TrackInfo) error {
	// If not in batch mode, let the tracks handle saving
	if !m.batchMode {
		if err := m.tracks.Add(track); err != nil {
			return err
		}
		return nil
	}
	
	// In batch mode, just add to the in-memory list
	m.tracks.mu.Lock()
	defer m.tracks.mu.Unlock()
	
	// Check if track already exists
	for i, existing := range m.tracks.tracks {
		if existing.ID == track.ID {
			m.tracks.tracks[i] = track
			return nil
		}
	}
	
	// Add new track
	m.tracks.tracks = append(m.tracks.tracks, track)
	return nil
}

// GetTrack returns a track by ID
func (m *Manager) GetTrack(id string) (*yt.TrackInfo, bool) {
	return m.tracks.Get(id)
}

// ListTracks returns all tracks
func (m *Manager) ListTracks() []yt.TrackInfo {
	return m.tracks.List()
}

// RemoveTrack removes a track from the library
func (m *Manager) RemoveTrack(id string) error {
	if err := m.tracks.Remove(id); err != nil {
		return fmt.Errorf("error removing track: %w", err)
	}

	// Save changes if not in batch mode
	if !m.batchMode {
		return m.Save()
	}
	return nil
}

// UpdateTrack updates a track in the manager.
func (m *Manager) UpdateTrack(track *yt.TrackInfo) error {
	// Find and update the track
	updated := false
	for i, t := range m.tracks.tracks {
		if t.ID == track.ID {
			m.tracks.tracks[i] = *track
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("track not found: %s", track.ID)
	}

	// Save changes if not in batch mode
	if !m.batchMode {
		return m.Save()
	}
	return nil
}

// Clear removes all tracks from the manager
func (m *Manager) Clear() error {
	return m.tracks.Clear()
}

// SaveAll saves all tracks to the storage
// This is more efficient than saving after each operation
func (m *Manager) SaveAll() error {
	// Create a copy of tracks to avoid holding the lock during save
	m.tracks.mu.RLock()
	tracksCopy := make([]yt.TrackInfo, len(m.tracks.tracks))
	copy(tracksCopy, m.tracks.tracks)
	m.tracks.mu.RUnlock()
	
	// Create a temporary Tracks instance for saving
	tmpTracks := &Tracks{
		path:   m.tracks.path,
		tracks: tracksCopy,
	}
	
	// Save the tracks (this will acquire its own lock)
	return saveTracks(tmpTracks)
}

// Save saves the current tracks to the tracks file
func (m *Manager) Save() error {
	return m.tracks.Save()
}
