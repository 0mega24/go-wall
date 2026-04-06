package main

import (
	"fmt"

	"github.com/0mega24/gowall/internal/cli"
	"github.com/0mega24/gowall/internal/wallpaper"
	"github.com/0mega24/gowall/pkg/gowall"
	"github.com/spf13/cobra"
)

var paletteFormat string

var paletteCmd = &cobra.Command{
	Use:   "palette",
	Short: "Work with color palettes",
}

var paletteShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current wallpaper palette",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := wallpaper.CurrentWallpaperPath()
		if err != nil {
			return fmt.Errorf("wallpaper detection failed: %w", err)
		}
		opts := gowall.DefaultOptions()
		result, err := gowall.RunFromPath(path, opts)
		if err != nil {
			return err
		}
		switch paletteFormat {
		case "gradient":
			fmt.Println(cli.GradientRow(result.Filtered, 80))
		default:
			fmt.Println(cli.RowOfSwatchHex(result.Filtered, 4, "  "))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(paletteCmd)
	paletteCmd.AddCommand(paletteShowCmd)
	paletteShowCmd.Flags().StringVar(&paletteFormat, "format", "labeled", "output format: labeled or gradient")
}
