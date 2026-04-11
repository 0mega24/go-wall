package wallpaper

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

type fehSource struct{}

func (fehSource) Name() string { return "feh" }

func (fehSource) WallpaperPath() (string, error) {
	path := os.ExpandEnv("$HOME/.fehbg")
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	re := regexp.MustCompile(`'(.*)'`)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		match := re.FindStringSubmatch(sc.Text())
		if len(match) == 2 {
			return match[1], nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no path in %s", path)
}

// Feh returns a Source that reads the wallpaper path from feh's ~/.fehbg.
func Feh() Source { return fehSource{} }
