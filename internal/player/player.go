// internal/player/player.go
package player

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	config "ytpl/internal/config" // Alias for internal/config
	state "ytpl/internal/state"   // Alias for internal/state
)

// IPCCommand represents a command to be sent to mpv's IPC socket.
type IPCCommand struct {
	Command []interface{} `json:"command"`
}

// IPCResponse represents a response from mpv's IPC socket.
type IPCResponse struct {
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
	// Add other fields like "request_id" if using asynchronous requests
}

// StartPlayer starts the mpv player in the background.
// This function is for single file playback (e.g., from search result).
func StartPlayer(cfg *config.Config, s *state.PlayerState, filePath string) error {
	// If player is already running, stop it first.
	// This ensures only one mpv instance is controlled by ytpl for single playback.
	if s.PID != 0 {
		log.Printf("player already running (pid: %d). stopping it first...", s.PID)
		if err := StopPlayer(s); err != nil {
			log.Printf("warning: could not stop existing player: %v", err)
			// Continue, as the old process might be gone already
		}
	}

	// Use saved volume if available, otherwise use default volume
	volume := s.Volume
	if volume == 0 {
		volume = cfg.DefaultVolume
	}

	// Ensure the socket directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.PlayerIPCSocketPath), 0755); err != nil {
		return fmt.Errorf("failed to create ipc socket directory %s: %w", filepath.Dir(cfg.PlayerIPCSocketPath), err)
	}
	// Remove old socket file if it exists, to prevent "address already in use" errors
	os.Remove(cfg.PlayerIPCSocketPath)

	// mpv arguments for background playback and IPC
	args := []string{
		filePath,
		fmt.Sprintf("--input-ipc-server=%s", cfg.PlayerIPCSocketPath),
		"--no-terminal",     // Do not open a terminal window for mpv
		fmt.Sprintf("--volume=%d", volume), // Use saved volume
		"--idle=yes",        // Keep mpv running in idle mode when playlist ends or no file is given
		"--force-window=no", // Do not force window display (for audio-only)
		"--no-video",        // Explicitly disable video display
	}

	cmd := exec.Command(cfg.PlayerPath, args...)

	// Detach mpv process from the parent ytpl process
	// This makes mpv run in the background even if ytpl exits
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create a new process group
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	s.PID = cmd.Process.Pid
	s.IPCSocketPath = cfg.PlayerIPCSocketPath
	s.IsPlaying = true // Assume playing immediately after start

	// Wait a moment for mpv to start and create the socket
	time.Sleep(500 * time.Millisecond)

	// Verify socket is created and reachable
	conn, err := net.DialTimeout("unix", s.IPCSocketPath, 2*time.Second)
	if err != nil {
		log.Printf("warning: mpv ipc socket not ready or failed to connect: %v", err)
		// Optionally, kill the mpv process if socket connection fails consistently
	} else {
		conn.Close()
	}

	return nil
}

// SendCommand sends an ipc command to mpv.
func SendCommand(s *state.PlayerState, command []interface{}) error {
	if s.PID == 0 || s.IPCSocketPath == "" {
		return fmt.Errorf("player is not running or ipc socket path is unknown")
	}

	conn, err := net.DialTimeout("unix", s.IPCSocketPath, 1*time.Second)
	if err != nil {
		// Log this warning to the file, not console, to keep tui clean.
		log.Printf("warning: failed to connect to mpv ipc socket (pid %d): %v. assuming player is no longer running.\n", s.PID, err)
		s.PID = 0 // Clear pid if connection fails
		state.SaveState() // Save updated state
		return fmt.Errorf("player not reachable, possibly stopped")
	}
	defer conn.Close()

	ipcCmd := IPCCommand{Command: command}
	encoder := json.NewEncoder(conn)
	// Add newline delimiter for mpv ipc
	if err := encoder.Encode(ipcCmd); err != nil {
		return fmt.Errorf("failed to send command to mpv: %w", err)
	}

	return nil
}

