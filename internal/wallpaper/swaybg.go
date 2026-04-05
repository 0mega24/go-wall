package wallpaper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type swaybgSource struct{}

func (swaybgSource) Name() string { return "swaybg" }

// We check common config paths (e.g. sway/bg, background) for the wallpaper path.
var swaybgPaths = []string{
	"$HOME/.config/sway/bg",
	"$HOME/.config/sway/background",
	"$XDG_CONFIG_HOME/sway/bg",
}

func (swaybgSource) WallpaperPath() (string, error) {
	for _, p := range swaybgPaths {
		exp := os.ExpandEnv(p)
		b, err := os.ReadFile(exp)
		if err != nil {
			continue
		}
		path := strings.TrimSpace(string(b))
		if path == "" {
			continue
		}
		if path[0] == '/' {
			return path, nil
		}
		// might be relative to config dir
		dir := filepath.Dir(exp)
		if abs, err := filepath.Abs(filepath.Join(dir, path)); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs, nil
			}
		}
	}
	return "", fmt.Errorf("swaybg: no known path file found")
}

// SwayBG returns a Source that looks for swaybg wallpaper path in common config files.
func SwayBG() Source { return swaybgSource{} }
