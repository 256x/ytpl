# ytpl - Command Line YouTube Music Player

[日本語版はこちら](README_jp.md)

`ytpl` is a command line YouTube music player. It allows you to download music from YouTube and manage and play it locally. It has a playlist function, allowing you to organize and play your favorite songs.

## Main Features

- Downloading and playing music from YouTube
- Managing local tracks
- Creating and managing playlists
- Shuffle playback
- Checking and controlling playback status

## Installation

### Prerequisites

- Go 1.16 or later
- MPV player
- yt-dlp

### Installation Steps

#### Using Pre-built Binaries

Download the binary suitable for your OS and architecture from the [releases page](https://github.com/256x/ytpl/releases), make it executable, and place it in a directory included in your PATH.

```bash
# Example: For Linux amd64
wget https://github.com/256x/ytpl/releases/download/vX.Y.Z/ytpl_linux_amd64
chmod +x ytpl_linux_amd64
sudo mv ytpl_linux_amd64 /usr/local/bin/ytpl
```
Using Go
If you have Go installed, you can install it with the following command:
```
# For Go 1.16 or later
go install github.com/256x/ytpl@latest
```
The go install command installs the binary to $GOPATH/bin or $GOBIN. Make sure these directories are included in your PATH.
Building from Source
```
# Clone the repository
git clone https://github.com/256x/ytpl.git
cd ytpl
# Get dependencies
go mod download
# Build
go build -o ytpl
# Place the binary in a directory included in PATH
sudo mv ytpl /usr/local/bin/
```
Usage
Basic Commands
```
# Display help
ytpl --help
# Search for and play music on YouTube
ytpl search "search query"
# Example: ytpl search "Artist Name Song Title"
# Example: ytpl search "Song Title Cover"
# Example: ytpl search "Artist Name Album Name Song Title"
# Example: ytpl search "https://www.youtube.com/watch?v=動画ID"
# You can freely enter keywords or URLs, similar to YouTube search.
# Display current playback status
ytpl status
# Play locally saved songs
# ytpl play # Display a list of local songs, interactively search/select and play
# ytpl play "search query"
# Example: ytpl play "Artist Name" # Search by artist name and play
# Example: ytpl play "Song Title" # Search by song title and play
# Pause playback
ytpl pause
# Resume playback
ytpl resume
# Stop playback
ytpl stop
# Skip to the next song
ytpl next
# Go back to the previous song
ytpl prev
# Shuffle play all locally saved songs
ytpl shuffle
```
Playlist Management
```
# Interact with playlists (if no subcommand is specified)
ytpl list
# Create a new playlist
ytpl list make myplaylist
# Add the currently playing song to a playlist
# If the specified playlist does not exist, a new one will be created
ytpl list add myplaylist
# Remove the currently playing song from a playlist
ytpl list remove myplaylist
# Delete a playlist
ytpl list del myplaylist
# Display the contents of a playlist
ytpl list show myplaylist
# Play a playlist
ytpl list play myplaylist
# Shuffle play a playlist
ytpl list shuffle myplaylist
```
Track Management
```
# Delete downloaded tracks
# Deleted tracks are automatically removed from all playlists as well
ytpl delete
# Adjust volume (0-100)
ytpl volume 80
```
Configuration
The configuration file is saved at ~/.config/ytpl/config.toml. The following configuration items are available, each with a default value:
```
# Directory to save YouTube audio files
# Environment variables like $HOME can be used
download_dir = "$HOME/.local/share/ytpl/mp3/"
# Path to the media player (mpv is recommended)
player_path = "mpv"
# Path to MPV's IPC (Inter-Process Communication) socket
player_ipc_socket_path = "/tmp/ytpl-mpv-socket"
# Default volume (0-100)
default_volume = 80
# Path to yt-dlp
yt_dlp_path = "yt-dlp"
# Directory to save playlists
playlist_dir = "$HOME/.local/share/ytpl/playlists/"
# Browser to load cookies from (e.g., "chrome", "firefox", "chromium", "brave", "edge")
cookie_browser = "chrome"
# Browser profile name (usually not needed)
# cookie_profile = ""
# Maximum number of search results to get from YouTube
max_search_results = 30
```
Description of Main Configuration Items
- download_dir: Directory for saving downloaded tracks
- player_path: Path to the MPV player (mpv means it must be in the PATH)
- player_ipc_socket_path: Path to the IPC socket used to control MPV
- default_volume: Default volume upon startup (0-100)
- yt_dlp_path: Path to yt-dlp (default: "yt-dlp")
- playlist_dir: Directory for saving playlists (default: "$HOME/.local/share/ytpl/playlists/")
- cookie_browser: Specifies the browser to load cookies from (needed for downloading videos that require login, default: "firefox")
- max_search_results: Maximum number of search results to display

License
This project is released under the MIT License. See the LICENSE file for details.

Notes
- Please use this software in compliance with YouTube's Terms of Service.
- Downloaded content should be limited to personal use.
- If downloading a large number of tracks, be mindful of YouTube's rate limits.
