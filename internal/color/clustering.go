package color

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/0mega24/gowall/internal/imageutil"
)

// newRNG returns a seeded rand.Rand. If seed == 0, a random seed is used.
func newRNG(seed int64) *rand.Rand {
	if seed != 0 {
		return rand.New(rand.NewSource(seed))
	}
	return rand.New(rand.NewSource(rand.Int63()))
}

const (
	// DefaultMaxSamples is the maximum number of pixels used for clustering when
	// the image has more pixels. Random sampling preserves the color distribution
	// better than stride-based sampling.
	DefaultMaxSamples = 50000
)

// Centroid represents a point in RGB color space (0–255).
type Centroid struct {
	R, G, B uint8
	Count   int
}

// FromColor converts color.Color to Centroid (RGBA 16-bit → 8-bit).
func FromColor(c color.Color) Centroid {
	r, g, b, _ := c.RGBA()
	return Centroid{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
	}
}

// DistanceSq returns the squared Euclidean distance between two centroids.
func DistanceSq(a, b Centroid) float32 {
	dr := float32(a.R) - float32(b.R)
	dg := float32(a.G) - float32(b.G)
	db := float32(a.B) - float32(b.B)
	return dr*dr + dg*dg + db*db
}

// weightedPoint holds a pre-decoded Centroid and its sample weight.
type weightedPoint struct {
	c      Centroid
	weight float32
}

// pixelsToWeightedPoints converts WeightedPixels to weightedPoints for internal use.
func pixelsToWeightedPoints(pixels []imageutil.WeightedPixel) []weightedPoint {
	out := make([]weightedPoint, len(pixels))
	for i, p := range pixels {
		out[i] = weightedPoint{
			c:      FromColor(p.Color),
			weight: p.Weight,
		}
	}
	return out
}

// samplePoints returns up to max points by random sampling when len(points) > max.
// When points fit in max, returns a copy so the caller can mutate safely.
func samplePoints(points []weightedPoint, max int, rng *rand.Rand) []weightedPoint {
	if len(points) <= max {
		out := make([]weightedPoint, len(points))
		copy(out, points)
		return out
	}
	indices := rng.Perm(len(points))[:max]
	out := make([]weightedPoint, max)
	for i, j := range indices {
		out[i] = points[j]
	}
	return out
}

// initCentroidsKMeansPP chooses k initial centroids using k-means++:
// first centroid uniformly at random, then each next with probability ∝ D(x)².
func initCentroidsKMeansPP(points []weightedPoint, k int, rng *rand.Rand) []Centroid {
	if len(points) == 0 || k <= 0 {
		return nil
	}
	if k > len(points) {
		k = len(points)
	}
	centroids := make([]Centroid, 0, k)

	// First centroid: uniform random
	centroids = append(centroids, points[rng.Intn(len(points))].c)

	for len(centroids) < k {
		// D(x)² = squared distance from each point to its nearest chosen centroid
		dSq := make([]float64, len(points))
		var sum float64
		for i, p := range points {
			minD := float64(DistanceSq(p.c, centroids[0]))
			for _, c := range centroids[1:] {
				if d := float64(DistanceSq(p.c, c)); d < minD {
					minD = d
				}
			}
			dSq[i] = minD
			sum += minD
		}
		if sum == 0 {
			// All points coincide with centroids; pick any remaining point
			centroids = append(centroids, points[rng.Intn(len(points))].c)
			continue
		}
		// Weighted random choice: pick next centroid with probability ∝ D(x)²
		u := rng.Float64() * sum
		var cum float64
		for i, d := range dSq {
			cum += d
			if cum >= u {
				centroids = append(centroids, points[i].c)
				break
			}
		}
	}
	return centroids
}

// assignPoints assigns each point to the nearest centroid (assignment step).
// Weights do not affect assignment — only centroid position.
func assignPoints(points []weightedPoint, centroids []Centroid) []int {
	assignments := make([]int, len(points))
	for i, p := range points {
		bestIdx := 0
		bestDist := DistanceSq(p.c, centroids[0])
		for j, c := range centroids[1:] {
			if d := DistanceSq(p.c, c); d < bestDist {
				bestDist = d
				bestIdx = j + 1
			}
		}
		assignments[i] = bestIdx
	}
	return assignments
}

// updateCentroids sets each centroid to the weighted mean of its assigned points.
// centroid.R = Σ(R * weight) / Σ(weight), likewise for G and B.
func updateCentroids(points []weightedPoint, assignments []int, centroids []Centroid) {
	k := len(centroids)
	sumR := make([]float64, k)
	sumG := make([]float64, k)
	sumB := make([]float64, k)
	sumW := make([]float64, k)
	counts := make([]int, k)
	for i, p := range points {
		idx := assignments[i]
		w := float64(p.weight)
		sumR[idx] += float64(p.c.R) * w
		sumG[idx] += float64(p.c.G) * w
		sumB[idx] += float64(p.c.B) * w
		sumW[idx] += w
		counts[idx]++
	}
	for i := range centroids {
		if sumW[i] > 0 {
			centroids[i].R = uint8(sumR[i] / sumW[i])
			centroids[i].G = uint8(sumG[i] / sumW[i])
			centroids[i].B = uint8(sumB[i] / sumW[i])
		}
		centroids[i].Count = counts[i]
	}
}

// KMeans runs k-means with k-means++ initialization on the image colors.
func KMeans(pixels []imageutil.WeightedPixel, k, iterations int) []Centroid {
	return KMeansWithOptions(pixels, k, iterations, DefaultMaxSamples, 0)
}

// KMeansWithOptions is like KMeans but allows setting max samples and RNG seed.
func KMeansWithOptions(pixels []imageutil.WeightedPixel, k, iterations, maxSamples int, seed int64) []Centroid {
	if len(pixels) == 0 || k <= 0 {
		return nil
	}
	if maxSamples <= 0 {
		maxSamples = DefaultMaxSamples
	}
	var rng *rand.Rand
	if seed != 0 {
		rng = rand.New(rand.NewSource(seed))
	} else {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	all := pixelsToWeightedPoints(pixels)
	points := samplePoints(all, maxSamples, rng)
	if len(points) < k {
		k = len(points)
	}
	if k == 0 {
		return nil
	}

	centroids := initCentroidsKMeansPP(points, k, rng)
	for i := 0; i < iterations; i++ {
		assignments := assignPoints(points, centroids)
		updateCentroids(points, assignments, centroids)
	}
	return centroids
}

// Hex returns a web hex color string (#rrggbb).
func (c Centroid) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// RawHex returns hex without the # prefix.
func (c Centroid) RawHex() string {
	return fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B)
}
