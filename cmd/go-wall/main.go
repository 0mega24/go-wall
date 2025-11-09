package main

import (
	"fmt"
	"log"

	"github.com/0mega24/go-wall/internal/utils"
	"github.com/0mega24/go-wall/internal/wallpaper"
)

func main() {
	path, err := wallpaper.GetCurrentWallpaper()
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Println("Current wallpaper:", path)

	img, err := wallpaper.LoadImage(path)
	if err != nil {
		log.Fatal("Error loading image:", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	fmt.Printf("Resolution: %dx%d\n", width, height)

	pixels := utils.GetPixels(img)
	fmt.Printf("Loaded image with %d pixels\n", len(pixels))
}
