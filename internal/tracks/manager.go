package tracks

import (
	"fmt"
	"log"
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
	log.Printf("initializing track manager with dataDir: %s", dataDir)
	tracksPath := filepath.Join(dataDir, ".tracks")
	log.Printf("tracks file path: %s", tracksPath)
	
	t := New(tracksPath)
	
	// Load existing tracks or create new file
	log.Println("loading tracks...")
	if err := t.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("error loading tracks: %v", err)
			return nil, err
		}
		log.Println("no existing tracks file found, will create new one")
	} else {
		log.Printf("successfully loaded tracks from %s", tracksPath)
	}
	
	log.Printf("track manager initialized with %d tracks", len(t.tracks))
	return &Manager{
		tracks:     t,
		batchMode:  false,
		configDir:  configDir,
		dataDir:    dataDir,
	}, nil
}

// AddTrack adds a track to the library
func (m *Manager) AddTrack(track yt.TrackInfo) error {
	log.Printf("manager.AddTrack called - ID: %s, Title: %s", track.ID, track.Title)
	currentCount := len(m.tracks.List())
	log.Printf("current tracks count before add: %d", currentCount)
	
	// If not in batch mode, let the tracks handle saving
	if !m.batchMode {
		err := m.tracks.Add(track)
		if err != nil {
			log.Printf("error adding track: %v", err)
			return err
		}
		log.Printf("successfully added track to library: %s (new count: %d)", track.Title, len(m.tracks.List()))
		log.Printf("track details: %+v", track)
		return nil
	}
	
	// In batch mode, just add to the in-memory list
	m.tracks.mu.Lock()
	defer m.tracks.mu.Unlock()
	
	// Check if track already exists
	for i, existing := range m.tracks.tracks {
		if existing.ID == track.ID {
			log.Printf("updating existing track in batch mode: %s", track.ID)
			m.tracks.tracks[i] = track
			return nil
		}
	}
	
	// Add new track
	log.Printf("adding new track in batch mode: %s", track.ID)
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
	logDebug("SaveAll: entering, batch mode: %v", m.batchMode)
	
	// Create a copy of tracks to avoid holding the lock during save
	m.tracks.mu.RLock()
	tracksCopy := make([]yt.TrackInfo, len(m.tracks.tracks))
	copy(tracksCopy, m.tracks.tracks)
	m.tracks.mu.RUnlock()
	
	logDebug("SaveAll: saving %d tracks in batch mode", len(tracksCopy))
	
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
