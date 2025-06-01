// internal/config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

// Config holds the application configuration.
type Config struct {
	DownloadDir         string `toml:"download_dir"`
	PlayerPath          string `toml:"player_path"`
	PlayerIPCSocketPath string `toml:"player_ipc_socket_path"`
	DefaultVolume       int    `toml:"default_volume"`
	YtDlpPath           string `toml:"yt_dlp_path"`
	PlaylistDir         string `toml:"playlist_dir"`
	CookieBrowser       string `toml:"cookie_browser"`
	CookieProfile       string `toml:"cookie_profile"`
	MaxSearchResults    int    `toml:"max_search_results"`
}

const (
	configFileName = "config.toml"
	stateFileName  = "state.json"
	appName        = "ytpl"
)

// LoadConfig loads the application configuration from config.toml.
// It also sets default paths based on XDG Base Directory Specification if not specified.
// It expands environment variables in path strings.
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	configPath, err := xdg.ConfigFile(filepath.Join(appName, configFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path: %w", err)
	}

	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to decode config file %s: %w", configPath, err)
		}
		log.Printf("config file not found at %s. using default settings.", configPath)
	}

	// Expand environment variables for all user-configurable paths
	cfg.DownloadDir = os.ExpandEnv(cfg.DownloadDir)
	cfg.PlayerIPCSocketPath = os.ExpandEnv(cfg.PlayerIPCSocketPath)
	cfg.PlaylistDir = os.ExpandEnv(cfg.PlaylistDir)
	cfg.PlayerPath = os.ExpandEnv(cfg.PlayerPath)
	cfg.YtDlpPath = os.ExpandEnv(cfg.YtDlpPath)
	cfg.CookieBrowser = os.ExpandEnv(cfg.CookieBrowser)
	cfg.CookieProfile = os.ExpandEnv(cfg.CookieProfile)

	// Set default values if not provided
	if cfg.DownloadDir == "" {
		dataDir, err := xdg.DataFile(filepath.Join(appName, "stock"))
		if err != nil {
			log.Printf("warning: could not determine default download directory based on xdg. using temp dir as fallback: %v", err)
			cfg.DownloadDir = filepath.Join(os.TempDir(), appName, "stock")
		} else {
			cfg.DownloadDir = dataDir
		}
	}
	if cfg.PlayerPath == "" {
		cfg.PlayerPath = "mpv"
	}
	if cfg.PlayerIPCSocketPath == "" {
		if xdg.RuntimeDir != "" {
			cfg.PlayerIPCSocketPath = filepath.Join(xdg.RuntimeDir, appName, "mpv-socket")
		} else {
			cfg.PlayerIPCSocketPath = filepath.Join(os.TempDir(), appName+"-mpv-socket")
		}
	}
	if cfg.DefaultVolume == 0 {
		cfg.DefaultVolume = 80
	}
	if cfg.YtDlpPath == "" {
		cfg.YtDlpPath = "yt-dlp"
	}
	if cfg.PlaylistDir == "" {
		playlistDefaultDir, err := xdg.DataFile(filepath.Join(appName, "playlists"))
		if err != nil {
			log.Printf("warning: could not determine default playlist directory based on xdg. using temp dir as fallback: %v", err)
			cfg.PlaylistDir = filepath.Join(os.TempDir(), appName, "playlists")
		} else {
			cfg.PlaylistDir = playlistDefaultDir
		}
	}
	// Set default for MaxSearchResults
	if cfg.MaxSearchResults == 0 { // If 0 or not set, default to 10
		cfg.MaxSearchResults = 10
	} else if cfg.MaxSearchResults < 1 { // Ensure it's at least 1
		cfg.MaxSearchResults = 1
	}

	// Ensure all necessary directories exist
	if err := os.MkdirAll(cfg.DownloadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory %s: %w", cfg.DownloadDir, err)
	}
	if err := os.MkdirAll(cfg.PlaylistDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create playlist directory %s: %w", cfg.PlaylistDir, err)
	}

	statePath, err := GetStatePath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory %s: %w", filepath.Dir(statePath), err)
	}

	if !filepath.IsAbs(cfg.PlayerIPCSocketPath) || !strings.HasPrefix(filepath.Clean(cfg.PlayerIPCSocketPath), filepath.Clean(os.TempDir())) {
		if err := os.MkdirAll(filepath.Dir(cfg.PlayerIPCSocketPath), 0755); err != nil {
			log.Printf("warning: could not create directory for player ipc socket %s: %v", filepath.Dir(cfg.PlayerIPCSocketPath), err)
		}
	}

	return cfg, nil
}

// GetConfigPath returns the expected path for the config file.
func GetConfigPath() (string, error) {
	return xdg.ConfigFile(filepath.Join(appName, configFileName))
}

// GetStatePath returns the expected path for the state file.
func GetStatePath() (string, error) {
	return xdg.StateFile(filepath.Join(appName, stateFileName))
}

// GetDefaultConfigContent returns a string with default config.toml content.
func GetDefaultConfigContent() string {
	return `
# ~/.config/ytpl/config.toml

# Directory to store downloaded YouTube audio files.
# Environment variables like $HOME are expanded.
download_dir = "$HOME/.local/share/ytpl/mp3/"

# Path to the media player executable (e.g., mpv).
# Garia is designed to work with mpv's IPC features.
player_path = "mpv"

# Path for mpv's IPC (Inter-Process Communication) socket.
# Used for controlling mpv.
player_ipc_socket_path = "/tmp/ytpl-mpv-socket"

# Default volume level (0-100).
default_volume = 80

# Path to yt-dlp executable.
yt_dlp_path = "yt-dlp"

# Directory to store playlist files.
playlist_dir = "$HOME/.local/share/ytpl/playlists/"

# Cookie settings for downloading age-restricted YouTube videos.
# Specify the browser from which to load cookies (e.g., "chrome", "firefox", "chromium", "brave", "edge").
# Refer to yt-dlp documentation for supported browsers: https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp
cookie_browser = "chrome"
# (Optional) Browser profile name.
# Usually not needed for default profiles, but specify if you use multiple profiles.
# e.g., "Default", "Profile 1", "Default-Browser-Profile"
# cookie_profile = ""

# Maximum number of search results to retrieve from YouTube.
max_search_results = 30
`
}
