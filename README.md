<div align="center">
  <h1>:musical_note: ytpl</h1>
  <p>
    <a href="README.md">English</a> | <a href="README.ja.md">日本語</a>
  </p>
  <p>A command-line YouTube music player with playlist management</p>
  <p>
    <a href="#features">Features</a> •
    <a href="#installation">Installation</a> •
    <a href="#usage">Usage</a> •
    <a href="#configuration">Configuration</a> •
    <a href="#license">License</a>
  </p>
</div>

## ✨ Features

- Search and play YouTube music directly from the terminal
- Download and manage music locally
- Create and manage playlists
- Simple and intuitive interface
- Lightweight and fast

## 🚀 Installation

### Prerequisites

- Go 1.24 or later
- yt-dlp
- mpv player

### Using go install

```bash
go install github.com/256x/ytpl@latest
```

### Manual Build

```bash
git clone https://github.com/256x/ytpl.git
cd ytpl
go build -o ytpl .
sudo mv ytpl /usr/local/bin/
```

## 🎮 Usage

### Local Playback

The `play` command is used to play songs that have already been downloaded to your local storage. It supports fuzzy search to find matching tracks.

### Search and Play

```bash
# Search for a song (supports any search terms that work on YouTube)
ytpl search "artist name song title"

# Play a locally downloaded song (fuzzy search available)
ytpl play "artist name song title"

# Play a specific locally downloaded file by exact name
ytpl play "exact_file_name"
```

### Shuffle Play

```bash
# Shuffle and play all locally downloaded songs
ytpl shuffle

# Shuffle and play songs matching a search term
ytpl shuffle "search term"
```

### Playlist Management

```bash
# Create a new playlist
ytpl list create my_playlist

# Add a song to a playlist
ytpl list add my_playlist VIDEO_ID

# List all playlists
ytpl list

# Play a playlist
ytpl list play my_playlist
```

### Player Controls

```bash
# Play/Pause
ytpl pause

# Resume playback
ytpl resume

# Stop
ytpl stop

# Next track
ytpl next

# Previous track
ytpl prev

# Volume control
ytpl vol 80
```

## ⚙️ Configuration

Configuration file is located at `~/.config/ytpl/config.toml`.

Example configuration:

```toml
# Directory to store downloaded YouTube audio files
download_dir = "~/.local/share/ytpl/mp3/"

# Path to the media player (mpv)
player_path = "mpv"

# Default volume level (0-100)
default_volume = 80

# Path to yt-dlp executable
yt_dlp_path = "yt-dlp"

# Directory to store playlist files
playlist_dir = "~/.local/share/ytpl/playlists/"

# Browser to use for cookies (optional)
# cookie_browser = "firefox"

# Maximum number of search results
max_search_results = 15
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">
  <p>Made with ❤️ by <a href="https://github.com/256x">256x</a></p>
</div>
