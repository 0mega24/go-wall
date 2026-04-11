package cli

import (
	"fmt"
	"os"

	"github.com/0mega24/gowall/internal/themes"
)

// ListTemplates prints built-in template IDs, default paths, and variable reference.
func ListTemplates() {
	fmt.Println("Built-in template IDs and default output paths:")
	fmt.Println()
	for _, t := range themes.BuiltinTemplates {
		fmt.Printf("  %-12s  %s\n", t.ID, os.ExpandEnv(t.DefaultOutput))
	}
	fmt.Println()
	fmt.Println(themes.TemplateVariableHelp())
}
