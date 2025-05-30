# YouTube Playlist Manager (ytpl)

[日本語版はこちら](README.md)

A command-line YouTube music player and playlist manager that allows you to download, manage, and play music from YouTube.

## Features

- Search and play YouTube videos
- Download and manage local music library
- Create and manage custom playlists
- Cross-platform support (Linux, macOS, Windows)
- Simple and intuitive command-line interface

## Installation

### Prerequisites

- Go 1.24 or later
- yt-dlp (for downloading)
- MPV player (for playback)
- FFmpeg (for audio conversion)

### Binary Installation

Download the latest release for your platform from the [Releases](https://github.com/256x/ytpl/releases) page.

### Build from Source

```bash
git clone https://github.com/256x/ytpl.git
cd ytpl
go build -o ytpl
sudo mv ytpl /usr/local/bin/
```

## Configuration

Create a config file at `~/.config/ytpl/config.toml` with the following content:

```toml
# Directory to store downloaded music
download_dir = "$HOME/.local/share/ytpl/mp3/"

# Path to the music player (mpv recommended)
player_path = "mpv"

# Path to the MPV IPC socket
player_ipc_socket_path = "/tmp/ytpl-mpv-socket"

# Default volume (0-100)
default_volume = 80

# Path to yt-dlp executable
yt_dlp_path = "yt-dlp"

# Directory to store playlists
playlist_dir = "$HOME/.local/share/ytpl/playlists/"

# Browser to load cookies from (e.g., "chrome", "firefox", "chromium", "brave", "edge")
cookie_browser = "chrome"

# (Optional) Browser profile name
# Usually not needed for default profiles
# cookie_profile = ""

# (Optional) Browser profile directory path
# cookie_profile_dir = ""

# Maximum number of search results to display
max_search_results = 30
```

## Usage

### Search and Play

```bash
# Search and play a song
ytpl play "search query"

# Play without arguments to show a list of local tracks
ytpl play
```

### Playlist Management

```bash
# List all playlists
ytpl playlist list

# Create a new playlist
ytpl playlist create myplaylist

# Add a track to a playlist
ytpl playlist add myplaylist "search query"

# Remove a track from a playlist
ytpl playlist remove myplaylist 1

# Play a playlist
ytpl playlist play myplaylist

# Shuffle and play a playlist
ytpl playlist shuffle myplaylist
```

### Download Music

```bash
# Download a song
ytpl download "search query"

# Download and add to a playlist
ytpl download "search query" --playlist myplaylist
```

### Player Controls

```bash
# Pause/Resume
ytpl pause

# Stop
ytpl stop

# Skip to next track
ytpl next

# Set volume (0-100)
ytpl volume 80
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

256x

---

[View in Japanese](README.md)
