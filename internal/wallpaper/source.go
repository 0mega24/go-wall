package wallpaper

import "fmt"

// Source returns the current wallpaper image path. Name identifies the source (e.g. "feh", "swaybg").
type Source interface {
	Name() string
	WallpaperPath() (string, error)
}

// FirstOf tries each source in order and returns the first path found.
func FirstOf(sources ...Source) (path, sourceName string, err error) {
	for _, s := range sources {
		path, err = s.WallpaperPath()
		if err == nil && path != "" {
			return path, s.Name(), nil
		}
	}
	return "", "", fmt.Errorf("no wallpaper path found (tried %d sources)", len(sources))
}