// GetProperty fetches a property value from mpv.
func GetProperty(s *state.PlayerState, property string) (interface{}, error) {
	if s.PID == 0 || s.IPCSocketPath == "" {
		return nil, fmt.Errorf("player is not running or ipc socket path is unknown")
	}

	conn, err := net.DialTimeout("unix", s.IPCSocketPath, 1*time.Second)
	if err != nil {
		log.Printf("warning: failed to connect to mpv ipc socket (pid %d) for property '%s': %v. assuming player is no longer running.\n", s.PID, property, err)
		s.PID = 0
		state.SaveState()
		return nil, fmt.Errorf("player not reachable, possibly stopped")
	}
	defer conn.Close()

	ipcCmd := IPCCommand{Command: []interface{}{"get_property", property}}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(ipcCmd); err != nil {
		return nil, fmt.Errorf("failed to send get_property command to mpv: %w", err)
	}

	decoder := json.NewDecoder(conn)
	var resp IPCResponse
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response from mpv: %w", err)
	}

	if resp.Error != "success" {
		return nil, fmt.Errorf("mpv returned error for property '%s': %s", property, strings.ToLower(resp.Error))
	}

	return resp.Data, nil
}

// GetCurrentlyPlayingTrackInfo fetches the currently playing track's file path and playlist position from mpv.
// Returns filePath, playlistPosition, error.
func GetCurrentlyPlayingTrackInfo(s *state.PlayerState) (string, int, error) {
	if s.PID == 0 || s.IPCSocketPath == "" {
		return "", -1, fmt.Errorf("player is not running or ipc socket path is unknown")
	}

	// Get current file path
	filePath, err := GetProperty(s, "path") // "path" property gives the full path of the currently playing file
	if err != nil {
		return "", -1, fmt.Errorf("failed to get 'path' property from mpv: %w", err)
	}
	filePathStr, ok := filePath.(string)
	if !ok {
		return "", -1, fmt.Errorf("mpv 'path' property is not a string: %v", filePath)
	}

	// Get current playlist position
	playlistPos, err := GetProperty(s, "playlist-pos") // "playlist-pos" gives 0-indexed position
	if err != nil {
		// This might fail if not playing from a playlist, so log and default
		log.Printf("warning: could not get 'playlist-pos' from mpv: %v", err)
		return filePathStr, -1, nil // Return file path, but -1 for position
	}
	playlistPosInt, ok := playlistPos.(float64) // mpv returns numbers as float64 via json ipc
	if !ok {
		return filePathStr, -1, fmt.Errorf("mpv 'playlist-pos' property is not a number: %v", playlistPos)
	}

	return filePathStr, int(playlistPosInt), nil
}


// StopPlayer sends a quit command to mpv and cleans up state.
func StopPlayer(s *state.PlayerState) error {
	if s.PID == 0 {
		log.Printf("player is not running.")
		return nil
	}

	err := SendCommand(s, []interface{}{"quit"})
	if err != nil {
		log.Printf("warning: failed to send quit command to mpv via ipc (pid %d): %v. trying to kill process...", s.PID, err)

		// FindProcess(s.PID) should only be called if s.PID is valid
		process, procErr := os.FindProcess(s.PID)
		if procErr != nil {
			log.Printf("error: could not find mpv process with pid %d: %v. it might have already exited.", s.PID, procErr)
			// Process might already be gone, just clean up state.
		} else {
			// Check if process is still alive before killing
			// On Unix, signal 0 can be used to check if a process exists
			if process.Signal(syscall.Signal(0)) == nil { // Check if process exists (Unix-like)
				if killErr := process.Kill(); killErr != nil {
					return fmt.Errorf("failed to kill mpv process with pid %d: %w", s.PID, killErr)
				}
				log.Printf("killed mpv process with pid %d.", s.PID)
			} else {
				log.Printf("mpv process with pid %d already exited.", s.PID)
			}
		}
	} else {
		// Ipc quit command sent successfully, give mpv a moment to shut down gracefully
		time.Sleep(200 * time.Millisecond)
		log.Printf("sent quit command to mpv (pid %d).", s.PID)
	}

	// Clean up socket file
	if s.IPCSocketPath != "" {
		os.Remove(s.IPCSocketPath)
		log.Printf("removed ipc socket file: %s", s.IPCSocketPath)
	}

	// Clear player state AFTER attempting to stop process and clean up socket
	s.PID = 0
	s.CurrentTrackID = ""
	s.CurrentTrackTitle = ""
	s.DownloadedFilePath = ""
	s.IsPlaying = false
	s.CurrentPlaylist = ""
	s.LastPlayedTrackIndex = 0
	s.PlaybackHistory = []string{}
	s.ShuffleQueue = []string{}

	return state.SaveState() // Save the cleared state
}

