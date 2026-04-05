package main

import (
	"github.com/0mega24/gowall/internal/cli"
	"github.com/0mega24/gowall/internal/wallpaper"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui [image]",
	Short: "Open the interactive terminal UI",
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	var imagePath string
	if len(args) > 0 {
		imagePath = args[0]
	} else {
		path, err := wallpaper.CurrentWallpaperPath()
		if err != nil {
			// TUI can handle empty path (it has a "need-path" state)
			imagePath = ""
		} else {
			imagePath = path
		}
	}
	return cli.RunTUI(imagePath)
}
