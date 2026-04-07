package themes

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	reColorRefBG = regexp.MustCompile(`(?i)^\s*\.\s*Background\s*=\s*#?([0-9a-fA-F]{6})\s*$`)
	reColorRefFG = regexp.MustCompile(`(?i)^\s*\.\s*Foreground\s*=\s*#?([0-9a-fA-F]{6})\s*$`)
	reColorRefAnsi = regexp.MustCompile(`(?i)^\s*\.\s*Ansi\s+(\d+)\s*=\s*#?([0-9a-fA-F]{6})\s*$`)
	reColorRefTones = regexp.MustCompile(`(?i)^\s*\.\s*Tones\s+(\d+)\s*=\s*#?([0-9a-fA-F]{6})\s*$`)
	reColorRefTFA = regexp.MustCompile(`(?i)^\s*\.\s*TonesFromANSI\s+(\d+)\s*=\s*#?([0-9a-fA-F]{6})\s*$`)
)

// ParseColorReference parses a Gowall Color Reference document (same format as
// exportColorReference writes). Lines starting with # are comments except for
// assignment lines that begin with a dot. Hex may include or omit a leading #.
func ParseColorReference(data string) (ThemeData, error) {
	var bg, fg string
	ansi := make([]string, 16)
	tones := make([]string, 16)
	tfa := make([]string, 16)
	var tfaSeen [16]bool

	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if strings.HasPrefix(trim, "#") {
			continue
		}

		if m := reColorRefBG.FindStringSubmatch(trim); m != nil {
			bg = strings.ToLower(m[1])
			continue
		}
		if m := reColorRefFG.FindStringSubmatch(trim); m != nil {
			fg = strings.ToLower(m[1])
			continue
		}
		if m := reColorRefAnsi.FindStringSubmatch(trim); m != nil {
			idx := parseIdx(m[1], 0, 15)
			if idx < 0 {
				return ThemeData{}, fmt.Errorf("color reference: invalid ANSI index %q", m[1])
			}
			ansi[idx] = strings.ToLower(m[2])
			continue
		}
		if m := reColorRefTones.FindStringSubmatch(trim); m != nil {
			idx := parseIdx(m[1], 0, 15)
			if idx < 0 {
				return ThemeData{}, fmt.Errorf("color reference: invalid Tones index %q", m[1])
			}
			tones[idx] = strings.ToLower(m[2])
			continue
		}
		if m := reColorRefTFA.FindStringSubmatch(trim); m != nil {
			idx := parseIdx(m[1], 0, 15)
			if idx < 0 {
				return ThemeData{}, fmt.Errorf("color reference: invalid TonesFromANSI index %q", m[1])
			}
			tfa[idx] = strings.ToLower(m[2])
			tfaSeen[idx] = true
			continue
		}
	}

	if bg == "" || fg == "" {
		return ThemeData{}, fmt.Errorf("color reference: missing .Background or .Foreground")
	}
	for i := range ansi {
		if ansi[i] == "" {
			return ThemeData{}, fmt.Errorf("color reference: missing .Ansi %d", i)
		}
	}
	for i := range tones {
		if tones[i] == "" {
			return ThemeData{}, fmt.Errorf("color reference: missing .Tones %d", i)
		}
	}

	var outTFA []string
	anyTFA := false
	for _, v := range tfaSeen {
		if v {
			anyTFA = true
			break
		}
	}
	if anyTFA {
		for i := range tfa {
			if !tfaSeen[i] || tfa[i] == "" {
				return ThemeData{}, fmt.Errorf("color reference: missing .TonesFromANSI %d", i)
			}
		}
		outTFA = tfa
	}

	return ThemeData{
		Background:    bg,
		Foreground:    fg,
		Ansi:          ansi,
		Tones:         tones,
		TonesFromANSI: outTFA,
	}, nil
}

func parseIdx(s string, min, max int) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil || n < min || n > max {
		return -1
	}
	return n
}

// LoadColorReferenceFile reads and parses a color reference file.
func LoadColorReferenceFile(path string) (ThemeData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ThemeData{}, err
	}
	return ParseColorReference(string(data))
}
