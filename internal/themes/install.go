package themes

import (
	"io/fs"
	"os"
	"path/filepath"
)

// EnsureTemplatesInstalled copies bundled templates to $HOME/.config/gowall/templates/
// if that directory does not already exist. It is idempotent: if the dir exists, no-op.
func EnsureTemplatesInstalled() error {
	home := os.Getenv("HOME")
	dir := filepath.Join(home, ".config", "gowall", "templates")
	if _, err := os.Stat(filepath.Clean(dir)); err == nil {
		return nil // already installed
	}
	if err := os.MkdirAll(filepath.Clean(dir), 0o750); err != nil {
		return err
	}
	return fs.WalkDir(embedTemplates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := embedTemplates.ReadFile(path)
		if err != nil {
			return err
		}
		dst := filepath.Clean(filepath.Join(dir, path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
			return err
		}
		return os.WriteFile(filepath.Clean(dst), content, 0o600)
	})
}
