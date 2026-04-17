package wallpaper

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnvSource reads the wallpaper path from environment variable GOWALL_IMAGE
// (or WALLPAPER if GOWALL_IMAGE is not set). Useful for scripts or WMs that
// don't store the path in a standard file.
type envSource struct{}

func (envSource) Name() string { return "env" }

func (envSource) WallpaperPath() (string, error) {
	path := os.Getenv("GOWALL_IMAGE")
	if path == "" {
		path = os.Getenv("WALLPAPER")
	}
	if path == "" {
		return "", fmt.Errorf("GOWALL_IMAGE and WALLPAPER not set")
	}
	path = os.ExpandEnv(path)
	if _, err := os.Stat(filepath.Clean(path)); err != nil {
		return "", err
	}
	return path, nil
}

// EnvSource returns a Source that uses GOWALL_IMAGE or WALLPAPER env vars.
func EnvSource() Source { return envSource{} }
