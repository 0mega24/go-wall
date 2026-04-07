package color

import (
	"fmt"
	"math/rand"
	"runtime"

	"github.com/0mega24/gowall/internal/imageutil"
)

// KMeansLloyd implements Lloyd's k-means algorithm with random (uniform) initialization.
// Unlike k-means++, the seed here selects the initial centroids uniformly at random.
type KMeansLloyd struct{}

func (KMeansLloyd) Name() string { return "kmeans" }

func (KMeansLloyd) Cluster(pixels []imageutil.WeightedPixel, k, iterations, maxSamples int, seed int64, onStep func(int, string, []Centroid)) []Centroid {
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

	// Random init: pick k unique points uniformly at random.
	centroids := initCentroidsRandom(points, k, rng)
	if onStep != nil {
		onStep(0, "init:random", cloneCentroids(centroids))
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

// initCentroidsRandom picks k initial centroids uniformly at random (no weighting).
func initCentroidsRandom(points []weightedPoint, k int, rng *rand.Rand) []Centroid {
	indices := rng.Perm(len(points))
	centroids := make([]Centroid, k)
	for i := 0; i < k; i++ {
		centroids[i] = points[indices[i]].c
	}
	return centroids
}
