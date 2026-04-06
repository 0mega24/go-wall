package main

import (
	"fmt"
	"os"

	"github.com/0mega24/gowall/internal/config"
	"github.com/spf13/cobra"
)

var (
	rootPreset string
	loadedCfg  config.Config
)

var rootCmd = &cobra.Command{
	Use:   "gowall",
	Short: "Generate terminal color themes from your wallpaper",
	Long:  "gowall extracts a color palette from your wallpaper and applies it to terminal applications.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
			return nil
		}
		loadedCfg = cfg
		return nil
	},
}

func init() {
	rootCmd.InitDefaultCompletionCmd()
	rootCmd.PersistentFlags().StringVar(&rootPreset, "preset", "", "apply a named preset from config")
}
