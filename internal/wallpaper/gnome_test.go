package wallpaper_test

import (
	"testing"

	"github.com/0mega24/gowall/internal/wallpaper"
	"github.com/stretchr/testify/require"
)

func TestGNOME_Name(t *testing.T) {
	src := wallpaper.GNOME()
	require.Equal(t, "gnome", src.Name())
}

func TestGNOME_NoBinary(t *testing.T) {
	// When gsettings is not in PATH, WallpaperPath should return an error (not panic)
	t.Setenv("PATH", "")
	src := wallpaper.GNOME()
	_, err := src.WallpaperPath()
	require.Error(t, err, "GNOME source should return error when gsettings not found")
}
