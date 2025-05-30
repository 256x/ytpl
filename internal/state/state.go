// internal/state/state.go
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	config "ytpl/internal/config" // Alias for internal/config
)

// PlayerState holds the current state of the application's player.
type PlayerState struct {
	PID                  int    `json:"pid"`
	IPCSocketPath        string `json:"ipc_socket_path"`
	CurrentTrackID       string `json:"current_track_id"`
	CurrentTrackTitle    string `json:"current_track_title"`
	CurrentPlaylist      string `json:"current_playlist"`
	IsPlaying            bool   `json:"is_playing"` // true: playing, false: paused
	Volume               int    `json:"volume"`
	DownloadedFilePath   string `json:"downloaded_file_path"`
	LastPlayedTrackIndex int    `json:"last_played_track_index"` // For playlist continuation
	PlaybackHistory      []string `json:"playback_history"`        // For 'shuffle' or 'next' tracking
	ShuffleQueue         []string `json:"shuffle_queue"`           // For shuffle mode
	mu                   sync.Mutex // Mutex for concurrent access
}

var (
	stateFilePath string
	currentState  *PlayerState // Global instance of the state
)

// LoadState loads the application state from the state file.
// It also sets the state file path using the config.
func LoadState(cfg *config.Config) (*PlayerState, error) {
	var err error
	stateFilePath, err = config.GetStatePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get state file path: %w", err)
	}

	currentState = &PlayerState{} // Initialize with default values

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// State file doesn't exist, return empty state and no error
			return currentState, nil
		}
		return nil, fmt.Errorf("failed to read state file %s: %w", stateFilePath, err)
	}

	if err := json.Unmarshal(data, currentState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state data from %s: %w", stateFilePath, err)
	}

	// Ensure IPCSocketPath is updated if config changes or it's empty
	if currentState.IPCSocketPath == "" || currentState.IPCSocketPath != cfg.PlayerIPCSocketPath {
		currentState.IPCSocketPath = cfg.PlayerIPCSocketPath
	}

	return currentState, nil
}

// SaveState saves the current application state to the state file.
func SaveState() error {
	if currentState == nil {
		return fmt.Errorf("application state not initialized")
	}

	currentState.mu.Lock()
	defer currentState.mu.Unlock()

	data, err := json.MarshalIndent(currentState, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state data: %w", err)
	}

	// Ensure directory exists before writing
	if err := os.MkdirAll(filepath.Dir(stateFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create state directory %s: %w", filepath.Dir(stateFilePath), err)
	}

	if err := os.WriteFile(stateFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file %s: %w", stateFilePath, err)
	}
	return nil
}

// GetState returns the current application state.
// This should be called after LoadState.
func GetState() *PlayerState {
	return currentState
}

// UpdateAndSave updates the state with the provided values and saves it.
func UpdateAndSave(updater func(*PlayerState)) error {
	currentState.mu.Lock()
	defer currentState.mu.Unlock()
	updater(currentState)
	return SaveState()
}
