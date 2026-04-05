package wallpaper

import (
	"fmt"
	"os/exec"
	"strings"
)

type gnomeSource struct{}

// GNOME returns a Source that reads the wallpaper path from gsettings.
func GNOME() Source {
	return gnomeSource{}
}

func (gnomeSource) Name() string { return "gnome" }

func (gnomeSource) WallpaperPath() (string, error) {
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.background", "picture-uri").Output()
	if err != nil {
		return "", fmt.Errorf("gnome: gsettings failed: %w", err)
	}
	// Output looks like: 'file:///home/user/wallpaper.jpg'\n
	s := strings.TrimSpace(string(out))
	s = strings.Trim(s, "'")
	s = strings.TrimPrefix(s, "file://")
	if s == "" || s == "none" {
		return "", fmt.Errorf("gnome: no wallpaper configured")
	}
	return s, nil
}
