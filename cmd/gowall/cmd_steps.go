package main

import (
	"fmt"
	"path/filepath"
	"strings"

	icolor "github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/imageutil"
	"github.com/0mega24/gowall/internal/wallpaper"
	"github.com/spf13/cobra"
)

var stepsCmd = &cobra.Command{
	Use:   "steps [image]",
	Short: "Export a quantized image at each clustering algorithm step",
	Long: `Runs the chosen clustering algorithm and saves a quantized version of the
image at every meaningful step — showing visually how the algorithm converges.

Output files are named: <step>-<label>.png  (e.g. 01-iter-1.png)
A final palette.png swatch is also saved.`,
	RunE: runSteps,
}

var (
	stepsAlgorithm string
	stepsOut       string
	stepsSeed      int64
	stepsK         int
	stepsIters     int
)

func init() {
	rootCmd.AddCommand(stepsCmd)
	stepsCmd.Flags().StringVar(&stepsAlgorithm, "algorithm", "kmeans++", "clustering algorithm (kmeans++, kmeans, mediancut, octree)")
	stepsCmd.Flags().StringVar(&stepsOut, "out", "./steps", "output directory for step images")
	stepsCmd.Flags().Int64Var(&stepsSeed, "seed", 0, "RNG seed (kmeans variants only)")
	stepsCmd.Flags().IntVar(&stepsK, "k", 32, "number of target colors")
	stepsCmd.Flags().IntVar(&stepsIters, "iters", 10, "iterations (kmeans variants only)")
}

func runSteps(cmd *cobra.Command, args []string) error {
	clusterer, err := icolor.Get(stepsAlgorithm)
	if err != nil {
		return err
	}

	var imagePath string
	if len(args) > 0 {
		imagePath = args[0]
	} else {
		p, _, e := wallpaper.FirstOf(wallpaper.DefaultSources()...)
		if e != nil {
			return fmt.Errorf("wallpaper: %w", e)
		}
		imagePath = p
	}

	img, err := wallpaper.LoadImage(imagePath)
	if err != nil {
		return err
	}

	pixels := imageutil.Pixels(img)

	var stepImages []struct {
		step  int
		label string
		cs    []icolor.Centroid
	}

	clusterer.Cluster(pixels, stepsK, stepsIters, 50000, stepsSeed, func(step int, label string, cs []icolor.Centroid) {
		stepImages = append(stepImages, struct {
			step  int
			label string
			cs    []icolor.Centroid
		}{step, label, cs})
	})

	// Save one quantized image per step.
	total := len(stepImages)
	digits := len(fmt.Sprintf("%d", total))
	format := fmt.Sprintf("%%0%dd-%%s.png", digits)

	for _, si := range stepImages {
		safeName := strings.ReplaceAll(si.label, ":", "-")
		fname := fmt.Sprintf(format, si.step, safeName)
		path := filepath.Join(stepsOut, fname)
		quantized := icolor.QuantizeImage(img, si.cs)
		if err := icolor.SavePNG(quantized, path); err != nil {
			return fmt.Errorf("save %s: %w", path, err)
		}
		fmt.Println("Wrote", path)
	}

	// Save final palette swatch.
	if len(stepImages) > 0 {
		final := stepImages[len(stepImages)-1].cs
		palettePath := filepath.Join(stepsOut, "palette.png")
		if err := icolor.SaveSwatchPNG(final, 32, 8, palettePath); err != nil {
			return fmt.Errorf("palette: %w", err)
		}
		fmt.Println("Wrote", palettePath)
	}

	return nil
}
