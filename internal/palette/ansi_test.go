package palette

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseANSIHex(t *testing.T) {
	hexes := []string{
		"#000000", "#cd0000", "#00cd00", "#cdcd00", "#0000ee", "#cd00cd", "#00cdcd", "#e5e5e5",
		"#7f7f7f", "#ff0000", "#00ff00", "#ffff00", "#5c5cff", "#ff00ff", "#00ffff", "#ffffff",
	}
	out, err := ParseANSIHex(hexes)
	require.NoError(t, err)
	require.Len(t, out, 16)
	assert.True(t, out[0].R == 0 && out[0].G == 0 && out[0].B == 0, "first (black) = %v", out[0])
	assert.True(t, out[15].R == 255 && out[15].G == 255 && out[15].B == 255, "last (white) = %v", out[15])
	// without #
	hexes2 := []string{"000000", "ffffff"}
	for i := 0; i < 14; i++ {
		hexes2 = append(hexes2, "808080")
	}
	_, err = ParseANSIHex(hexes2)
	require.NoError(t, err)
}

func TestParseANSIHex_Errors(t *testing.T) {
	_, err := ParseANSIHex(nil)
	assert.Error(t, err, "expected error for nil")
	_, err = ParseANSIHex([]string{"#000"})
	assert.Error(t, err, "expected error for wrong count")
	short := make([]string, 16)
	for i := range short {
		short[i] = "abc"
	}
	_, err = ParseANSIHex(short)
	assert.Error(t, err, "expected error for 3-char hex")
}

func TestLoadANSIHexFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ansi.txt")
	content := "# my colors\n000000\ncd0000\n00cd00\ncdcd00\n0000ee\ncd00cd\n00cdcd\ne5e5e5\n7f7f7f\nff0000\n00ff00\nffff00\n5c5cff\nff00ff\n00ffff\nffffff\n"
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
	out, err := LoadANSIHexFile(path)
	require.NoError(t, err)
	require.Len(t, out, 16)
	assert.True(t, out[0].R == 0 && out[1].R == 0xcd, "out[0]=%v out[1]=%v", out[0], out[1])
}
