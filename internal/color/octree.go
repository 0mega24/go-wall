package color

import (
	"fmt"
	"runtime"

	"github.com/0mega24/gowall/internal/imageutil"
)

// OctreeQuantizer implements octree color quantization.
// It is fully deterministic — seed is accepted by the interface but ignored.
type OctreeQuantizer struct{}

func (OctreeQuantizer) Name() string { return "octree" }

func (OctreeQuantizer) Cluster(pixels []imageutil.WeightedPixel, k, _, maxSamples int, _ int64, onStep func(int, string, []Centroid)) []Centroid {
	if len(pixels) == 0 || k <= 0 {
		return nil
	}
	if maxSamples <= 0 {
		maxSamples = DefaultMaxSamples
	}

	tree := newOctree()
	step := 0
	for i, p := range pixels {
		runtime.Gosched()
		if i >= maxSamples {
			break
		}
		c := FromColor(p.Color)
		tree.insert(c)
	}

	// Report levels as we understand the tree structure.
	for level := 1; level <= 8; level++ {
		if onStep != nil {
			label := fmt.Sprintf("level:%d", level)
			onStep(step, label, tree.palette())
			step++
		}
	}

	// Reduce until we have at most k colors.
	for tree.leafCount > k {
		runtime.Gosched()
		tree.reduce()
	}

	if onStep != nil {
		label := fmt.Sprintf("reduce:%d", k)
		onStep(step, label, tree.palette())
	}

	return tree.palette()
}

// octreeNode is a node in an RGB octree (up to 8 children per level, 8 levels deep).
type octreeNode struct {
	children         [8]*octreeNode
	sumR, sumG, sumB int
	count            int
	isLeaf           bool
	level            int
}

type octree struct {
	root      *octreeNode
	leafCount int
	levels    [8][]*octreeNode // reducible nodes per level
}

func newOctree() *octree {
	return &octree{root: &octreeNode{}}
}

func colorIndex(c Centroid, level int) int {
	shift := uint(7 - level)
	r := (c.R >> shift) & 1
	g := (c.G >> shift) & 1
	b := (c.B >> shift) & 1
	return int(r<<2 | g<<1 | b)
}

func (t *octree) insert(c Centroid) {
	node := t.root
	for level := 0; level < 8; level++ {
		idx := colorIndex(c, level)
		if node.children[idx] == nil {
			child := &octreeNode{level: level + 1}
			node.children[idx] = child
			if level < 7 {
				t.levels[level] = append(t.levels[level], child)
			}
		}
		node = node.children[idx]
	}
	// Leaf node
	if !node.isLeaf {
		node.isLeaf = true
		t.leafCount++
	}
	node.sumR += int(c.R)
	node.sumG += int(c.G)
	node.sumB += int(c.B)
	node.count++
}

// reduce merges the deepest reducible level's first node with its parent.
func (t *octree) reduce() {
	// Find deepest non-empty level.
	for level := 6; level >= 0; level-- {
		if len(t.levels[level]) > 0 {
			// Take the last node from this level.
			nodes := t.levels[level]
			node := nodes[len(nodes)-1]
			t.levels[level] = nodes[:len(nodes)-1]
			// Merge children into this node.
			for _, child := range node.children {
				if child == nil {
					continue
				}
				if child.isLeaf {
					t.leafCount--
				}
				node.sumR += child.sumR
				node.sumG += child.sumG
				node.sumB += child.sumB
				node.count += child.count
				child.sumR, child.sumG, child.sumB, child.count = 0, 0, 0, 0
				child.isLeaf = false
			}
			node.isLeaf = true
			t.leafCount++
			return
		}
	}
}

func (t *octree) palette() []Centroid {
	var result []Centroid
	collectLeaves(t.root, &result)
	return result
}

func collectLeaves(node *octreeNode, result *[]Centroid) {
	if node == nil {
		return
	}
	if node.isLeaf && node.count > 0 {
		*result = append(*result, Centroid{
			R:     uint8(node.sumR / node.count),
			G:     uint8(node.sumG / node.count),
			B:     uint8(node.sumB / node.count),
			Count: node.count,
		})
		return
	}
	for _, child := range node.children {
		collectLeaves(child, result)
	}
}
