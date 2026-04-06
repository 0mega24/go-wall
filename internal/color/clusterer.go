package color

import (
	"fmt"

	"github.com/0mega24/gowall/internal/imageutil"
)

// Clusterer extracts k representative colors from a set of weighted pixels.
// onStep is called after each meaningful algorithm step with a step index,
// a human-readable label, and the current centroid state. Pass nil for onStep
// in normal (non-visualization) use.
type Clusterer interface {
	Cluster(pixels []imageutil.WeightedPixel, k, iterations, maxSamples int, seed int64, onStep func(step int, label string, centroids []Centroid)) []Centroid
	// Name returns the algorithm identifier used in the registry.
	Name() string
}

var registry = map[string]Clusterer{}

func init() {
	register(&KMeansPP{})
	register(&KMeansLloyd{})
	register(&MedianCut{})
	register(&OctreeQuantizer{})
}

func register(c Clusterer) {
	registry[c.Name()] = c
}

// Get returns a Clusterer by name, or an error if the name is unknown.
// Registered names: "kmeans++", "kmeans", "mediancut", "octree".
func Get(name string) (Clusterer, error) {
	c, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("clustering: unknown algorithm %q (available: kmeans++, kmeans, mediancut, octree)", name)
	}
	return c, nil
}

// Available returns all registered algorithm names.
func Available() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