// LoadFile loads a new file into the currently running mpv player.
// Use 'replace' mode to stop current playback and play new file.
// This is used for single track playback or manually switching.
func LoadFile(s *state.PlayerState, filePath string) error {
	if s.PID == 0 {
		return fmt.Errorf("player is not running. cannot load file.")
	}
	return SendCommand(s, []interface{}{"loadfile", filePath, "replace"})
}

// Next sends a 'playlist-next' command to mpv.
func Next(s *state.PlayerState) error {
	return SendCommand(s, []interface{}{"playlist-next"})
}

// Prev sends a 'playlist-prev' command to mpv.
func Prev(s *state.PlayerState) error {
	return SendCommand(s, []interface{}{"playlist-prev"})
}

// LoadPlaylistIntoPlayer loads a list of files into mpv as a playlist.
// This function starts a new mpv process with the entire playlist.
func LoadPlaylistIntoPlayer(cfg *config.Config, s *state.PlayerState, filePaths []string, startIndex int) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no files to load into playlist")
	}

	// If player is already running, stop it first (to clear old playlist/state)
	if s.PID != 0 {
		log.Printf("player already running (pid: %d). stopping it to load new playlist...", s.PID)
		if err := StopPlayer(s); err != nil {
			log.Printf("warning: could not stop existing player: %v", err)
		}
	}

	// Use saved volume if available, otherwise use default volume
	volume := s.Volume
	if volume == 0 {
		volume = cfg.DefaultVolume
	}

	// Ensure socket directory exists and remove old socket
	if err := os.MkdirAll(filepath.Dir(cfg.PlayerIPCSocketPath), 0755); err != nil {
		return fmt.Errorf("failed to create ipc socket directory %s: %w", filepath.Dir(cfg.PlayerIPCSocketPath), err)
	}
	os.Remove(cfg.PlayerIPCSocketPath)

	// Build mpv args for starting with a playlist
	// Start with --no-terminal, --volume, --idle, --input-ipc-server, --no-video
	baseArgs := []string{
		fmt.Sprintf("--input-ipc-server=%s", cfg.PlayerIPCSocketPath),
		"--no-terminal",
		fmt.Sprintf("--volume=%d", volume), // Use saved volume
		"--idle=yes",
		"--force-window=no", // Do not force window display for audio playback
		"--no-video",        // Explicitly disable video display
	}

	// Add each file to the arguments for mpv to treat as a playlist
	for _, p := range filePaths {
		baseArgs = append(baseArgs, p)
	}

	// For playback from specific index, mpv has --playlist-start=N
	// Note: mpv playlist-start is 0-indexed
	if startIndex >= 0 && startIndex < len(filePaths) {
		baseArgs = append(baseArgs, fmt.Sprintf("--playlist-start=%d", startIndex))
	} else if startIndex != 0 { // If startIndex is out of bounds but not 0, set to 0.
		baseArgs = append(baseArgs, "--playlist-start=0")
	}


	cmd := exec.Command(cfg.PlayerPath, baseArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create a new process group for mpv
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv with playlist: %w", err)
	}

	s.PID = cmd.Process.Pid
	s.IPCSocketPath = cfg.PlayerIPCSocketPath
	s.IsPlaying = true // Assume playing immediately after start

	time.Sleep(500 * time.Millisecond) // Give mpv time to start and create the socket

	// Verify socket is created and reachable
	conn, err := net.DialTimeout("unix", s.IPCSocketPath, 2*time.Second)
	if err != nil {
		log.Printf("warning: mpv ipc socket not ready or failed to connect after playlist load: %v\n", err)
		// Optionally, kill the mpv process if socket connection fails consistently
	} else {
		conn.Close()
	}

	return nil
}

// SetVolume sets the volume of the mpv player.
func SetVolume(s *state.PlayerState, volume int) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}
	err := SendCommand(s, []interface{}{"set_property", "volume", volume})
	if err == nil {
		s.Volume = volume
		state.SaveState()
	}
	return err
}

// Pause pauses the mpv player.
func Pause(s *state.PlayerState) error {
	err := SendCommand(s, []interface{}{"set_property", "pause", true})
	if err == nil {
		s.IsPlaying = false
		state.SaveState()
	}
	return err
}

// Resume resumes the mpv player.
func Resume(s *state.PlayerState) error {
	err := SendCommand(s, []interface{}{"set_property", "pause", false})
	if err == nil {
		s.IsPlaying = true
		state.SaveState()
	}
	return err
}
