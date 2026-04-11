package wallpaper

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// We parse hyprland.conf (and common paths) for wallpaper = path.
type hyprlandSource struct{}

func (hyprlandSource) Name() string { return "hyprland" }

var hyprlandConfigPaths = []string{
	"$XDG_CONFIG_HOME/hypr/hyprland.conf",
	"$HOME/.config/hypr/hyprland.conf",
}

// wallpaper = path or monitor,path
var hyprWallpaperRe = regexp.MustCompile(`wallpaper\s*=\s*([^,\s]+)`)

func (hyprlandSource) WallpaperPath() (string, error) {
	for _, p := range hyprlandConfigPaths {
		exp := os.ExpandEnv(p)
		f, err := os.Open(exp)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if strings.HasPrefix(line, "#") {
				continue
			}
			m := hyprWallpaperRe.FindStringSubmatch(line)
			if len(m) != 2 {
				continue
			}
			path := m[1]
			if path[0] != '/' {
				path = filepath.Join(filepath.Dir(exp), path)
			}
			if _, err := os.Stat(path); err == nil {
				_ = f.Close()
				return path, nil
			}
		}
		_ = f.Close()
	}
	return "", fmt.Errorf("hyprland: no wallpaper= in config")
}

// Hyprland returns a Source that parses hyprland.conf for wallpaper=.
func Hyprland() Source { return hyprlandSource{} }
