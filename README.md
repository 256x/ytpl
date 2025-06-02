# ytpl - CLI YouTube Music Player and Playlist Manager
[日本語版](README_jp.md)

`ytpl` is a command-line YouTube music player that allows you to download music from YouTube and manage/play it locally. It features playlist functionality to organize and play your favorite tracks.

## Key Features

- Download and play music from YouTube
- Manage local tracks
- Create and manage playlists
- Edit track metadata (titles, artists, etc.)
- Shuffle playback
- Check and control playback status

## Installation

### Prerequisites

- Go 1.16 or higher
- MPV player
- yt-dlp

### Installation Steps

#### Using Pre-built Binaries

Download the binary for your OS and architecture from the [releases page](https://github.com/256x/ytpl/releases), make it executable, and place it in a directory in your PATH.

```bash
# Example: For Linux amd64
wget https://github.com/256x/ytpl/releases/download/vX.Y.Z/ytpl_linux_amd64
chmod +x ytpl_linux_amd64
sudo mv ytpl_linux_amd64 /usr/local/bin/ytpl
```

#### Using Go

If you have Go installed, you can install using the following command:

```bash
# For Go 1.16 and later
go install github.com/256x/ytpl@latest
```

The `go install` command installs the binary to `$GOPATH/bin` or `$GOBIN`. Make sure these directories are included in your `PATH`.

#### Building from Source

```bash
# Clone the repository
git clone https://github.com/256x/ytpl.git
cd ytpl

# Download dependencies
go mod download

# Build
go build -o ytpl

# Move binary to a directory in PATH
sudo mv ytpl /usr/local/bin/
```

## Usage

### Basic Commands

```
# Display help
ytpl --help

# Search for music on YouTube and download/play
ytpl search [query]
# Examples:
# ytpl search "Artist Name Song Title"     # Search by artist and song
# ytpl search "Song Title Cover"           # Search for cover songs
# ytpl search "Artist Name Album Name"     # Search from album
# ytpl search "https://youtube.com/..."    # Play directly from YouTube URL
# ytpl search "Playlist Name"              # Search and play playlist
# ytpl search "Live Song Title"            # Search for live recordings
# ytpl search "Cover Song Title"           # Search for cover videos

# Edit track metadata (title, artist, etc.)
ytpl edit [query]
# Examples:
# ytpl edit                  # Interactive track selection
# ytpl edit "Song Title"    # Search and edit specific track

# Play locally saved tracks
ytpl play [query]
# Examples:
# ytpl play                     # Interactive track selection
# ytpl play "Artist Name"       # Search by artist name
# ytpl play "Song Title"        # Search by song title
# ytpl play "Album Name"        # Search by album name

# Manage and play playlists
ytpl list

# Play playlist
# ytpl list play <playlist_name>    # Play in order
# ytpl list shuffle <playlist_name> # Shuffle play

# Create and manage playlists
# ytpl list make <playlist_name>     # Create new playlist
# ytpl list add <playlist> <track_id> # Add to playlist
# ytpl list remove <playlist> <track_id> # Remove from playlist
# ytpl list delete <playlist>        # Delete playlist


# Shuffle play all local tracks
ytpl shuffle

# Display current playback status
ytpl status

# Playback controls
ytpl play [query]  # Play locally saved tracks
# Examples:
# ytpl play
# ytpl play "Artist Name Song Title" # Fuzzy search with list display

ytpl pause   # Pause playback
ytpl resume  # Resume from pause
ytpl stop    # Stop playback
ytpl next    # Skip to next track
ytpl prev    # Go back to previous track
ytpl volume <0-100>  # Set volume (0-100)

# Delete tracks from local storage
ytpl delete [query]

# Display version information
ytpl --version or ytpl -v
```

### Playlist Management

```
# Interactive playlist operations (when no subcommand is specified)
ytpl list

# Create a new playlist
ytpl list create MyPlaylist

# Add currently playing track to playlist
# If the specified playlist doesn't exist, it will be created
ytpl list add MyPlaylist

# Remove currently playing track from playlist
ytpl list remove MyPlaylist

# Delete playlist
ytpl list del MyPlaylist

# Show playlist contents
ytpl list show MyPlaylist

# Play playlist (in order)
ytpl list play MyPlaylist

# Shuffle play playlist
ytpl list shuffle MyPlaylist
```

### Track Management

```
# Delete downloaded tracks
# Deleted tracks are automatically removed from all playlists
ytpl delete

# Edit track metadata (title, artist, etc.)
ytpl edit [query]

# Adjust volume (0-100)
ytpl volume 80
```

## Configuration

The configuration file is saved at `~/.config/ytpl/config.toml`. The following configuration options are available, each with default values:

```toml
# Directory to save YouTube audio files
# Environment variables like $HOME can be used
download_dir = "$HOME/.local/share/ytpl/mp3/"

# Media player path (mpv is recommended)
player_path = "mpv"

# MPV IPC (Inter-Process Communication) socket path
player_ipc_socket_path = "/tmp/ytpl-mpv-socket"

# Default volume (0-100)
default_volume = 80

# yt-dlp path
yt_dlp_path = "yt-dlp"

# Directory to save playlists
playlist_dir = "$HOME/.local/share/ytpl/playlists/"

# Browser to load cookies from (e.g., "chrome", "firefox", "chromium", "brave", "edge")
cookie_browser = "chrome"

# Browser profile name (usually not needed)
# cookie_profile = ""

# Maximum number of search results to fetch from YouTube
max_search_results = 30
```

### Main Configuration Options Explained

- `download_dir`: Directory to save downloaded tracks
- `player_path`: Path to MPV player (specifying `mpv` requires it to be in PATH)
- `player_ipc_socket_path`: IPC socket path used for MPV control
- `default_volume`: Default volume at startup (0-100)
- `yt_dlp_path`: Path to yt-dlp (default: "yt-dlp")
- `playlist_dir`: Directory to save playlists (default: "$HOME/.local/share/ytpl/playlists/")
- `cookie_browser`: Specify browser to load cookies from (needed for downloading videos that require login, default: "firefox")
- `max_search_results`: Maximum number of search results to display

## License

This project is released under the MIT License. See the [LICENSE](LICENSE) file for details.

## Important Notes

- Please use this software in compliance with YouTube's Terms of Service.
- Downloaded content should be limited to personal use only.
- When downloading large numbers of tracks, be mindful of YouTube's rate limits.
