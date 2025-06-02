package tracks

import (
	"sort"
	"sync"

	"ytpl/internal/yt"
)


// Tracks manages the collection of all tracks
type Tracks struct {
	mu     sync.RWMutex
	tracks []yt.TrackInfo
	path   string
}

// New creates a new Tracks instance
func New(path string) *Tracks {
	return &Tracks{
		tracks: make([]yt.TrackInfo, 0),
		path:   path,
	}
}

// Add adds a track to the tracks list if it doesn't already exist
func (t *Tracks) Add(track yt.TrackInfo) error {
	t.mu.Lock()
	
	// Check if track already exists
	for i, existing := range t.tracks {
		if existing.ID == track.ID {
			t.tracks[i] = track
			t.mu.Unlock()
			return t.Save()
		}
	}

	// Add new track
	t.tracks = append(t.tracks, track)
	sort.Slice(t.tracks, func(i, j int) bool {
		return t.tracks[i].Title < t.tracks[j].Title
	})

	// Make a copy of the tracks slice to avoid holding the lock during Save
	tracksToSave := make([]yt.TrackInfo, len(t.tracks))
	copy(tracksToSave, t.tracks)
	
	t.mu.Unlock()

	// Save the tracks (this will acquire its own lock)
	return t.Save()
}

// Remove removes a track by ID
func (t *Tracks) Remove(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, track := range t.tracks {
		if track.ID == id {
			t.tracks = append(t.tracks[:i], t.tracks[i+1:]...)
			return nil
		}
	}
	return nil
}

// Get returns a track by ID
func (t *Tracks) Get(id string) (*yt.TrackInfo, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, track := range t.tracks {
		if track.ID == id {
			return &track, true
		}
	}
	return nil, false
}

// List returns all tracks
func (t *Tracks) List() []yt.TrackInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]yt.TrackInfo, len(t.tracks))
	copy(result, t.tracks)
	return result
}

// Load loads tracks from file
func (t *Tracks) Load() error {
	return loadTracks(t)
}

// Clear removes all tracks
func (t *Tracks) Clear() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tracks = nil
	return nil
}

// Save saves tracks to file
func (t *Tracks) Save() error {
	// Create a copy of the tracks to avoid holding the lock during file I/O
	t.mu.Lock()
	tracksCopy := make([]yt.TrackInfo, len(t.tracks))
	copy(tracksCopy, t.tracks)
	t.mu.Unlock()

	// Create a temporary Tracks instance for saving
	tmpTracks := &Tracks{
		path:   t.path,
		tracks: tracksCopy,
	}

	// Save the tracks (this will acquire its own lock)
	return saveTracks(tmpTracks)
}

// SaveAll saves all tracks to file in a single operation
// This is more efficient than calling Save() after each Add()
func (t *Tracks) SaveAll() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	return saveTracks(t)
}
