package imageutil

import (
	"image"
	"image/color"
)

// WeightedPixel is a color sample with a weight in [0.0, 1.0].
// Weight reflects alpha: fully opaque = 1.0, fully transparent = 0.0.
type WeightedPixel struct {
	Color  color.Color
	Weight float32
}

// Pixels returns all non-transparent pixels from the image as WeightedPixels.
// Fully transparent pixels (alpha == 0) are skipped. Semi-transparent pixels
// are included with weight = alpha / 65535. Fully opaque pixels have weight 1.0.
func Pixels(img image.Image) []WeightedPixel {
	bounds := img.Bounds()
	var out []WeightedPixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			_, _, _, a := c.RGBA() // returns [0, 65535]
			if a == 0 {
				continue
			}
			out = append(out, WeightedPixel{
				Color:  c,
				Weight: float32(a) / 65535,
			})
		}
	}
	return out
}

// PixelsUnweighted returns all pixel colors from the image in row-major order.
// Kept for internal use where a flat []color.Color slice is needed.
func PixelsUnweighted(img image.Image) []color.Color {
	bounds := img.Bounds()
	pixels := make([]color.Color, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixels = append(pixels, img.At(x, y))
		}
	}
	return pixels
}
