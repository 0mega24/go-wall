package main

import (
	"fmt"
	"log"

	"github.com/0mega24/go-wall/internal/colors"
	"github.com/0mega24/go-wall/internal/utils"
	"github.com/0mega24/go-wall/internal/wallpaper"
)

func main() {
	path, err := wallpaper.CurrentWallpaperPath()
	if err != nil {
		log.Fatal("Error: Could not determine current wallpaper path:", err)
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

	pixels := utils.Colors(img)
	fmt.Printf("Loaded image with %d pixels\n", len(pixels))

	centroids := colors.KMeans(pixels, 32, 10)

	fmt.Println("\nDominant Colors (Hex):")
	for _, c := range centroids {
		fmt.Println(c.Hex())
	}
}
