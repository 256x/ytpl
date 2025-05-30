## v0.1.0 Initial Release

### New Features
- Search and play music from YouTube
- Playlist management
- Local music playback
- Shuffle playback

### Supported Platforms
- Linux (amd64, arm64)
- macOS (amd64, arm64)

### Installation

#### macOS
```bash
# amd64
curl -L https://github.com/256x/ytpl/releases/download/v0.1.0/ytpl-darwin-amd64 -o ytpl
chmod +x ytpl
sudo mv ytpl /usr/local/bin/

# Apple Silicon (arm64)
curl -L https://github.com/256x/ytpl/releases/download/v0.1.0/ytpl-darwin-arm64 -o ytpl
chmod +x ytpl
sudo mv ytpl /usr/local/bin/
```

#### Linux
```bash
# amd64
curl -L https://github.com/256x/ytpl/releases/download/v0.1.0/ytpl-linux-amd64 -o ytpl
chmod +x ytpl
sudo mv ytpl /usr/local/bin/

# arm64
curl -L https://github.com/256x/ytpl/releases/download/v0.1.0/ytpl-linux-arm64 -o ytpl
chmod +x ytpl
sudo mv ytpl /usr/local/bin/
```

#### Manual Build
```bash
git clone https://github.com/256x/ytpl.git
cd ytpl
go build -o ytpl .
sudo mv ytpl /usr/local/bin/
```

### Dependencies
- yt-dlp
- mpv

### Available Commands

#### Playback Control
```bash
ytpl play <query>    # Play a locally stocked song
ytpl pause          # Pause the current playback
ytpl resume         # Resume the paused song
ytpl stop           # Stop the current playback
ytpl next           # Play the next song in the queue
ytpl prev           # Play the previous song in the queue
ytpl vol <0-100>    # Set playback volume (0-100)
```

#### Playlist Management
```bash
ytpl list                          # List all playlists
ytpl list create <name>           # Create a new playlist
ytpl list add <playlist> <video>  # Add a video to playlist
ytpl list play <playlist>         # Play a playlist
```

#### Music Management
```bash
ytpl search <query>   # Search YouTube for music
ytpl del <query>      # Delete a downloaded track
ytpl shuffle          # Shuffle and play all local songs
```

### Getting Started
```bash
# Search and play a song
ytpl search "artist name song title"
ytpl play "artist name song title"

# Control playback
ytpl pause    # Pause
ytpl resume   # Resume

# Manage playlists
ytpl list create my_playlist
ytpl list add my_playlist VIDEO_ID
ytpl list play my_playlist
```

### Notes
- The `play` command is for playing locally downloaded files
- Use fuzzy search to find and play your music
- All YouTube search queries are supported

### License
MIT
