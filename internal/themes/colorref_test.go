package themes

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseColorReference_sample(t *testing.T) {
	const sample = `# Gowall Color Reference
# Use in templates as: {{ .Background }}, {{ index .Ansi 3 }}, etc.

.Background          = #192224
.Foreground          = #dde2e3

# ANSI slots  ({{ index .Ansi N }})
.Ansi 0               = #6e6e6e
.Ansi 1               = #f41835
.Ansi 2               = #ff0f21
.Ansi 3               = #f56d77
.Ansi 4               = #08b7bb
.Ansi 5               = #07c3ca
.Ansi 6               = #12cecf
.Ansi 7               = #37d7dc
.Ansi 8               = #cccccc
.Ansi 9               = #ff1837
.Ansi 10              = #ff0f20
.Ansi 11              = #ff717d
.Ansi 12              = #0cf8ff
.Ansi 13              = #0af2ff
.Ansi 14              = #15fdff
.Ansi 15              = #3ef7ff

# Tone ramp bg→fg  ({{ index .Tones N }})
.Tones 0              = #192224
.Tones 1              = #242f33
.Tones 2              = #2f4042
.Tones 3              = #394d51
.Tones 4              = #455b60
.Tones 5              = #4f686f
.Tones 6              = #5a797e
.Tones 7              = #65868d
.Tones 8              = #72939a
.Tones 9              = #819ea5
.Tones 10             = #90a9af
.Tones 11             = #9eb6bb
.Tones 12             = #aec2c5
.Tones 13             = #bdcdd0
.Tones 14             = #ccd8db
.Tones 15             = #dde2e3

# Raw filtered palette (no template variable — reference only)
# Filtered[0 ]         = #040404
`

	theme, err := ParseColorReference(sample)
	require.NoError(t, err)
	require.Equal(t, "192224", theme.Background)
	require.Equal(t, "dde2e3", theme.Foreground)
	require.Len(t, theme.Ansi, 16)
	require.Len(t, theme.Tones, 16)
	require.Nil(t, theme.TonesFromANSI)
	require.Equal(t, "6e6e6e", theme.Ansi[0])
	require.Equal(t, "192224", theme.Tones[0])
}

func TestParseColorReference_tonesFromANSI(t *testing.T) {
	var b strings.Builder
	b.WriteString(".Background = #111111\n.Foreground = #eeeeee\n")
	for i := 0; i < 16; i++ {
		_, _ = fmt.Fprintf(&b, ".Ansi %d = #%02x0000\n", i, i)
	}
	for i := 0; i < 16; i++ {
		_, _ = fmt.Fprintf(&b, ".Tones %d = #00%02x00\n", i, i)
	}
	for i := 0; i < 16; i++ {
		_, _ = fmt.Fprintf(&b, ".TonesFromANSI %d = #0000%02x\n", i, i)
	}
	theme, err := ParseColorReference(b.String())
	require.NoError(t, err)
	require.Len(t, theme.TonesFromANSI, 16)
	require.Equal(t, "000000", theme.TonesFromANSI[0])
}
