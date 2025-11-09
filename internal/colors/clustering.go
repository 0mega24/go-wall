package colors

import (
	"fmt"
	"image/color"
)

// Centroid represents a single point in the 3D RGB color space.
// The components are stored as 8-bit unsigned integers (0–255).
type Centroid struct {
	R, G, B uint8
}

// FromColor converts Go's standard color.Color interface into a Centroid.
// color.Color components are 16-bit (0–65535);
// the conversion to 8-bit (0–255) is performed via right-shifting (>> 8).
func FromColor(c color.Color) Centroid {
	r, g, b, _ := c.RGBA()
	return Centroid{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
	}
}

// distanceSq calculates the squared Euclidean distance between two Centroids (points).
func distanceSq(a, b Centroid) float32 {
	dr := float32(a.R) - float32(b.R)
	dg := float32(a.G) - float32(b.G)
	db := float32(a.B) - float32(b.B)
	return dr*dr + dg*dg + db*db
}

// initCentroids selects the initial positions for the K cluster centers.
func initCentroids(sampledColors []Centroid, k int) []Centroid {
	centroids := make([]Centroid, k)
	step := len(sampledColors) / k

	for i := 0; i < k; i++ {
		centroids[i] = sampledColors[i*step]
	}
	return centroids
}

// assignPoints implements the 'Assignment Step' of the K-Means algorithm.
func assignPoints(points []Centroid, centroids []Centroid) []int {
	assignments := make([]int, len(points))

	for i, p := range points {
		bestIdx := 0
		// Initialize with the distance to the first centroid
		bestDist := distanceSq(p, centroids[0])

		for j, c := range centroids {
			if d := distanceSq(p, c); d < bestDist {
				bestDist = d
				bestIdx = j
			}
		}
		assignments[i] = bestIdx
	}
	return assignments
}

// updateCentroids implements the 'Update Step' of the K-Means algorithm.
func updateCentroids(points []Centroid, assignments []int, centroids []Centroid) {
	numCentroids := len(centroids)
	counts := make([]int, numCentroids)
	sumR := make([]int, numCentroids)
	sumG := make([]int, numCentroids)
	sumB := make([]int, numCentroids)

	// Aggregate R, G, B sums and counts for each cluster
	for i, p := range points {
		idx := assignments[i]
		sumR[idx] += int(p.R)
		sumG[idx] += int(p.G)
		sumB[idx] += int(p.B)
		counts[idx]++
	}

	for i := range centroids {
		if counts[i] > 0 {
			centroids[i].R = uint8(sumR[i] / counts[i])
			centroids[i].G = uint8(sumG[i] / counts[i])
			centroids[i].B = uint8(sumB[i] / counts[i])
		}
	}
}

// KMeans performs K-Means clustering on pixel colors to find k dominant colors.
func KMeans(pixels []color.Color, k int, iterations int) []Centroid {
	sampledColors := make([]Centroid, 0, len(pixels)/10)
	for i := 0; i < len(pixels); i += 10 {
		sampledColors = append(sampledColors, FromColor(pixels[i]))
	}

	if len(sampledColors) < k {
		k = len(sampledColors)
	}
	if k == 0 {
		return nil
	}

	centroids := initCentroids(sampledColors, k)

	for i := 0; i < iterations; i++ {
		assignments := assignPoints(sampledColors, centroids)
		updateCentroids(sampledColors, assignments, centroids)
	}

	return centroids
}

// Hex returns a standard web hex color string (#rrggbb) for the Centroid.
func (c Centroid) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}
