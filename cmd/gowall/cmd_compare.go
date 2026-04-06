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

var compareCmd = &cobra.Command{
	Use:   "compare [image]",
	Short: "Compare clustering algorithms side-by-side",
	Long: `Runs multiple clustering algorithms on the image and saves a quantized PNG
and palette swatch for each. Also prints a side-by-side table of the results.

Output files:
  <out>/<algorithm>-final.png    — quantized image with final centroids
  <out>/<algorithm>-palette.png  — palette swatch`,
	RunE: runCompare,
}

var (
	compareAlgorithms string
	compareOut        string
	compareSeed       int64
	compareK          int
	compareIters      int
)

func init() {
	rootCmd.AddCommand(compareCmd)
	compareCmd.Flags().StringVar(&compareAlgorithms, "algorithms", "kmeans++,kmeans,mediancut,octree", "comma-separated list of algorithms")
	compareCmd.Flags().StringVar(&compareOut, "out", "./compare", "output directory")
	compareCmd.Flags().Int64Var(&compareSeed, "seed", 0, "RNG seed (kmeans variants only)")
	compareCmd.Flags().IntVar(&compareK, "k", 32, "number of target colors")
	compareCmd.Flags().IntVar(&compareIters, "iters", 10, "iterations (kmeans variants only)")
}

func runCompare(cmd *cobra.Command, args []string) error {
	algos := strings.Split(compareAlgorithms, ",")
	for i, a := range algos {
		algos[i] = strings.TrimSpace(a)
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

	// Header for side-by-side table.
	fmt.Printf("%-12s  %s\n", "Algorithm", "Colors (hex)")
	fmt.Println(strings.Repeat("-", 60))

	for _, algoName := range algos {
		clusterer, err := icolor.Get(algoName)
		if err != nil {
			fmt.Printf("%-12s  error: %v\n", algoName, err)
			continue
		}

		centroids := clusterer.Cluster(pixels, compareK, compareIters, 50000, compareSeed, nil)

		// Save quantized image.
		finalPath := filepath.Join(compareOut, algoName+"-final.png")
		quantized := icolor.QuantizeImage(img, centroids)
		if err := icolor.SavePNG(quantized, finalPath); err != nil {
			return fmt.Errorf("save %s: %w", finalPath, err)
		}
		fmt.Println("Wrote", finalPath)

		// Save palette swatch.
		palettePath := filepath.Join(compareOut, algoName+"-palette.png")
		if err := icolor.SaveSwatchPNG(centroids, 32, 8, palettePath); err != nil {
			return fmt.Errorf("save %s: %w", palettePath, err)
		}
		fmt.Println("Wrote", palettePath)

		// Print color table row.
		hexes := make([]string, len(centroids))
		for i, c := range centroids {
			hexes[i] = fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
		}
		fmt.Printf("%-12s  %s\n", algoName, strings.Join(hexes, " "))
	}

	return nil
}
