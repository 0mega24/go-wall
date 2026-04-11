package wallpaper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadImage(t *testing.T) {
	// Use testdata if present
	path := filepath.Join("..", "..", "testdata", "sample-wallpaper.png")
	if _, err := os.Stat(path); err != nil {
		t.Skip("testdata/sample-wallpaper.png not found")
	}
	img, err := LoadImage(path)
	require.NoError(t, err)
	assert.True(t, img.Bounds().Dx() > 0 && img.Bounds().Dy() > 0, "image bounds empty")
}

func TestEnvSource(t *testing.T) {
	s := EnvSource()
	assert.Equal(t, "env", s.Name())
	_, err := s.WallpaperPath()
	if err == nil {
		// GOWALL_IMAGE or WALLPAPER might be set in env
		return
	}
	// expect error when unset
}

func TestFirstOf_Empty(t *testing.T) {
	_, _, err := FirstOf()
	assert.Error(t, err, "FirstOf() with no sources should error")
}

func TestFeh_Name(t *testing.T) {
	assert.Equal(t, "feh", Feh().Name())
}

func TestHyprland_Name(t *testing.T) {
	assert.Equal(t, "hyprland", Hyprland().Name())
}

func TestSwayBG_Name(t *testing.T) {
	assert.Equal(t, "swaybg", SwayBG().Name())
}
