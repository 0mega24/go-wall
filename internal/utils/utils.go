package utils

import (
	"image"
	"image/color"
)

func GetPixels(img image.Image) []color.Color {
	bounds := img.Bounds()
	pixels := make([]color.Color, 0, bounds.Dx()*bounds.Dy())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixels = append(pixels, img.At(x, y))
		}
	}

	return pixels
}
