package main

import (
	"fmt"
	"os"

	"github.com/0mega24/gowall/internal/cli"
	"github.com/0mega24/gowall/internal/themes"
	"github.com/0mega24/gowall/internal/wallpaper"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage color theme templates",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := themes.EnsureTemplatesInstalled(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
		cli.ListTemplates()
		templates, err := themes.DiscoverTemplates()
		if err != nil {
			return err
		}
		fmt.Println("\nDiscovered templates:")
		for _, t := range templates {
			kind := "user"
			if t.IsBuiltin {
				kind = "builtin"
			}
			fmt.Printf("  %-20s %s → %s\n", t.ID, kind, t.TargetPath)
		}
		return nil
	},
}

var templateApplyCmd = &cobra.Command{
	Use:   "apply <id>",
	Short: "Apply a template with the current wallpaper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := themes.EnsureTemplatesInstalled(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
		path, err := wallpaper.CurrentWallpaperPath()
		if err != nil {
			return fmt.Errorf("wallpaper detection failed: %w", err)
		}
		cfg := cli.Config{
			ImagePath: path,
			Templates: id,
		}
		return cli.Run(cfg)
	},
}

var templatePreviewCmd = &cobra.Command{
	Use:   "preview <id>",
	Short: "Preview a template with the current wallpaper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement preview (render template to stdout)
		fmt.Println("preview not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd, templateApplyCmd, templatePreviewCmd)
}
