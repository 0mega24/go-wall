package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	icolor "github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/palette"
	"github.com/0mega24/gowall/internal/pipeline"
	"github.com/0mega24/gowall/internal/themes"
	"github.com/0mega24/gowall/internal/wallpaper"
)

// Config holds CLI configuration (flags).
type Config struct {
	ImagePath           string // empty = detect from wallpaper sources
	ColorReferencePath  string // if set, load theme from this Gowall Color Reference file (skip image pipeline)
	Templates           string // comma-separated built-in IDs
	TemplateList        string // path to file listing templates
	RetoneANSI          bool   // use standard 16 ANSI retoned to wallpaper
	CustomANSIColors    string // path to file with 16 hex colors (one per line); implies retone with that set
	Quiet               bool
	Seed                int64 // if non-zero, deterministic k-means (e.g. for tests)
	Constraints         map[int]pipeline.SlotConstraint // per-slot constraints: pin, H/S/V lock, post-gen tweak
	BackgroundHex       string                          // optional: override background color (#rrggbb)
	Algorithm           string                          // clustering algorithm name (empty = kmeans++)
	GlobalAdjust        pipeline.GlobalAdjust           // global hue (deg) + S/V % scale before per-slot constraints
}

// Run runs the full flow: resolve image, run pipeline, print progress and swatches, apply templates.
func Run(cfg Config) error {
	if cfg.Templates == "" {
		cfg.Templates = "alacritty,polybar,rofi"
	}

	if cfg.ColorReferencePath != "" {
		return runFromColorReference(cfg)
	}

	var imagePath string
	var sourceName string
	if cfg.ImagePath != "" {
		imagePath = os.ExpandEnv(cfg.ImagePath)
		if _, err := os.Stat(imagePath); err != nil {
			return fmt.Errorf("image: %w", err)
		}
		if !cfg.Quiet {
			fmt.Println("Using image:", imagePath)
		}
	} else {
		path, name, err := wallpaper.FirstOf(wallpaper.DefaultSources()...)
		if err != nil {
			return fmt.Errorf("wallpaper: %w", err)
		}
		imagePath = path
		sourceName = name
		if !cfg.Quiet {
			fmt.Println("Wallpaper ("+name+"):", imagePath)
		}
	}

	img, err := wallpaper.LoadImage(imagePath)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	if !cfg.Quiet {
		fmt.Printf("Resolution: %d×%d\n", bounds.Dx(), bounds.Dy())
	}

	opts := pipeline.DefaultOptions()
	opts.RetoneANSI = cfg.RetoneANSI
	opts.Seed = cfg.Seed
	opts.Constraints = cfg.Constraints
	opts.BackgroundHex = cfg.BackgroundHex
	opts.GlobalAdjust = cfg.GlobalAdjust
	if cfg.Algorithm != "" {
		c, err := icolor.Get(cfg.Algorithm)
		if err != nil {
			return fmt.Errorf("algorithm: %w", err)
		}
		opts.Clusterer = c
	}
	if cfg.CustomANSIColors != "" {
		custom, err := palette.LoadANSIHexFile(cfg.CustomANSIColors)
		if err != nil {
			return fmt.Errorf("ansi-colors: %w", err)
		}
		opts.CustomANSI = custom
		opts.RetoneANSI = true
	}

	var progress pipeline.ProgressFunc
	if !cfg.Quiet {
		progress = func(stage, detail string) {
			fmt.Printf("  %s… %s\n", stage, detail)
		}
	}

	result, err := pipeline.Run(img, opts, progress)
	if err != nil {
		return err
	}

	if !cfg.Quiet {
		fmt.Println("\nFiltered palette")
		fmt.Println(RowOfSwatchHex(result.Filtered, 4, ""))
		fmt.Println("\nANSI colors")
		fmt.Println(RowOfSwatchHex(result.ANSI, 4, ""))
		fmt.Println("\nTones (bg → fg)")
		fmt.Println(RowOfSwatchHex(result.Tones, 4, ""))
	}

	if err := applyTemplatesFromResult(cfg, result); err != nil {
		return err
	}
	_ = sourceName
	return nil
}

