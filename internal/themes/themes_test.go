package themes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinByID(t *testing.T) {
	assert.NotNil(t, BuiltinByID("alacritty"), "BuiltinByID(alacritty) should be non-nil")
	assert.Nil(t, BuiltinByID("nonexistent"), "BuiltinByID(nonexistent) should be nil")
	assert.Equal(t, "rofi/color.rasi.tmpl", BuiltinByID("rofi").EmbedPath, "rofi EmbedPath = %s", BuiltinByID("rofi").EmbedPath)
}

func TestTemplateVariableHelp(t *testing.T) {
	s := TemplateVariableHelp()
	assert.NotEmpty(t, s, "TemplateVariableHelp should be non-empty")
	assert.True(t, len(BuiltinTemplates) >= 3, "BuiltinTemplates len = %d", len(BuiltinTemplates))
}

func TestApplyEmbedded(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.toml")
	data := ThemeData{
		Background: "1a1a1a",
		Foreground: "e0e0e0",
		Ansi:       make([]string, 16),
		Tones:      make([]string, 16),
	}
	for i := range data.Ansi {
		data.Ansi[i] = "808080"
	}
	for i := range data.Tones {
		data.Tones[i] = "404040"
	}
	err := ApplyEmbedded("alacritty.toml.tmpl", out, data)
	require.NoError(t, err)
	b, err := os.ReadFile(out)
	require.NoError(t, err)
	content := string(b)
	assert.NotEmpty(t, content, "output file empty")
	assert.False(t, strings.Contains(content, "# target:"), "written file should not contain # target line: %s", content)
	assert.True(t, strings.Contains(content, "1a1a1a") && strings.Contains(content, "e0e0e0"),
		"output should contain bg/fg: %s", content)
}

func TestParseTemplateHeaders(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cases := []struct {
		desc       string
		input      string
		wantTarget string
		wantName   string
		wantBody   string
	}{
		{
			desc:       "target only",
			input:      "# target: $HOME/.config/foo.conf\nbody\n",
			wantTarget: filepath.Join(home, ".config", "foo.conf"),
			wantBody:   "body\n",
		},
		{
			desc:       "target then name",
			input:      "# target: $HOME/.config/foo.conf\n# name: My Config\nbody\n",
			wantTarget: filepath.Join(home, ".config", "foo.conf"),
			wantName:   "My Config",
			wantBody:   "body\n",
		},
		{
			desc:       "name then target",
			input:      "# name: My Config\n# target: $HOME/.config/foo.conf\nbody\n",
			wantTarget: filepath.Join(home, ".config", "foo.conf"),
			wantName:   "My Config",
			wantBody:   "body\n",
		},
		{
			desc:     "name only (no target)",
			input:    "# name: No Target\nbody\n",
			wantName: "No Target",
			wantBody: "body\n",
		},
		{
			desc:     "no headers",
			input:    "body\n",
			wantBody: "body\n",
		},
		{
			desc:       "tilde expansion in target",
			input:      "# target: ~/.config/foo.conf\n# name: Foo\nbody\n",
			wantTarget: filepath.Join(home, ".config", "foo.conf"),
			wantName:   "Foo",
			wantBody:   "body\n",
		},
		{
			desc:       "unrecognised comment stops header block",
			input:      "# target: $HOME/.config/foo.conf\n# random comment\nbody\n",
			wantTarget: filepath.Join(home, ".config", "foo.conf"),
			wantBody:   "# random comment\nbody\n",
		},
		{
			desc:       "blank line between headers is skipped",
			input:      "# target: $HOME/.config/foo.conf\n\n# name: My Config\nbody\n",
			wantTarget: filepath.Join(home, ".config", "foo.conf"),
			wantName:   "My Config",
			wantBody:   "body\n",
		},
		{
			desc:     "empty input",
			input:    "",
			wantBody: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			h := ParseTemplateHeaders(tc.input)
			assert.Equal(t, tc.wantTarget, h.Target)
			assert.Equal(t, tc.wantName, h.Name)
			assert.Equal(t, tc.wantBody, h.Body)
		})
	}
}
