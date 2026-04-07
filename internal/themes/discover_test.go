package themes_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0mega24/gowall/internal/themes"
)

func TestDiscoverTemplates(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".config", "gowall", "templates")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	// Write a test template
	content := "# target: $HOME/.config/test/out.conf\nfoo = {{.Background}}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mytest.tmpl"), []byte(content), 0o644))

	templates, err := themes.DiscoverTemplates()
	require.NoError(t, err)

	found := false
	for _, tpl := range templates {
		if tpl.ID == "mytest" {
			found = true
			assert.Equal(t, filepath.Join(home, ".config", "test", "out.conf"), tpl.TargetPath)
		}
	}
	assert.True(t, found, "mytest template should be discovered")
}

func TestTemplateBodyForExecute(t *testing.T) {
	const withTarget = "# target: $HOME/.config/x.toml\n[k]\nv = 1\n"
	assert.Equal(t, "[k]\nv = 1\n", themes.TemplateBodyForExecute(withTarget))
	assert.Equal(t, "no leading target\n", themes.TemplateBodyForExecute("no leading target\n"))
}

func TestParseTargetComment(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cases := []struct {
		name       string
		input      string
		wantTarget string
		wantBody   string
		wantErr    bool
	}{
		{
			name:       "valid",
			input:      "# target: $HOME/.config/foo/bar.conf\nrest of template\n",
			wantTarget: filepath.Join(home, ".config", "foo", "bar.conf"),
			wantBody:   "rest of template\n",
		},
		{
			name:    "missing comment",
			input:   "no target comment\nfoo\n",
			wantErr: true,
		},
		{
			name:       "tilde expansion",
			input:      "# target: ~/.config/foo/bar.conf\nbody\n",
			wantTarget: filepath.Join(home, ".config", "foo", "bar.conf"),
			wantBody:   "body\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target, body, err := themes.ParseTargetComment(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantTarget, target)
			assert.Equal(t, tc.wantBody, body)
		})
	}
}
