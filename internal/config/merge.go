package config

import (
	"github.com/0mega24/gowall/internal/pipeline"
)

func globalAdjustIsZero(g pipeline.GlobalAdjust) bool {
	return g.IsZero()
}

// CLIConfig is a subset interface matching cli.Config fields we can merge into.
// We use a concrete struct to avoid an import cycle.
type CLIConfig struct {
	ImagePath        string
	Templates        string
	TemplateList     string
	RetoneANSI       bool
	CustomANSIColors string
	Quiet            bool
	Seed             int64
	Algorithm        string
	Constraints      map[int]pipeline.SlotConstraint
	BackgroundHex    string
	GlobalAdjust pipeline.GlobalAdjust
}

// Merge applies config defaults to a CLIConfig.
// Zero/empty CLI values are filled in from config defaults; non-zero CLI values are preserved.
func Merge(cli CLIConfig, cfg Config) CLIConfig {
	return mergeDefaults(cli, cfg.Defaults)
}

// MergePreset applies the named preset on top of defaults, then merges with CLI config.
// CLI flags still take precedence over preset and defaults.
func MergePreset(cli CLIConfig, cfg Config, presetName string) (CLIConfig, error) {
	preset, err := GetPreset(cfg, presetName)
	if err != nil {
		return cli, err
	}
	// Build base from defaults, then let preset override (not just fill), then CLI wins.
	base := mergeDefaults(CLIConfig{}, cfg.Defaults)
	base = applyPreset(base, preset.Defaults)
	return mergeOver(cli, base), nil
}

// applyPreset overrides base fields with non-zero preset values.
func applyPreset(base CLIConfig, d Defaults) CLIConfig {
	if d.Pipeline.Algorithm != "" {
		base.Algorithm = d.Pipeline.Algorithm
	}
	if d.RetoneANSI {
		base.RetoneANSI = d.RetoneANSI
	}
	if d.Background != "" {
		base.BackgroundHex = d.Background
	}
	if len(d.Templates) > 0 {
		joined := ""
		for i, t := range d.Templates {
			if i > 0 {
				joined += ","
			}
			joined += t
		}
		base.Templates = joined
	}
	if len(d.Constraints) > 0 {
		base.Constraints = convertConstraints(d.Constraints)
	}
	if d.GlobalTweakH != 0 || d.GlobalTweakS != 0 || d.GlobalTweakV != 0 {
		base.GlobalAdjust = pipeline.GlobalAdjust{
			HueDeg: d.GlobalTweakH,
			SatPct: d.GlobalTweakS,
			ValPct: d.GlobalTweakV,
		}
	}
	return base
}

// mergeDefaults fills zero-value fields in cli from d.
func mergeDefaults(cli CLIConfig, d Defaults) CLIConfig {
	if cli.Algorithm == "" && d.Pipeline.Algorithm != "" {
		cli.Algorithm = d.Pipeline.Algorithm
	}
	if !cli.RetoneANSI && d.RetoneANSI {
		cli.RetoneANSI = d.RetoneANSI
	}
	if cli.BackgroundHex == "" && d.Background != "" {
		cli.BackgroundHex = d.Background
	}
	if cli.Templates == "" && len(d.Templates) > 0 {
		joined := ""
		for i, t := range d.Templates {
			if i > 0 {
				joined += ","
			}
			joined += t
		}
		cli.Templates = joined
	}
	if len(d.Constraints) > 0 && cli.Constraints == nil {
		cli.Constraints = convertConstraints(d.Constraints)
	}
	if globalAdjustIsZero(cli.GlobalAdjust) && (d.GlobalTweakH != 0 || d.GlobalTweakS != 0 || d.GlobalTweakV != 0) {
		cli.GlobalAdjust = pipeline.GlobalAdjust{
			HueDeg: d.GlobalTweakH,
			SatPct: d.GlobalTweakS,
			ValPct: d.GlobalTweakV,
		}
	}
	return cli
}

// mergeOver overlays non-zero fields from src onto base (CLI wins).
func mergeOver(src, base CLIConfig) CLIConfig {
	if src.Algorithm != "" {
		base.Algorithm = src.Algorithm
	}
	if src.RetoneANSI {
		base.RetoneANSI = src.RetoneANSI
	}
	if src.BackgroundHex != "" {
		base.BackgroundHex = src.BackgroundHex
	}
	if src.Templates != "" {
		base.Templates = src.Templates
	}
	if src.Seed != 0 {
		base.Seed = src.Seed
	}
	if src.ImagePath != "" {
		base.ImagePath = src.ImagePath
	}
	if src.TemplateList != "" {
		base.TemplateList = src.TemplateList
	}
	if src.CustomANSIColors != "" {
		base.CustomANSIColors = src.CustomANSIColors
	}
	if src.Quiet {
		base.Quiet = src.Quiet
	}
	if len(src.Constraints) > 0 {
		base.Constraints = src.Constraints
	}
	if !globalAdjustIsZero(src.GlobalAdjust) {
		base.GlobalAdjust = src.GlobalAdjust
	}
	return base
}

// convertConstraints converts config ConstraintEntry map to pipeline.SlotConstraint map.
func convertConstraints(entries map[int]ConstraintEntry) map[int]pipeline.SlotConstraint {
	if len(entries) == 0 {
		return nil
	}
	out := make(map[int]pipeline.SlotConstraint, len(entries))
	for slot, e := range entries {
		sc := pipeline.SlotConstraint{
			LockH: e.LockH,
			LockS: e.LockS,
			LockV: e.LockV,
			Tweak: pipeline.SlotTweak{
				DeltaH: e.TweakH,
				DeltaS: e.TweakS,
				DeltaV: e.TweakV,
			},
		}
		if e.Pin != "" {
			c, err := pipeline.ParseHex(e.Pin)
			if err == nil {
				sc.Pin = &c
			}
		}
		out[slot] = sc
	}
	return out
}
