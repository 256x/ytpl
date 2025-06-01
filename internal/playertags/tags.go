// internal/playertags/tags.go
package playertags

import (
	"fmt"
	"os"

	"github.com/dhowden/tag" // Library for reading MP3 tags
)

// AudioInfo represents extracted audio metadata.
type AudioInfo struct {
	Title  string
	Artist string
	Album  string
	Year   int
}

// ReadTagsFromMP3 reads ID3 tags from an MP3 file.
// It also takes fallbackTitle and fallbackArtist to use if no tags are found.
func ReadTagsFromMP3(filePath string, fallbackTitle, fallbackArtist string) (*AudioInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	m, err := tag.ReadFrom(file)
	if err != nil {
		// If tags cannot be read, use fallback info provided by caller
		info := &AudioInfo{
			Title:  fallbackTitle,
			Artist: fallbackArtist,
			Album:  "", // No album fallback
			Year:   0,  // No year fallback
		}
		// Final fallback for empty info
		if info.Title == "" {
			info.Title = "unknown title"
		}
		if info.Artist == "" {
			info.Artist = "unknown artist"
		}
		return info, fmt.Errorf("failed to read tags from %s: %w", filePath, err) // Return error but provide fallback info
	}

	info := &AudioInfo{
		Title:  m.Title(),
		Artist: m.Artist(),
		Album:  m.Album(),
		Year:   m.Year(),
	}

	// Provide fallback for empty tags even if ReadFrom succeeded but tags are missing
	if info.Title == "" {
		info.Title = fallbackTitle
		if info.Title == "" { // If fallback title is also empty
			info.Title = "unknown title"
		}
	}
	if info.Artist == "" {
		info.Artist = fallbackArtist
		if info.Artist == "" { // If fallback artist is also empty
			info.Artist = "unknown artist"
		}
	}

	return info, nil
}

// FormatAudioInfo is no longer needed here as display formatting is handled by the caller.
// (You can delete this function entirely or keep it for future use as a simple formatter)
/*
func FormatAudioInfo(info *AudioInfo) string {
	if info.Title == "" && info.Artist == "" {
		return "Unknown Track"
	}
	if info.Artist == "" || info.Artist == "Unknown Artist" {
		return info.Title
	}
	return fmt.Sprintf("%s - %s", info.Artist, info.Title)
}
*/
