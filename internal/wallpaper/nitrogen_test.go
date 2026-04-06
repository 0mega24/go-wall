package wallpaper_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0mega24/gowall/internal/wallpaper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNitrogen_ParseConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".config", "nitrogen")
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))

	cfg := "[xin_-1]\nfile=/home/user/wallpaper.jpg\nmode=4\nbgcolor=#000000\n"
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "bg-saved.cfg"), []byte(cfg), 0o644))

	src := wallpaper.Nitrogen()
	assert.Equal(t, "nitrogen", src.Name())

	path, err := src.WallpaperPath()
	require.NoError(t, err)
	assert.Equal(t, "/home/user/wallpaper.jpg", path)
}

func TestNitrogen_MissingConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := wallpaper.Nitrogen()
	_, err := src.WallpaperPath()
	require.Error(t, err)
}
