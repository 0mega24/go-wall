package themes_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0mega24/gowall/internal/themes"
)

func TestEnsureTemplatesInstalled(t *testing.T) {
	// Use a temp dir as HOME
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".config", "gowall", "templates")

	// First call: should create and populate
	require.NoError(t, themes.EnsureTemplatesInstalled())
	_, err := os.Stat(dir)
	require.NoError(t, err, "templates dir should exist after first install")

	// Should have at least one .tmpl file
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries, "templates dir should contain files")

	// Second call: idempotent (no error, no changes)
	require.NoError(t, themes.EnsureTemplatesInstalled())
}
