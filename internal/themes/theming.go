package themes

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed *.tmpl polybar/*.tmpl rofi/*.tmpl
var embedTemplates embed.FS

type ThemeData struct {
	Background    string   // hex without # (e.g. "1a1a1a")
	Foreground    string   // hex
	Ansi          []string // 16 elements: [0]=black .. [7]=white, [8]=bright black .. [15]=bright white
	Tones         []string // 16 steps dark→light (from image palette)
	TonesFromANSI []string // 16 steps from retoned ANSI; only non-nil when -retone-ansi (or -ansi-colors) was used
}

// BuiltinTemplate defines a built-in template (embedded in the binary).
type BuiltinTemplate struct {
	ID            string // e.g. "alacritty"
	EmbedPath     string // path in embed FS
	DefaultOutput string // path with $HOME expanded (e.g. "$HOME/.config/alacritty/colors.toml")
}

// BuiltinTemplates is the registry of all embedded templates. Add new entries here and
// ensure the .tmpl file is included in the go:embed directive above.
var BuiltinTemplates = []BuiltinTemplate{
	{ID: "alacritty", EmbedPath: "alacritty.toml.tmpl", DefaultOutput: "$HOME/.config/alacritty/colors.toml"},
	{ID: "polybar", EmbedPath: "polybar/colors.ini.tmpl", DefaultOutput: "$HOME/.config/polybar/colors.ini"},
	{ID: "rofi", EmbedPath: "rofi/color.rasi.tmpl", DefaultOutput: "$HOME/.config/rofi/color.rasi"},
	{ID: "bspwmrc", EmbedPath: "bspwmrc.tmpl", DefaultOutput: "$HOME/.config/bspwm/bspwmrc"},
	{ID: "gtk", EmbedPath: "gtk.css.tmpl", DefaultOutput: "$HOME/.config/gtk-4.0/colors.css"},
	{ID: "vscode", EmbedPath: "vscode.json.tmpl", DefaultOutput: "$HOME/.config/Code/User/settings.json"},
}

// BuiltinByID returns the builtin template for id, or nil if not found.
func BuiltinByID(id string) *BuiltinTemplate {
	for i := range BuiltinTemplates {
		if BuiltinTemplates[i].ID == id {
			return &BuiltinTemplates[i]
		}
	}
	return nil
}

// Template describes a discovered template (either builtin or user-defined).
type Template struct {
	ID         string
	SourcePath string // filesystem path to the .tmpl file
	TargetPath string // expanded target path from # target: comment
	IsBuiltin  bool
}

// TemplateBodyForExecute returns content suitable for text/template.Parse.
// All leading "# target:" and "# name:" header lines are stripped so they are
// not written to generated config files.
func TemplateBodyForExecute(templateContent string) string {
	return ParseTemplateHeaders(templateContent).Body
}

// ParseTargetComment reads template headers and returns the target path and body.
// Returns an error if no "# target:" header is present.
func ParseTargetComment(templateContent string) (targetPath, body string, err error) {
	h := ParseTemplateHeaders(templateContent)
	if h.Target == "" {
		return "", "", fmt.Errorf("template missing # target: comment")
	}
	return h.Target, h.Body, nil
}

// TemplateHeaders holds metadata parsed from the leading comment block of a template file.
type TemplateHeaders struct {
	Target string // expanded path from "# target: <path>"; empty if absent
	Name   string // display name from "# name: <name>"; empty if absent
	Body   string // template content after the header block
}

// ParseTemplateHeaders scans leading lines of content for "# target: " and "# name: "
// comments in any order, stopping at the first line that does not match a known header.
// All matched header lines are stripped from Body.
func ParseTemplateHeaders(content string) TemplateHeaders {
	home := os.Getenv("HOME")
	var h TemplateHeaders
	rest := content
	for {
		newline := strings.Index(rest, "\n")
		var line, after string
		if newline == -1 {
			line = rest
			after = ""
		} else {
			line = rest[:newline]
			after = rest[newline+1:]
		}
		line = strings.TrimRight(line, "\r")

		if strings.HasPrefix(line, "# target: ") {
			raw := strings.TrimSpace(line[len("# target: "):])
			raw = strings.ReplaceAll(raw, "$HOME", home)
			// Only ~/path and bare ~ are expanded; ~username forms are left as-is.
			if strings.HasPrefix(raw, "~/") {
				raw = home + raw[1:]
			} else if raw == "~" {
				raw = home
			}
			h.Target = raw
			rest = after
			continue
		}
		if strings.HasPrefix(line, "# name: ") {
			h.Name = strings.TrimSpace(line[len("# name: "):])
			rest = after
			continue
		}
		if strings.TrimSpace(line) == "" && newline != -1 {
			rest = after
			continue
		}
		break
	}
	h.Body = rest
	return h
}

// DiscoverTemplates reads all *.tmpl files from $HOME/.config/gowall/templates/
// and returns a Template slice. Returns an empty slice (no error) if the directory
// does not exist.
func DiscoverTemplates() ([]Template, error) {
	home := os.Getenv("HOME")
	dir := filepath.Join(home, ".config", "gowall", "templates")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Build a set of builtin IDs for quick lookup.
	builtinIDs := make(map[string]bool, len(BuiltinTemplates))
	for _, bt := range BuiltinTemplates {
		builtinIDs[bt.ID] = true
	}

	var templates []Template
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".tmpl")
		srcPath := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", entry.Name(), err)
		}
		targetPath, _, err := ParseTargetComment(string(content))
		if err != nil {
			return nil, fmt.Errorf("template %s: %w", entry.Name(), err)
		}
		templates = append(templates, Template{
			ID:         id,
			SourcePath: srcPath,
			TargetPath: targetPath,
			IsBuiltin:  builtinIDs[id],
		})
	}
	return templates, nil
}

// Apply renders the template identified by id from $HOME/.config/gowall/templates/
// and writes the result to the target path declared in the template's # target: comment.
// ThemeData is required for rendering.
func Apply(id string, data ThemeData) error {
	templates, err := DiscoverTemplates()
	if err != nil {
		return fmt.Errorf("discovering templates: %w", err)
	}
	var found *Template
	for i := range templates {
		if templates[i].ID == id {
			found = &templates[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("template %q not found in $HOME/.config/gowall/templates/", id)
	}
	content, err := os.ReadFile(found.SourcePath)
	if err != nil {
		return fmt.Errorf("reading template %s: %w", found.SourcePath, err)
	}
	_, body, err := ParseTargetComment(string(content))
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", found.SourcePath, err)
	}
	tmpl, err := template.New(id).Parse(body)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", found.SourcePath, err)
	}
	if err := os.MkdirAll(filepath.Dir(found.TargetPath), 0o755); err != nil {
		return fmt.Errorf("creating output directory for %s: %w", found.TargetPath, err)
	}
	f, err := os.Create(found.TargetPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", found.TargetPath, err)
	}
	defer func() { _ = f.Close() }()
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template %s: %w", found.SourcePath, err)
	}
	return nil
}

// TemplateVariableHelp returns documentation for template authors (variables available in .tmpl files).
func TemplateVariableHelp() string {
	return `Template variables (use in .tmpl with Go template syntax, e.g. {{ .Background }}):

  .Background     string   Background color (hex without #)
  .Foreground     string   Foreground/text color
  .Ansi           []string 16 ANSI colors: index 0–7 normal (black,red,green,yellow,blue,magenta,cyan,white),
                           8–15 bright. Use: {{ index .Ansi 0 }} … {{ index .Ansi 15 }}
  .Tones          []string 16 tone steps dark→light (from image). Use: {{ index .Tones 0 }} … {{ index .Tones 15 }}
  .TonesFromANSI  []string 16 tone steps from retoned ANSI (only set when -retone-ansi). Use: {{ index .TonesFromANSI 0 }} … {{ index .TonesFromANSI 15 }}

Examples:
  #{{ .Background }}        →  #1a1a1a
  #{{ index .Ansi 1 }}      →  #ff0000 (red)
  {{ index .Tones 0 }}      →  darkest tone (palette ramp)
  {{ index .TonesFromANSI 0 }}  →  darkest tone from ANSI ramp (when retone used)
`
}

// ApplyTemplate renders a template from the filesystem (for tests or custom paths).
func ApplyTemplate(templatePath, outputPath string, data ThemeData) error {
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	body := TemplateBodyForExecute(string(raw))
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(body)
	if err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return tmpl.Execute(f, data)
}

// ApplyEmbedded renders a template embedded in the binary. name is the path
// inside the embed (e.g. "alacritty.toml.tmpl", "polybar/colors.ini.tmpl").
// Templates are embedded in the binary; no external template files are required.
func ApplyEmbedded(name, outputPath string, data ThemeData) error {
	raw, err := embedTemplates.ReadFile(name)
	if err != nil {
		return err
	}
	body := TemplateBodyForExecute(string(raw))
	tmpl, err := template.New(name).Parse(body)
	if err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return tmpl.Execute(f, data)
}

// RenderTemplate renders a template from a raw string (e.g. a user-defined template
// loaded from disk) to a string without writing to disk.
func RenderTemplate(raw string, data ThemeData) (string, error) {
	body := TemplateBodyForExecute(raw)
	tmpl, err := template.New("user").Parse(body)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderEmbedded renders an embedded template to a string without writing to disk.
func RenderEmbedded(name string, data ThemeData) (string, error) {
	raw, err := embedTemplates.ReadFile(name)
	if err != nil {
		return "", err
	}
	body := TemplateBodyForExecute(string(raw))
	tmpl, err := template.New(name).Parse(body)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
