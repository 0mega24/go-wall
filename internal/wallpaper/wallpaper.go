package wallpaper

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func GetCurrentWallpaper() (string, error) {
	file, err := os.Open(os.ExpandEnv("$HOME/.fehbg"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`'(.*)'`)

	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if len(match) == 2 {
			return match[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("wallpaper path not found")
}
