package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/0mega24/gowall/internal/config"
	"github.com/spf13/cobra"
)

var presetCmd = &cobra.Command{
	Use:   "preset",
	Short: "Manage named presets",
}

var presetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Presets) == 0 {
			fmt.Println("No presets saved.")
			return nil
		}
		names := make([]string, 0, len(cfg.Presets))
		for n := range cfg.Presets {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			p := cfg.Presets[n]
			var parts []string
			if p.Background != "" {
				parts = append(parts, "background="+p.Background)
			}
			if p.Pipeline.Algorithm != "" {
				parts = append(parts, "algorithm="+p.Pipeline.Algorithm)
			}
			if len(p.Templates) > 0 {
				parts = append(parts, "templates="+strings.Join(p.Templates, ","))
			}
			if p.RetoneANSI {
				parts = append(parts, "retone-ansi=true")
			}
			detail := strings.Join(parts, "  ")
			if detail == "" {
				detail = "(no settings)"
			}
			fmt.Printf("  %-20s  %s\n", n, detail)
		}
		return nil
	},
}

var presetSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save current flags as a named preset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.Presets == nil {
			cfg.Presets = make(map[string]config.Preset)
		}

		// Build preset from current run flags.
		p := config.Preset{Defaults: config.Defaults{
			Background: runBackground,
			RetoneANSI: runRetoneANSI,
		}}
		if runAlgorithm != "" {
			p.Pipeline.Algorithm = runAlgorithm
		}
		if runTemplates != "" {
			for _, t := range strings.Split(runTemplates, ",") {
				if t = strings.TrimSpace(t); t != "" {
					p.Templates = append(p.Templates, t)
				}
			}
		}

		cfg.Presets[name] = p
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Saved preset %q\n", name)
		return nil
	},
}

var presetApplyCmd = &cobra.Command{
	Use:   "apply <name> [image]",
	Short: "Run gowall with a named preset",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runPresetApply,
}

func runPresetApply(cmd *cobra.Command, args []string) error {
	presetName := args[0]
	if len(args) > 1 {
		return runRunWithPreset(presetName, args[1:])
	}
	return runRunWithPreset(presetName, nil)
}

var presetDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Remove a named preset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if _, ok := cfg.Presets[name]; !ok {
			return fmt.Errorf("preset %q not found", name)
		}
		delete(cfg.Presets, name)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Deleted preset %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(presetCmd)
	presetCmd.AddCommand(presetListCmd)
	presetCmd.AddCommand(presetSaveCmd)
	presetCmd.AddCommand(presetApplyCmd)
	presetCmd.AddCommand(presetDeleteCmd)
}

// runRunWithPreset runs the pipeline with a named preset applied.
func runRunWithPreset(presetName string, extraArgs []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var imagePath string
	if len(extraArgs) > 0 {
		imagePath = extraArgs[0]
	}

	base := config.CLIConfig{ImagePath: imagePath}
	merged, err := config.MergePreset(base, cfg, presetName)
	if err != nil {
		return err
	}

	// Re-use the run machinery via the run flags.
	rootPreset = presetName
	loadedCfg = cfg

	runAlgorithm = merged.Algorithm
	runBackground = merged.BackgroundHex
	runRetoneANSI = merged.RetoneANSI
	if merged.Templates != "" {
		runTemplates = merged.Templates
	}

	return runRun(presetSaveCmd, extraArgs)
}
