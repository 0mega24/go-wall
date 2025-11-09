package main

import (
	"fmt"
	"log"

	"github.com/0mega24/go-wall/internal/colors"
	"github.com/0mega24/go-wall/internal/colors/palette"
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

	palette.PrintHex(centroids)

	filtered := palette.FilterSimilar(centroids, 1000)
	palette.SortByBrightness(filtered)

	ansiColors := palette.GenerateANSI(filtered, .6)
	tones := palette.GenerateTones(filtered, 8)

	fmt.Println("\nFiltered")
	palette.PrintHex(filtered)

	fmt.Println("\nANSI Colors:")
	palette.PrintHex(ansiColors)

	fmt.Println("\nTones Colors:")
	palette.PrintHex(tones)
}
