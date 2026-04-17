package pipeline

import (
	"fmt"

	"github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/themes"
)

// ResultFromThemeData builds a Result from theme data only (no image extraction).
// Filtered is empty. Used when loading a Gowall Color Reference file.
func ResultFromThemeData(theme themes.ThemeData) (Result, error) {
	if theme.Background == "" || theme.Foreground == "" {
		return Result{}, fmt.Errorf("theme: background and foreground are required")
	}
	if len(theme.Ansi) != 16 || len(theme.Tones) != 16 {
		return Result{}, fmt.Errorf("theme: require 16 ANSI and 16 tone colors")
	}
	ansi := make([]color.Centroid, 16)
	tones := make([]color.Centroid, 16)
	for i := 0; i < 16; i++ {
		c, err := ParseHex(theme.Ansi[i])
		if err != nil {
			return Result{}, fmt.Errorf("ANSI %d: %w", i, err)
		}
		ansi[i] = c
	}
	for i := 0; i < 16; i++ {
		c, err := ParseHex(theme.Tones[i])
		if err != nil {
			return Result{}, fmt.Errorf("tones %d: %w", i, err)
		}
		tones[i] = c
	}
	var tfa []color.Centroid
	if len(theme.TonesFromANSI) == 16 {
		tfa = make([]color.Centroid, 16)
		for i := 0; i < 16; i++ {
			c, err := ParseHex(theme.TonesFromANSI[i])
			if err != nil {
				return Result{}, fmt.Errorf("TonesFromANSI %d: %w", i, err)
			}
			tfa[i] = c
		}
	}
	return Result{
		Theme:         theme,
		Filtered:      nil,
		ANSI:          ansi,
		Tones:         tones,
		TonesFromANSI: tfa,
	}, nil
}
