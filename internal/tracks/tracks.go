package tracks

import (
	"log"
	"sort"
	"strings"
	"sync"

	"ytpl/internal/yt"
)

// Log level constants
const (
	LogLevelError = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// Global log level
var logLevel = LogLevelInfo

// SetLogLevel sets the log level for the tracks package
func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "error":
		logLevel = LogLevelError
	case "warn":
		logLevel = LogLevelWarn
	case "info":
		logLevel = LogLevelInfo
	case "debug":
		logLevel = LogLevelDebug
	}
}

// logDebug logs a debug message if debug logging is enabled
func logDebug(format string, v ...interface{}) {
	if logLevel >= LogLevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// logInfo logs an info message if info logging is enabled
func logInfo(format string, v ...interface{}) {
	if logLevel >= LogLevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

// logWarn logs a warning message if warning logging is enabled
func logWarn(format string, v ...interface{}) {
	if logLevel >= LogLevelWarn {
		log.Printf("[WARN] "+format, v...)
	}
}

// logError logs an error message if error logging is enabled
func logError(format string, v ...interface{}) {
	if logLevel >= LogLevelError {
		log.Printf("[ERROR] "+format, v...)
	}
}

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
	logDebug("Add called - ID: %s, Title: %s", track.ID, track.Title)
	
	t.mu.Lock()
	logDebug("Add: lock acquired")
	
	// Check if track already exists
	logDebug("checking for existing track with ID: %s", track.ID)
	for i, existing := range t.tracks {
		if existing.ID == track.ID {
			logInfo("updating existing track: %s", track.ID)
			t.tracks[i] = track
			t.mu.Unlock()
			logDebug("Add: lock released before Save")
			return t.Save()
		}
	}

	// Add new track
	logDebug("adding new track: %s", track.ID)
	t.tracks = append(t.tracks, track)
	logDebug("sorting tracks, current count: %d", len(t.tracks))
	sort.Slice(t.tracks, func(i, j int) bool {
		return t.tracks[i].Title < t.tracks[j].Title
	})

	// Make a copy of the tracks slice to avoid holding the lock during Save
	tracksToSave := make([]yt.TrackInfo, len(t.tracks))
	copy(tracksToSave, t.tracks)
	
	t.mu.Unlock()
	logDebug("Add: lock released before Save")

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
			return t.Save()
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
	logDebug("Save called, saving tracks to %s", t.path)
	logDebug("Save: attempting to acquire lock")
	
	// Create a copy of the tracks to avoid holding the lock during file I/O
	t.mu.Lock()
	logDebug("Save: lock acquired")
	tracksCopy := make([]yt.TrackInfo, len(t.tracks))
	copy(tracksCopy, t.tracks)
	t.mu.Unlock()
	logDebug("Save: lock released for file I/O")

	// Create a temporary Tracks instance for saving
	tmpTracks := &Tracks{
		path:   t.path,
		tracks: tracksCopy,
	}

	// Save the tracks (this will acquire its own lock)
	logDebug("Save: calling saveTracks")
	err := saveTracks(tmpTracks)
	if err != nil {
		logError("error saving tracks: %v", err)
		return err
	}

	logInfo("successfully saved %d tracks to %s", len(tracksCopy), t.path)
	return nil
}

// SaveAll saves all tracks to file in a single operation
// This is more efficient than calling Save() after each Add()
func (t *Tracks) SaveAll() error {
	logDebug("SaveAll called, saving all tracks to %s", t.path)
	
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	logDebug("SaveAll: saving %d tracks", len(t.tracks))
	return saveTracks(t)
}
