package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/0mega24/gowall/internal/cli"
	"github.com/0mega24/gowall/internal/config"
	"github.com/0mega24/gowall/internal/pipeline"
	"github.com/0mega24/gowall/internal/themes"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [image]",
	Short: "Generate and apply a color theme from a wallpaper",
	RunE:  runRun,
}

var (
	runTemplates    string
	runTemplateList string
	runRetoneANSI   bool
	runANSIColors   string
	runQuiet        bool
	runSeed         int64
	runAlgorithm    string   // clustering algorithm
	runPinSlot      []string // e.g. ["1=#cc0000"]
	runLockHue      []string // e.g. ["1=0"]
	runLockSat      []string // e.g. ["1=0.8"]
	runLockVal      []string // e.g. ["0=0.05"]
	runTweakSlot    []string // e.g. ["3=+30,+0.1,0"]
	runBackground   string   // e.g. "#0a0a0a"
	runGlobalTweak  string   // e.g. "0,-30,-40" — hue deg, sat %, val % (see --global-tweak help)
	runColorRef     string   // Gowall Color Reference file (skip image; apply templates from file)
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(&runTemplates, "templates", "alacritty,polybar,rofi", "comma-separated template IDs to apply")
	runCmd.Flags().StringVar(&runTemplateList, "template-list", "", "path to file listing templates")
	runCmd.Flags().BoolVar(&runRetoneANSI, "retone-ansi", false, "retone standard ANSI 16 colors to match wallpaper")
	runCmd.Flags().StringVar(&runANSIColors, "ansi-colors", "", "path to file with 16 custom hex ANSI colors")
	runCmd.Flags().BoolVar(&runQuiet, "quiet", false, "minimal output")
	runCmd.Flags().Int64Var(&runSeed, "seed", 0, "RNG seed for deterministic results")
	runCmd.Flags().StringVar(&runAlgorithm, "algorithm", "", "clustering algorithm (kmeans++, kmeans, mediancut, octree)")
	runCmd.Flags().StringArrayVar(&runPinSlot, "pin-slot", nil, "pin ANSI slot to exact color: slot=hex (repeatable)")
	runCmd.Flags().StringArrayVar(&runLockHue, "lock-hue", nil, "lock slot hue channel: slot=degrees (repeatable)")
	runCmd.Flags().StringArrayVar(&runLockSat, "lock-sat", nil, "lock slot saturation channel: slot=value (repeatable)")
	runCmd.Flags().StringArrayVar(&runLockVal, "lock-val", nil, "lock slot HSV value channel: slot=value (repeatable)")
	runCmd.Flags().StringArrayVar(&runTweakSlot, "tweak-slot", nil, "post-gen delta: slot=dH,dS,dV (repeatable)")
	runCmd.Flags().StringVar(&runGlobalTweak, "global-tweak", "", "global ANSI adjust: hueDeg,satPct,valPct — rotation in degrees, then S×(1+sat/100), V×(1+val/100)")
	runCmd.Flags().StringVar(&runBackground, "background", "", "override background color (#rrggbb)")
	runCmd.Flags().StringVar(&runColorRef, "color-reference", "", "load theme from a Gowall Color Reference file (skips image; same format as palette export)")
}

func runRun(cmd *cobra.Command, args []string) error {
	if err := themes.EnsureTemplatesInstalled(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not install templates: %v\n", err)
	}

	var imagePath string
	if len(args) > 0 {
		imagePath = args[0]
	}

	var templateStr string
	if runTemplates != "" {
		var ids []string
		for _, id := range strings.Split(runTemplates, ",") {
			if t := strings.TrimSpace(id); t != "" {
				ids = append(ids, t)
			}
		}
		templateStr = strings.Join(ids, ",")
	}

	constraints, err := buildConstraints()
	if err != nil {
		return err
	}

	globalAdjust, err := parseGlobalAdjust(runGlobalTweak)
	if err != nil {
		return err
	}

	cliCfg := config.CLIConfig{
		ImagePath:     imagePath,
		Templates:     templateStr,
		TemplateList:  runTemplateList,
		RetoneANSI:    runRetoneANSI,
		Quiet:         runQuiet,
		Seed:          runSeed,
		Algorithm:     runAlgorithm,
		Constraints:   constraints,
		BackgroundHex: runBackground,
		GlobalAdjust:  globalAdjust,
	}

	// Apply config file defaults, then preset (if requested), CLI flags win.
	if rootPreset != "" {
		cliCfg, err = config.MergePreset(cliCfg, loadedCfg, rootPreset)
		if err != nil {
			return err
		}
	} else {
		cliCfg = config.Merge(cliCfg, loadedCfg)
	}

	cfg := cli.Config{
		ImagePath:          cliCfg.ImagePath,
		ColorReferencePath: runColorRef,
		Templates:          cliCfg.Templates,
		TemplateList:       cliCfg.TemplateList,
		RetoneANSI:         cliCfg.RetoneANSI,
		CustomANSIColors:   runANSIColors,
		Quiet:              cliCfg.Quiet,
		Seed:               cliCfg.Seed,
		Algorithm:          cliCfg.Algorithm,
		Constraints:        cliCfg.Constraints,
		BackgroundHex:      cliCfg.BackgroundHex,
		GlobalAdjust:       cliCfg.GlobalAdjust,
	}
	return cli.Run(cfg)
}

