# ytpl - Command Line YouTube Music Player

[日本語版はこちら](README.md)

ytpl is a YouTube music player that runs on the command line. It allows you to download music from YouTube and manage/play it locally [1]. It has a playlist function, allowing you to organize and play your favorite songs [1].

## Main Features [1]

- Download and play music from YouTube [1]
- Manage local tracks [1]
- Create and manage playlists [1]
- Shuffle playback [1]
- Check and control playback status [1]

## Installation [1]

### Prerequisites [1]

- Go 1.16 or later [1]
- MPV player [1]
- YouTube-DL or yt-dlp [1]

### Installation Steps [1]

#### Using a Released Binary [1]

Download the binary suitable for your OS and architecture from the [releases page](https://github.com/256x/ytpl/releases), make it executable, and place it in a directory included in your PATH [1].

```bash
# Example: For Linux amd64 [2]
wget https://github.com/256x/ytpl/releases/download/vX.Y.Z/ytpl_linux_amd64 [2]
chmod +x ytpl_linux_amd64 [2]
sudo mv ytpl_linux_amd64 /usr/local/bin/ytpl [2]

Using Go
In an environment where Go is installed, you can install with the following command
:

# For Go 1.16 or later [2]
go install github.com/256x/ytpl@latest [2]

The go install command installs the binary to $GOPATH/bin or $GOBIN. Ensure these directories are included in your PATH
.
Building from Source

# Clone the repository [3]
git clone https://github.com/256x/ytpl.git [3]
cd ytpl [3]
# Get dependencies [3]
go mod download [3]
# Build [3]
go build -o ytpl [3]
# Place the binary in a directory included in PATH [3]
sudo mv ytpl /usr/local/bin/ [3]

Usage
Basic Commands

# Display help [3]
ytpl --help [3]
# Search for music on YouTube and play [3]
ytpl search "search query" [3]
# Example: ytpl search "Artist Name Song Title" [3]
# Example: ytpl search "Song Title Cover" [3]
# Example: ytpl search "Artist Name Album Name Song Title" [3]
# Example: ytpl search "https://www.youtube.com/watch?v=VideoID" [3]
# You can freely enter keywords or URLs, similar to YouTube search. [3]
# Display current playback status [3]
ytpl status [3]
# Play locally saved songs [3]
# ytpl play # Display list of local songs and interactively search/select to play [3]
# ytpl play "search query" [3]
# Example: ytpl play "Artist Name" # Search by artist name and play [4]
# Example: ytpl play "Song Title" # Search by song title and play [4]
# Pause playback [4]
ytpl pause [4]
# Resume playback [4]
ytpl resume [4]
# Stop playback [4]
ytpl stop [4]
# Skip to the next song [4]
ytpl next [4]
# Go back to the previous song [4]
ytpl prev [4]
# Shuffle play all locally saved songs [4]
ytpl shuffle [4]

Playlist Management

# Interact with playlists interactively (if no subcommand is specified) [4]
ytpl list [4]
# Create a new playlist [4]
ytpl list make myplaylist [4]
# Add the currently playing song to a playlist [4]
# If the specified playlist does not exist, a new one will be created [4]
ytpl list add myplaylist [4]
# Remove the currently playing song from a playlist [4]
ytpl list remove myplaylist [4]
# Delete a playlist [4]
ytpl list del myplaylist [4]
# Display the contents of a playlist [4]
ytpl list show myplaylist [4]
# Play a playlist [4]
ytpl list play myplaylist [4]
# Shuffle play a playlist [5]
ytpl list shuffle myplaylist [5]

Track Management

# Delete downloaded tracks [5]
# Deleted tracks will automatically be removed from all playlists as well [5]
ytpl delete [5]
# Adjust volume (0-100) [5]
ytpl volume 80 [5]

Configuration
The configuration file is saved at ~/.config/ytpl/config.toml
. The following configuration items are available, each with a default value set
:

# Directory to save YouTube audio files [5]
# Environment variables like $HOME can be used [5]
download_dir = "$HOME/.local/share/ytpl/mp3/" [5]
# Path to the media player (mpv is recommended) [5]
player_path = "mpv" [5]
# Path to MPV's IPC (Inter-Process Communication) socket [5]
player_ipc_socket_path = "/tmp/ytpl-mpv-socket" [5]
# Default volume (0-100) [6]
default_volume = 80 [6]
# Path to yt-dlp [6]
yt_dlp_path = "yt-dlp" [6]
# Directory to save playlists [6]
playlist_dir = "$HOME/.local/share/ytpl/playlists/" [6]
# Browser to load cookies from (e.g., "chrome", "firefox", "chromium", "brave", "edge") [6]
cookie_browser = "chrome" [6]
# Browser profile name (usually not needed) [6]
# cookie_profile = "" [6]
# Maximum number of search results to fetch from YouTube [6]
max_search_results = 30 [6]

Explanation of Main Configuration Items
•
download_dir: Directory for saving downloaded tracks
•
player_path: Path to the MPV player (if "mpv" is specified, it must be in your PATH)
•
player_ipc_socket_path: Path to the IPC socket used for controlling MPV
•
default_volume: Default volume upon startup (0-100)
•
yt_dlp_path: Path to yt-dlp (default: "yt-dlp")
•
playlist_dir: Directory for saving playlists (default: "$HOME/.local/share/ytpl/playlists/")
•
cookie_browser: Specify the browser to load cookies from (necessary for downloading videos requiring login, default: "firefox")
•
max_search_results: Maximum number of search results to display
License
This project is released under the MIT License. Refer to the LICENSE file for details
.
Notes
•
Please use this software in compliance with YouTube's terms of service
.
•
Downloaded content should be limited to personal use purposes
.
•
If downloading a large number of tracks, be mindful of YouTube's rate limits
.