package color

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// QuantizeImage maps each pixel of img to its nearest Centroid by squared Euclidean
// distance and returns the quantized result as an *image.NRGBA.
func QuantizeImage(img image.Image, centroids []Centroid) image.Image {
	bounds := img.Bounds()
	out := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px := FromColor(img.At(x, y))
			nearest := nearestCentroid(px, centroids)
			out.SetNRGBA(x, y, color.NRGBA{R: nearest.R, G: nearest.G, B: nearest.B, A: 255})
		}
	}
	return out
}

// SavePNG writes any image.Image to a PNG file at path.
// Parent directories are created as needed.
func SavePNG(img image.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

func nearestCentroid(px Centroid, centroids []Centroid) Centroid {
	if len(centroids) == 0 {
		return px
	}
	best := centroids[0]
	bestDist := DistanceSq(px, best)
	for _, c := range centroids[1:] {
		if d := DistanceSq(px, c); d < bestDist {
			bestDist = d
			best = c
		}
	}
	return best
}
