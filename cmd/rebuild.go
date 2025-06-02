package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"ytpl/internal/tracks"
	"ytpl/internal/yt"

	"github.com/spf13/cobra"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild the .tracks file from existing downloads",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize track manager
		trackManager, err := tracks.NewManager("", cfg.DownloadDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize track manager: %v\n", err)
			os.Exit(1)
		}

		// Enable batch mode for better performance
		trackManager.BatchMode(true)
		defer trackManager.BatchMode(false) // Ensure batch mode is disabled when we're done

		// Clear existing tracks
		// Clearing existing tracks...
		if err := trackManager.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to clear tracks: %v\n", err)
			os.Exit(1)
		}

		// Scan download directory for MP3 files
		// Scanning download directory
		files, err := ioutil.ReadDir(cfg.DownloadDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read download directory: %v\n", err)
			os.Exit(1)
		}

		// Process files in batches
		batchSize := 50
		var wg sync.WaitGroup
		errChan := make(chan error, 1)
		sem := make(chan struct{}, 10) // Limit concurrent goroutines
		processed := 0

		for _, file := range files {
			// Skip non-MP3 files and directories
			if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".mp3") {
				continue
			}

			// Process file in a goroutine with semaphore for concurrency control
			wg.Add(1)
			go func(f os.FileInfo) {
				defer wg.Done()
				sem <- struct{}{} // Acquire semaphore
				defer func() { <-sem }() // Release semaphore

				// Process the file
				if err := processFile(f, trackManager); err != nil {
					// Non-fatal error, just log it
					// Warning processing file: %v
				}

				// Update progress
				processed++
				if processed%batchSize == 0 {
					// Processed %d files...
				}
			}(file)

			// Check for errors from goroutines
			select {
			case err := <-errChan:
				fmt.Fprintf(os.Stderr, "fatal error processing files: %v\n", err)
				os.Exit(1)
			default:
			}
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Final save
		// Saving tracks to database...
		if err := trackManager.SaveAll(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save tracks: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "rebuild completed. Processed %d files.\n", processed)
	},
}

// processFile processes a single file and adds it to the track manager
func processFile(file os.FileInfo, trackManager *tracks.Manager) error {
	// Extract video ID from filename
	videoID := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
	infoPath := filepath.Join(cfg.DownloadDir, videoID+".info.json")

	// Processing file: %s

	// Check if info file exists
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		return fmt.Errorf("info file not found for %s", videoID)
	}

	// Read track info from JSON file
	infoData, err := ioutil.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to read info file for %s: %w", videoID, err)
	}

	// Optimize the info.json file to remove unnecessary fields
	if err := yt.OptimizeInfoJSON(cfg, videoID); err != nil {
		// Warning: failed to optimize info.json for %s: %v
		// Continue processing even if optimization fails
	}

	// Re-read the optimized info.json
	infoData, err = ioutil.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to read optimized info file for %s: %w", videoID, err)
	}

	var trackInfo yt.TrackInfo
	if err := json.Unmarshal(infoData, &trackInfo); err != nil {
		return fmt.Errorf("failed to parse info file for %s: %w", videoID, err)
	}

	// Use Uploader as artist if available, otherwise use Creator or "Unknown Artist"
	artist := trackInfo.Uploader
	if artist == "" {
		artist = trackInfo.Creator
	}
	if artist == "" {
		artist = "Unknown Artist"
	}
	// Adding track: %s - %s

	// Add track to library
	if err := trackManager.AddTrack(trackInfo); err != nil {
		return fmt.Errorf("failed to add track %s: %w", videoID, err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(rebuildCmd)
}
