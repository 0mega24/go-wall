package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTemplates_DefaultIDs(t *testing.T) {
	cfg := Config{Templates: "alacritty,polybar"}
	items, err := resolveTemplates(cfg)
	require.NoError(t, err)
	assert.Equal(t, 2, len(items))
	assert.True(t, items[0].builtin != nil && items[0].builtin.ID == "alacritty", "first item: %v", items[0])
	assert.True(t, items[1].builtin != nil && items[1].builtin.ID == "polybar", "second item: %v", items[1])
}

func TestResolveTemplates_EmptyAndComments(t *testing.T) {
	cfg := Config{Templates: "alacritty,  , rofi"}
	items, err := resolveTemplates(cfg)
	require.NoError(t, err)
	assert.Equal(t, 2, len(items))
}

func TestResolveTemplates_UnknownIDSkipped(t *testing.T) {
	cfg := Config{Templates: "alacritty,nonexistent,rofi"}
	items, err := resolveTemplates(cfg)
	require.NoError(t, err)
	assert.Equal(t, 2, len(items), "unknown ID should be skipped: len = %d", len(items))
}

func TestResolveTemplates_TemplateListFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "list.txt")
	err := os.WriteFile(f, []byte("alacritty\n# comment\npolybar\n"), 0o644)
	require.NoError(t, err)
	cfg := Config{TemplateList: f}
	items, err := resolveTemplates(cfg)
	require.NoError(t, err)
	assert.Equal(t, 2, len(items))
}

func TestResolveTemplates_TemplateListFile_CustomPath(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "list.txt")
	err := os.WriteFile(f, []byte("vscode\t/tmp/code-settings.json\n"), 0o644)
	require.NoError(t, err)
	cfg := Config{TemplateList: f}
	items, err := resolveTemplates(cfg)
	require.NoError(t, err)
	assert.True(t, len(items) == 1 && items[0].outPath == "/tmp/code-settings.json", "items = %v", items)
}

func TestResolveTemplates_TemplateListFile_FileEntry(t *testing.T) {
	dir := t.TempDir()
	listPath := filepath.Join(dir, "list.txt")
	tmplPath := filepath.Join(dir, "custom.tmpl")
	outPath := filepath.Join(dir, "out.css")
	err := os.WriteFile(tmplPath, []byte("{{ .Background }}"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(listPath, []byte("file\t"+tmplPath+"\t"+outPath+"\n"), 0o644)
	require.NoError(t, err)
	cfg := Config{TemplateList: listPath}
	items, err := resolveTemplates(cfg)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.True(t, items[0].builtin == nil && items[0].tmplPath == tmplPath && items[0].outPath == outPath, "items[0] = %+v", items[0])
}

func TestResolveTemplates_MissingListFile(t *testing.T) {
	cfg := Config{TemplateList: "/nonexistent/list.txt"}
	_, err := resolveTemplates(cfg)
	assert.Error(t, err, "expected error for missing file")
}
