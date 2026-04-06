package color

import (
	"image"
	"image/color"
)

// SwatchPNG renders a palette as a tiled image where each color occupies a
// tileSize×tileSize square. Tiles are laid out left-to-right in rows of perRow.
func SwatchPNG(colors []Centroid, tileSize, perRow int) image.Image {
	if len(colors) == 0 || tileSize <= 0 || perRow <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}
	cols := perRow
	rows := (len(colors) + cols - 1) / cols
	w := cols * tileSize
	h := rows * tileSize
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i, c := range colors {
		col := i % cols
		row := i / cols
		x0 := col * tileSize
		y0 := row * tileSize
		fill := color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}
		for y := y0; y < y0+tileSize; y++ {
			for x := x0; x < x0+tileSize; x++ {
				img.SetNRGBA(x, y, fill)
			}
		}
	}
	return img
}

// SaveSwatchPNG renders and saves a palette swatch PNG to path.
func SaveSwatchPNG(colors []Centroid, tileSize, perRow int, path string) error {
	img := SwatchPNG(colors, tileSize, perRow)
	return SavePNG(img, path)
}
