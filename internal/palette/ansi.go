package palette

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/0mega24/gowall/internal/color"
)

// ParseANSIHex parses 16 hex color strings (with or without #) into ANSI-compliant Centroids.
// Each string must be 6 hex digits; returns error if not exactly 16 or any parse fails.
func ParseANSIHex(hexes []string) ([]color.Centroid, error) {
	if len(hexes) != 16 {
		return nil, fmt.Errorf("need exactly 16 ANSI colors, got %d", len(hexes))
	}
	out := make([]color.Centroid, 16)
	for i, s := range hexes {
		s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
		if len(s) != 6 {
			return nil, fmt.Errorf("ANSI color %d: need 6 hex digits, got %q", i, s)
		}
		r, err := strconv.ParseUint(s[0:2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("ANSI color %d: %w", i, err)
		}
		g, _ := strconv.ParseUint(s[2:4], 16, 8)
		b, _ := strconv.ParseUint(s[4:6], 16, 8)
		out[i] = color.Centroid{R: uint8(r), G: uint8(g), B: uint8(b)}
	}
	return out, nil
}

// LoadANSIHexFile reads a file with one hex color per line. Blank lines and comment lines
// (lines starting with # that are not #rrggbb) are skipped. Use for -ansi-colors path.
func LoadANSIHexFile(path string) ([]color.Centroid, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var hexes []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if len(line) == 7 && isHex6(line[1:]) {
				hexes = append(hexes, line[1:])
			}
			continue
		}
		hexes = append(hexes, strings.TrimPrefix(line, "#"))
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return ParseANSIHex(hexes)
}

func isHex6(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, c := range s {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return false
	}
	return true
}

// StandardANSI16 is the classic 16 ANSI/terminal colors (xterm-style).
var StandardANSI16 = []color.Centroid{
	{R: 0, G: 0, B: 0},
	{R: 205, G: 0, B: 0},
	{R: 0, G: 205, B: 0},
	{R: 205, G: 205, B: 0},
	{R: 0, G: 0, B: 238},
	{R: 205, G: 0, B: 205},
	{R: 0, G: 205, B: 205},
	{R: 229, G: 229, B: 229},
	{R: 127, G: 127, B: 127},
	{R: 255, G: 0, B: 0},
	{R: 0, G: 255, B: 0},
	{R: 255, G: 255, B: 0},
	{R: 92, G: 92, B: 255},
	{R: 255, G: 0, B: 255},
	{R: 0, G: 255, B: 255},
	{R: 255, G: 255, B: 255},
}

// RetoneANSIToPalette retones the 16 ANSI colors to match the palette’s S and V (HSV):
// each color keeps its hue but gets the palette’s average saturation and value.
// If bg is non-nil, each retoned color is adjusted (by raising V) so it meets MinANSIContrast
// with the background, so the result stays readable without later passes washing out the look.
func RetoneANSIToPalette(std, paletteRef []color.Centroid, bg *color.Centroid) []color.Centroid {
	if len(std) != 16 || len(paletteRef) == 0 {
		return std
	}
	var sumS, sumV, weight float32
	for _, c := range paletteRef {
		r, g, b := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255
		_, s, v := rgbToHsv(r, g, b)
		w := s + 0.1
		sumS += s * w
		sumV += v * w
		weight += w
	}
	if weight <= 0 {
		return std
	}
	avgS := sumS / weight
	avgV := sumV / weight
	if avgS > 0.85 {
		avgS = 0.85
	}
	out := make([]color.Centroid, 16)
	blend := float32(0.5)
	for i, c := range std {
		r, g, b := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255
		h, s, v := rgbToHsv(r, g, b)
		// Keep hue; match palette S and V.
		s = s*(1-blend) + avgS*blend
		v = v*(1-blend) + avgV*blend
		if s > 1 {
			s = 1
		}
		if v > 1 {
			v = 1
		}
		rr, gg, bb := hsvToRgb(h, s, v)
		out[i] = color.Centroid{R: clampByte(rr * 255), G: clampByte(gg * 255), B: clampByte(bb * 255)}
		// If we have a background, ensure this color meets contrast by raising V (keeps H and S).
		if bg != nil && ContrastRatio(out[i], *bg) < MinANSIContrast {
			vLow, vHigh := v, float32(1)
			for step := 0; step < 20; step++ {
				vm := (vLow + vHigh) / 2
				rr, gg, bb = hsvToRgb(h, s, vm)
				try := color.Centroid{R: clampByte(rr * 255), G: clampByte(gg * 255), B: clampByte(bb * 255)}
				if ContrastRatio(try, *bg) >= MinANSIContrast {
					vHigh = vm
					out[i] = try
				} else {
					vLow = vm
				}
			}
		}
	}
	return out
}