func runFromColorReference(cfg Config) error {
	path := os.ExpandEnv(cfg.ColorReferencePath)
	theme, err := themes.LoadColorReferenceFile(path)
	if err != nil {
		return fmt.Errorf("color reference: %w", err)
	}
	result, err := pipeline.ResultFromThemeData(theme)
	if err != nil {
		return err
	}
	if !cfg.Quiet {
		fmt.Println("Using color reference:", path)
		fmt.Println("\nANSI colors")
		fmt.Println(RowOfSwatchHex(result.ANSI, 4, ""))
		fmt.Println("\nTones (bg → fg)")
		fmt.Println(RowOfSwatchHex(result.Tones, 4, ""))
	}
	return applyTemplatesFromResult(cfg, result)
}

func applyTemplatesFromResult(cfg Config, result pipeline.Result) error {
	toApply, err := resolveTemplates(cfg)
	if err != nil {
		return err
	}
	for _, item := range toApply {
		if item.builtin != nil {
			out := item.outPath
			if out == "" {
				out = os.ExpandEnv(item.builtin.DefaultOutput)
			}
			if err := themes.ApplyEmbedded(item.builtin.EmbedPath, out, result.Theme); err != nil {
				return fmt.Errorf("template %s: %w", item.builtin.ID, err)
			}
			fmt.Println("Wrote", out)
		} else {
			if err := themes.ApplyTemplate(item.tmplPath, item.outPath, result.Theme); err != nil {
				return fmt.Errorf("template %s: %w", item.tmplPath, err)
			}
			fmt.Println("Wrote", item.outPath)
		}
	}
	return nil
}

// ApplyTheme writes built-in templates for the given theme data to their default paths.
// Used by the TUI when applying from an already-computed result.
func ApplyTheme(theme themes.ThemeData, templateIDs []string) error {
	for _, id := range templateIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if bt := themes.BuiltinByID(id); bt != nil {
			out := os.ExpandEnv(bt.DefaultOutput)
			if err := themes.ApplyEmbedded(bt.EmbedPath, out, theme); err != nil {
				return fmt.Errorf("template %s: %w", bt.ID, err)
			}
			fmt.Println("Wrote", out)
		} else {
			if err := themes.Apply(id, theme); err != nil {
				return fmt.Errorf("template %s: %w", id, err)
			}
		}
	}
	return nil
}

// ApplyThemeFromList reads a template-list file and applies each entry to the given theme.
// Same format as -template-list: id, id<tab>out, or file<tab>tmpl<tab>out per line.
func ApplyThemeFromList(theme themes.ThemeData, listPath string) error {
	f, err := os.Open(listPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 && parts[0] == "file" {
			if err := themes.ApplyTemplate(parts[1], parts[2], theme); err != nil {
				return fmt.Errorf("template %s: %w", parts[1], err)
			}
			fmt.Println("Wrote", parts[2])
			continue
		}
		id := parts[0]
		out := ""
		if len(parts) >= 2 {
			out = strings.TrimSpace(parts[1])
		}
		bt := themes.BuiltinByID(id)
		if bt == nil {
			continue
		}
		if out == "" {
			out = os.ExpandEnv(bt.DefaultOutput)
		}
		if err := themes.ApplyEmbedded(bt.EmbedPath, out, theme); err != nil {
			return fmt.Errorf("template %s: %w", bt.ID, err)
		}
		fmt.Println("Wrote", out)
	}
	return sc.Err()
}

type applyItem struct {
	builtin  *themes.BuiltinTemplate
	tmplPath string
	outPath  string
}

func resolveTemplates(cfg Config) ([]applyItem, error) {
	var items []applyItem
	if cfg.TemplateList != "" {
		f, err := os.Open(cfg.TemplateList)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Split(line, "\t")
			if len(parts) >= 3 && parts[0] == "file" {
				items = append(items, applyItem{tmplPath: parts[1], outPath: parts[2]})
				continue
			}
			id := parts[0]
			out := ""
			if len(parts) >= 2 {
				out = strings.TrimSpace(parts[1])
			}
			bt := themes.BuiltinByID(id)
			if bt == nil {
				continue
			}
			if out == "" {
				out = os.ExpandEnv(bt.DefaultOutput)
			}
			items = append(items, applyItem{builtin: bt, outPath: out})
		}
		return items, sc.Err()
	}
	for _, id := range strings.Split(cfg.Templates, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		bt := themes.BuiltinByID(id)
		if bt == nil {
			continue
		}
		items = append(items, applyItem{builtin: bt, outPath: os.ExpandEnv(bt.DefaultOutput)})
	}
	return items, nil
}
