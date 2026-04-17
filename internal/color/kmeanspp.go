package color

import (
	"fmt"
	"runtime"

	"github.com/0mega24/gowall/internal/imageutil"
)

// KMeansPP implements k-means++ clustering (seeded probabilistic init).
// The seed controls the random initial centroid placement for reproducibility.
type KMeansPP struct{}

// Name returns the algorithm identifier "kmeans++".
func (KMeansPP) Name() string { return "kmeans++" }

// Cluster runs k-means++ on pixels and returns k centroids.
func (KMeansPP) Cluster(pixels []imageutil.WeightedPixel, k, iterations, maxSamples int, seed int64, onStep func(int, string, []Centroid)) []Centroid {
	if len(pixels) == 0 || k <= 0 {
		return nil
	}
	if maxSamples <= 0 {
		maxSamples = DefaultMaxSamples
	}
	rng := newRNG(seed)
	all := pixelsToWeightedPoints(pixels)
	points := samplePoints(all, maxSamples, rng)
	if len(points) < k {
		k = len(points)
	}
	if k == 0 {
		return nil
	}

	centroids := initCentroidsKMeansPP(points, k, rng)
	if onStep != nil {
		onStep(0, "init:seed", cloneCentroids(centroids))
	}

	for i := 0; i < iterations; i++ {
		runtime.Gosched()
		assignments := assignPoints(points, centroids)
		updateCentroids(points, assignments, centroids)
		if onStep != nil {
			label := fmt.Sprintf("iter:%d", i+1)
			onStep(i+1, label, cloneCentroids(centroids))
		}
	}
	if onStep != nil {
		onStep(iterations+1, "converged", cloneCentroids(centroids))
	}
	return centroids
}

func cloneCentroids(cs []Centroid) []Centroid {
	out := make([]Centroid, len(cs))
	copy(out, cs)
	return out
}
