# Changelog

All notable changes to this project will be documented in this file.

## [0.1.4] - 2025-06-03

### Added
- Added new `edit` command for modifying track titles
- Added `.info.json` optimization to reduce file size
- Added support for editing track metadata (title, artist, album, etc.)

### Changed
- Refactored track metadata handling to use `.tracks` file exclusively
- Improved performance by removing redundant MP3 tag reads
- Enhanced metadata consistency across the application
- Updated fuzzy finder prompt formatting for better readability

## [Unreleased]

### Changed
- Removed debug logs for cleaner release build
- Simplified error handling for state saving operations
- Removed unnecessary newlines in search and download output
- Improved code consistency in error handling patterns

## [0.1.3] - 2025-06-01

### Added
- Improved playlist display: Removed duration display for a cleaner interface
- Enhanced error message when a playlist doesn't exist
- Updated playlist prompt to `[ play from playlist_name ] >` format

### Changed
- Improved formatting of empty playlist messages

### Fixed
- Enhanced error handling for non-existent playlists

## [0.1.2] - 2025-06-01

### Added
- Initial implementation of playlist functionality
- Basic playlist creation, display, and deletion features

## [0.1.1] - 2025-06-01

### Added
- Basic playback functionality
- Search and download features

## [0.1.0] - 2025-06-01

### Added
- Initial project setup
- Basic command-line interface implementation
