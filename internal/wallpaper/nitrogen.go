package wallpaper

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type nitrogenSource struct{}

// Nitrogen returns a Source that reads the wallpaper path from Nitrogen's config.
func Nitrogen() Source {
	return nitrogenSource{}
}

func (nitrogenSource) Name() string { return "nitrogen" }

func (nitrogenSource) WallpaperPath() (string, error) {
	home := os.Getenv("HOME")
	cfgPath := filepath.Join(home, ".config", "nitrogen", "bg-saved.cfg")
	f, err := os.Open(filepath.Clean(cfgPath))
	if err != nil {
		return "", fmt.Errorf("nitrogen: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	inSection := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			inSection = strings.HasPrefix(line, "[xin_")
			continue
		}
		if inSection && strings.HasPrefix(line, "file=") {
			return strings.TrimPrefix(line, "file="), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("nitrogen: scanning config: %w", err)
	}
	return "", fmt.Errorf("nitrogen: no wallpaper path found in %s", cfgPath)
}
