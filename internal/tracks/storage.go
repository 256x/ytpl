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
	// Ensure directory exists
	dir := filepath.Dir(t.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal tracks to JSON
	data, err := json.MarshalIndent(t.tracks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tracks: %w", err)
	}

	// Create a temporary file in the same directory
	tmpFile, err := os.CreateTemp(dir, ".tracks.tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up if we fail

	// Write to the temporary file
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Ensure the file is written to disk
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	// Close the file before renaming
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Rename the temporary file to the target file
	if err := os.Rename(tmpPath, t.path); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}
	return nil
}

// Rebuild rescans the download directory and rebuilds the tracks file
func (t *Tracks) Rebuild(downloadDir string) error {
	// TODO: Implement directory scanning and track rebuilding
	return nil
}
