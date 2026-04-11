package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the full gowall config file structure.
type Config struct {
	Defaults Defaults          `toml:"defaults"`
	Presets  map[string]Preset `toml:"presets"`
}

// Defaults holds settings applied on every run unless overridden by CLI flags.
type Defaults struct {
	Templates   []string                `toml:"templates"`
	RetoneANSI  bool                    `toml:"retone_ansi"`
	Background  string                  `toml:"background"`
	Algorithm   string                  `toml:"algorithm"`
	Pipeline    PipelineDefaults        `toml:"pipeline"`
	Constraints map[int]ConstraintEntry `toml:"constraints"`
	HideBuiltin bool `toml:"hide_builtin"`
	// GlobalTweakH: hue shift in degrees. GlobalTweakS/V: percent multipliers on S/V (see pipeline.GlobalAdjust).
	GlobalTweakH float64 `toml:"global_tweak_h"`
	GlobalTweakS float64 `toml:"global_tweak_s"` // Sat%  (multiply S by 1+S/100)
	GlobalTweakV float64 `toml:"global_tweak_v"` // Val% (multiply V by 1+V/100)
}

// PipelineDefaults holds pipeline tuning parameters.
type PipelineDefaults struct {
	Algorithm  string `toml:"algorithm"`
	K          int    `toml:"k"`
	Iters      int    `toml:"iters"`
	MaxSamples int    `toml:"max_samples"`
}

// ConstraintEntry holds per-slot color constraint configuration from TOML.
type ConstraintEntry struct {
	Pin    string   `toml:"pin"`      // hex string, empty = not set
	LockH  *float64 `toml:"lock_hue"` // nil = free
	LockS  *float64 `toml:"lock_sat"` // nil = free
	LockV  *float64 `toml:"lock_val"` // nil = free
	TweakH float64  `toml:"tweak_h"`
	TweakS float64  `toml:"tweak_s"`
	TweakV float64  `toml:"tweak_v"`
}

// Preset embeds Defaults so it can override any subset of the default settings.
type Preset struct {
	Defaults
}

// configPath returns the path to the gowall config file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gowall", "config.toml"), nil
}

// Load reads the gowall config file. Returns an empty Config if the file is absent.
func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

// Save writes the config to the gowall config file, creating parent dirs as needed.
func Save(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return toml.NewEncoder(f).Encode(cfg)
}

// ConfigPath returns the path to the gowall config file (for display purposes).
func ConfigPath() (string, error) {
	return configPath()
}

// tomlEncode encodes cfg to w (used by tests and Save).
func tomlEncode(w io.Writer, cfg Config) error {
	return toml.NewEncoder(w).Encode(cfg)
}

// tomlDecode decodes a Config from path (used by tests and Load).
func tomlDecode(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

// GetPreset returns the named preset or an error if it does not exist.
func GetPreset(cfg Config, name string) (Preset, error) {
	p, ok := cfg.Presets[name]
	if !ok {
		return Preset{}, fmt.Errorf("preset %q not found", name)
	}
	return p, nil
}
