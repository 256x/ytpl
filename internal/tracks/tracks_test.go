package tracks

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ytpl/internal/yt"
)

func TestTracks(t *testing.T) {
	tmpDir := t.TempDir()
	tracksPath := filepath.Join(tmpDir, ".tracks")

	t.Run("new and load empty", func(t *testing.T) {
		tracks := New(tracksPath)
		err := tracks.Load()
		require.NoError(t, err, "failed to load tracks")
		assert.Empty(t, tracks.List(), "expected empty track list")
	})

	t.Run("add and list tracks", func(t *testing.T) {
		t.Log("creating new Tracks instance")
		tracks := New(tracksPath)
		
		t.Log("creating test tracks")
		track1 := yt.TrackInfo{ID: "1", Title: "b track"}
		track2 := yt.TrackInfo{ID: "2", Title: "a track"}
		
		t.Log("adding track1")
		require.NoError(t, tracks.Add(track1), "failed to add track1")
		t.Log("adding track2")
		require.NoError(t, tracks.Add(track2), "failed to add track2")
		
		t.Log("listing tracks")
		list := tracks.List()
		t.Logf("got %d tracks: %v", len(list), list)
		require.Len(t, list, 2, "expected 2 tracks")
		assert.Equal(t, "a track", list[0].Title, "tracks should be sorted by title")
		assert.Equal(t, "b track", list[1].Title, "tracks should be sorted by title")
	})

	t.Run("update existing track", func(t *testing.T) {
		tracks := New(tracksPath)
		
		track1 := yt.TrackInfo{ID: "1", Title: "original title"}
		require.NoError(t, tracks.Add(track1), "failed to add initial track")
		
		updatedTrack := yt.TrackInfo{ID: "1", Title: "updated title"}
		require.NoError(t, tracks.Add(updatedTrack), "failed to update track")
		
		list := tracks.List()
		require.Len(t, list, 1, "should only have one track after update")
		assert.Equal(t, "updated title", list[0].Title, "track title should be updated")
	})

	t.Run("remove track", func(t *testing.T) {
		tracks := New(tracksPath)
		
		track := yt.TrackInfo{ID: "1", Title: "test track"}
		require.NoError(t, tracks.Add(track), "failed to add track")
		require.NoError(t, tracks.Remove("1"), "failed to remove track")
		
		assert.Empty(t, tracks.List(), "track list should be empty after removal")
	})

	t.Run("persistence", func(t *testing.T) {
		// first instance - add tracks
		tracks1 := New(tracksPath)
		require.NoError(t, tracks1.Add(yt.TrackInfo{ID: "1", Title: "test track"}), "failed to add track")
		
		// second instance - should load the same tracks
		tracks2 := New(tracksPath)
		require.NoError(t, tracks2.Load(), "failed to load tracks")
		
		list := tracks2.List()
		require.Len(t, list, 1, "expected one track after loading")
		assert.Equal(t, "test track", list[0].Title, "track title should match")
	})
}
