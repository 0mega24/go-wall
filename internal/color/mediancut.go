package color

import (
	"fmt"
	"runtime"
	"sort"

	"github.com/0mega24/gowall/internal/imageutil"
)

// MedianCut implements median cut color quantization.
// It is fully deterministic — seed is accepted by the interface but ignored.
type MedianCut struct{}

// Name returns the algorithm identifier "mediancut".
func (MedianCut) Name() string { return "mediancut" }

// Cluster runs median-cut quantization on pixels and returns k centroids.
func (MedianCut) Cluster(pixels []imageutil.WeightedPixel, k, _, maxSamples int, _ int64, onStep func(int, string, []Centroid)) []Centroid {
	if len(pixels) == 0 || k <= 0 {
		return nil
	}
	if maxSamples <= 0 {
		maxSamples = DefaultMaxSamples
	}

	// Decode pixels to Centroid slice.
	pts := make([]Centroid, 0, len(pixels))
	for _, p := range pixels {
		pts = append(pts, FromColor(p.Color))
	}
	if len(pts) > maxSamples {
		pts = pts[:maxSamples]
	}

	buckets := [][]Centroid{pts}
	step := 0

	for len(buckets) < k && len(buckets) > 0 {
		runtime.Gosched()
		// Find the bucket with the widest channel range.
		idx := widestBucket(buckets)
		bucket := buckets[idx]

		// Sort along the widest axis, split at median.
		axis := widestAxis(bucket)
		sortBucket(bucket, axis)
		mid := len(bucket) / 2
		left := bucket[:mid]
		right := bucket[mid:]

		buckets = append(buckets[:idx], buckets[idx+1:]...)
		buckets = append(buckets, left, right)

		step++
		if onStep != nil {
			label := fmt.Sprintf("split:%d", step)
			onStep(step, label, bucketsToAverages(buckets))
		}
	}

	return bucketsToAverages(buckets)
}

type axis int

const (
	axisR axis = iota
	axisG
	axisB
)

func widestAxis(bucket []Centroid) axis {
	if len(bucket) == 0 {
		return axisR
	}
	minR, maxR := bucket[0].R, bucket[0].R
	minG, maxG := bucket[0].G, bucket[0].G
	minB, maxB := bucket[0].B, bucket[0].B
	for _, c := range bucket[1:] {
		if c.R < minR {
			minR = c.R
		}
		if c.R > maxR {
			maxR = c.R
		}
		if c.G < minG {
			minG = c.G
		}
		if c.G > maxG {
			maxG = c.G
		}
		if c.B < minB {
			minB = c.B
		}
		if c.B > maxB {
			maxB = c.B
		}
	}
	rangeR := int(maxR) - int(minR)
	rangeG := int(maxG) - int(minG)
	rangeB := int(maxB) - int(minB)
	if rangeR >= rangeG && rangeR >= rangeB {
		return axisR
	}
	if rangeG >= rangeB {
		return axisG
	}
	return axisB
}

// widestBucket returns the index of the bucket with the widest single-channel range.
func widestBucket(buckets [][]Centroid) int {
	best := 0
	bestRange := -1
	for i, b := range buckets {
		ax := widestAxis(b)
		r := axisRange(b, ax)
		if r > bestRange {
			bestRange = r
			best = i
		}
	}
	return best
}

func axisRange(bucket []Centroid, ax axis) int {
	if len(bucket) == 0 {
		return 0
	}
	var lo, hi uint8
	switch ax {
	case axisR:
		lo, hi = bucket[0].R, bucket[0].R
		for _, c := range bucket[1:] {
			if c.R < lo {
				lo = c.R
			}
			if c.R > hi {
				hi = c.R
			}
		}
	case axisG:
		lo, hi = bucket[0].G, bucket[0].G
		for _, c := range bucket[1:] {
			if c.G < lo {
				lo = c.G
			}
			if c.G > hi {
				hi = c.G
			}
		}
	case axisB:
		lo, hi = bucket[0].B, bucket[0].B
		for _, c := range bucket[1:] {
			if c.B < lo {
				lo = c.B
			}
			if c.B > hi {
				hi = c.B
			}
		}
	}
	return int(hi) - int(lo)
}

func sortBucket(bucket []Centroid, ax axis) {
	sort.Slice(bucket, func(i, j int) bool {
		switch ax {
		case axisR:
			return bucket[i].R < bucket[j].R
		case axisG:
			return bucket[i].G < bucket[j].G
		case axisB:
			return bucket[i].B < bucket[j].B
		default:
			return false
		}
	})
}

func bucketsToAverages(buckets [][]Centroid) []Centroid {
	out := make([]Centroid, 0, len(buckets))
	for _, b := range buckets {
		if len(b) == 0 {
			continue
		}
		var sumR, sumG, sumB int
		for _, c := range b {
			sumR += int(c.R)
			sumG += int(c.G)
			sumB += int(c.B)
		}
		n := len(b)
		out = append(out, Centroid{
			R:     uint8(sumR / n),
			G:     uint8(sumG / n),
			B:     uint8(sumB / n),
			Count: n,
		})
	}
	return out
}
