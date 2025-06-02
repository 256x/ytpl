package tracks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ytpl/internal/yt"
)

// loadTracks loads tracks from file
// Note: Caller must hold the lock
func loadTracks(t *Tracks) error {
	data, err := os.ReadFile(t.path)
	if err != nil {
		if os.IsNotExist(err) {
			t.tracks = []yt.TrackInfo{}
			return nil
		}
		return err
	}

	if len(data) == 0 {
		t.tracks = []yt.TrackInfo{}
		return nil
	}

	var tracks []yt.TrackInfo
	if err := json.Unmarshal(data, &tracks); err != nil {
		return err
	}

	t.tracks = tracks
	return nil
}

// saveTracks saves tracks to file
// Note: This function should not be called with a lock held
func saveTracks(t *Tracks) error {
	logDebug("saveTracks: preparing to save %d tracks to %s", len(t.tracks), t.path)
	
	// Ensure directory exists
	dir := filepath.Dir(t.path)
	logDebug("saveTracks: ensuring directory exists: %s", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logError("saveTracks: failed to create directory %s: %v", dir, err)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal tracks to JSON
	logDebug("saveTracks: marshaling %d tracks to JSON", len(t.tracks))
	data, err := json.MarshalIndent(t.tracks, "", "  ")
	if err != nil {
		logError("saveTracks: failed to marshal tracks: %v", err)
		return fmt.Errorf("failed to marshal tracks: %w", err)
	}

	// Create a temporary file in the same directory
	tmpFile, err := os.CreateTemp(dir, ".tracks.tmp.*")
	if err != nil {
		logError("saveTracks: failed to create temporary file: %v", err)
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up if we fail

	// Write to the temporary file
	logDebug("saveTracks: writing to temporary file: %s", tmpPath)
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		logError("saveTracks: failed to write to temporary file %s: %v", tmpPath, err)
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Ensure the file is written to disk
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		logError("saveTracks: failed to sync temporary file %s: %v", tmpPath, err)
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	// Close the file before renaming
	if err := tmpFile.Close(); err != nil {
		logError("saveTracks: failed to close temporary file %s: %v", tmpPath, err)
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Rename the temporary file to the target file
	logDebug("saveTracks: renaming %s to %s", tmpPath, t.path)
	if err := os.Rename(tmpPath, t.path); err != nil {
		logError("saveTracks: failed to rename %s to %s: %v", tmpPath, t.path, err)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	logInfo("saveTracks: successfully saved %d tracks to %s", len(t.tracks), t.path)
	return nil
}

// Rebuild rescans the download directory and rebuilds the tracks file
func (t *Tracks) Rebuild(downloadDir string) error {
	// TODO: Implement directory scanning and track rebuilding
	return nil
}