func parseGlobalAdjust(s string) (pipeline.GlobalAdjust, error) {
	if strings.TrimSpace(s) == "" {
		return pipeline.GlobalAdjust{}, nil
	}
	dvals := strings.Split(s, ",")
	if len(dvals) != 3 {
		return pipeline.GlobalAdjust{}, fmt.Errorf("--global-tweak: expected hueDeg,satPct,valPct (three comma-separated numbers)")
	}
	dH, err1 := strconv.ParseFloat(strings.TrimSpace(dvals[0]), 64)
	dS, err2 := strconv.ParseFloat(strings.TrimSpace(dvals[1]), 64)
	dV, err3 := strconv.ParseFloat(strings.TrimSpace(dvals[2]), 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return pipeline.GlobalAdjust{}, fmt.Errorf("--global-tweak: values must be numbers")
	}
	return pipeline.GlobalAdjust{HueDeg: dH, SatPct: dS, ValPct: dV}, nil
}

// buildConstraints parses all constraint flags into a map[int]SlotConstraint.
func buildConstraints() (map[int]pipeline.SlotConstraint, error) {
	constraints := make(map[int]pipeline.SlotConstraint)

	getOrNew := func(slot int) pipeline.SlotConstraint {
		sc := constraints[slot]
		return sc
	}

	for _, p := range runPinSlot {
		slot, hexStr, err := parseSlotValue(p)
		if err != nil {
			return nil, fmt.Errorf("--pin-slot %q: %w", p, err)
		}
		c, err := pipeline.ParseHex(hexStr)
		if err != nil {
			return nil, fmt.Errorf("--pin-slot %q: %w", p, err)
		}
		sc := getOrNew(slot)
		sc.Pin = &c
		constraints[slot] = sc
	}

	for _, p := range runLockHue {
		slot, val, err := parseSlotFloat(p)
		if err != nil {
			return nil, fmt.Errorf("--lock-hue %q: %w", p, err)
		}
		sc := getOrNew(slot)
		sc.LockH = &val
		constraints[slot] = sc
	}

	for _, p := range runLockSat {
		slot, val, err := parseSlotFloat(p)
		if err != nil {
			return nil, fmt.Errorf("--lock-sat %q: %w", p, err)
		}
		sc := getOrNew(slot)
		sc.LockS = &val
		constraints[slot] = sc
	}

	for _, p := range runLockVal {
		slot, val, err := parseSlotFloat(p)
		if err != nil {
			return nil, fmt.Errorf("--lock-val %q: %w", p, err)
		}
		sc := getOrNew(slot)
		sc.LockV = &val
		constraints[slot] = sc
	}

	for _, p := range runTweakSlot {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("--tweak-slot %q: expected slot=dH,dS,dV", p)
		}
		slot, err := parseSlot(parts[0])
		if err != nil {
			return nil, fmt.Errorf("--tweak-slot %q: %w", p, err)
		}
		dvals := strings.Split(parts[1], ",")
		if len(dvals) != 3 {
			return nil, fmt.Errorf("--tweak-slot %q: expected 3 delta values dH,dS,dV", p)
		}
		dH, err1 := strconv.ParseFloat(strings.TrimSpace(dvals[0]), 64)
		dS, err2 := strconv.ParseFloat(strings.TrimSpace(dvals[1]), 64)
		dV, err3 := strconv.ParseFloat(strings.TrimSpace(dvals[2]), 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, fmt.Errorf("--tweak-slot %q: delta values must be numbers", p)
		}
		sc := getOrNew(slot)
		sc.Tweak.DeltaH += dH
		sc.Tweak.DeltaS += dS
		sc.Tweak.DeltaV += dV
		constraints[slot] = sc
	}

	if len(constraints) == 0 {
		return nil, nil
	}
	return constraints, nil
}

func parseSlot(s string) (int, error) {
	s = strings.TrimSpace(s)
	slot, err := strconv.Atoi(s)
	if err != nil || slot < 0 || slot > 15 {
		return 0, fmt.Errorf("slot must be 0-15, got %q", s)
	}
	return slot, nil
}

func parseSlotValue(p string) (int, string, error) {
	parts := strings.SplitN(p, "=", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("expected slot=value")
	}
	slot, err := parseSlot(parts[0])
	if err != nil {
		return 0, "", err
	}
	return slot, parts[1], nil
}

func parseSlotFloat(p string) (int, float64, error) {
	slot, valStr, err := parseSlotValue(p)
	if err != nil {
		return 0, 0, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(valStr), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("value must be a number, got %q", valStr)
	}
	return slot, val, nil
}
