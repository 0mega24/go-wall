package main

import (
	"fmt"
	"os"

	"github.com/0mega24/gowall/internal/wallpaper"
	"github.com/spf13/cobra"
)

var wallpaperCmd = &cobra.Command{
	Use:   "wallpaper",
	Short: "Wallpaper detection and management",
}

var wallpaperDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect and print the current wallpaper path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, sourceName, err := wallpaper.FirstOf(wallpaper.DefaultSources()...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: no wallpaper detected: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s (source: %s)\n", path, sourceName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(wallpaperCmd)
	wallpaperCmd.AddCommand(wallpaperDetectCmd)
}
