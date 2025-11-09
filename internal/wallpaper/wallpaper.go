package wallpaper

import (
	"bufio"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"regexp"
)

func CurrentWallpaperPath() (string, error) {
	// feh stores the current wallpaper path in $HOME/.fehbg
	file, err := os.Open(os.ExpandEnv("$HOME/.fehbg"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Regex to capture the path string inside single quotes in the feh command line
	re := regexp.MustCompile(`'(.*)'`)

	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		// match[0] is the full match, match[1] is the captured path (inside the quotes)
		if len(match) == 2 {
			return match[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("wallpaper path not found in .fehbg")
}

func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return img, nil
}
