package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTemplates_ContainsBuiltins(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = old }()

	ListTemplates()
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	for _, id := range []string{"alacritty", "polybar", "rofi", "Built-in"} {
		assert.True(t, strings.Contains(out, id), "output should contain %q: %s", id, out)
	}
	assert.True(t, strings.Contains(out, ".Background"), "output should contain variable help (.Background)")
}
