package view

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Engine is the template rendering engine
type Engine struct {
	baseDir string
	funcs   template.FuncMap
}

// NewEngine creates a new template engine
func NewEngine(baseDir string) (*Engine, error) {
	// Check if base directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("template directory does not exist: %s", baseDir)
	}

	funcs := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
	}

	return &Engine{
		baseDir: baseDir,
		funcs:   funcs,
	}, nil
}

// Render renders a template with the given data
func (e *Engine) Render(w http.ResponseWriter, name string, data any) error {
	// Construct full path
	fullPath := filepath.Join(e.baseDir, name+".html")

	// Read template file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Preprocess PHP-like syntax to Go template syntax
	processed := e.preprocess(string(content))

	// Parse template
	tmpl, err := template.New(name).Funcs(e.funcs).Parse(processed)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute template
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// RenderString renders a template and returns the result as a string
func (e *Engine) RenderString(name string, data any) (string, error) {
	// Construct full path
	fullPath := filepath.Join(e.baseDir, name+".html")

	// Read template file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Preprocess PHP-like syntax to Go template syntax
	processed := e.preprocess(string(content))

	// Parse template
	tmpl, err := template.New(name).Funcs(e.funcs).Parse(processed)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template to string
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// preprocess converts goBastion-specific syntax to Go template syntax
// Supports two main constructs:
// 1. go:: ... ::end - Logic blocks (if, for, range, with, etc.)
// 2. @expr - Echo expressions (HTML-escaped output)
func (e *Engine) preprocess(content string) string {
	// First handle backward compatibility with PHP-style tags (deprecated)
	content = e.handleLegacyPHPTags(content)

	// Process @expr echo expressions FIRST
	// This ensures they work both inside and outside logic blocks
	// Matches: @identifier or @func(args) or @obj.field
	// This regex handles:
	// - @variable
	// - @.field (dot notation for current context)
	// - @object.field
	// - @object.method()
	// - @func(args)
	echoRegex := regexp.MustCompile(`@([a-zA-Z_.][a-zA-Z0-9_.]*(?:\([^)]*\))?)`)
	content = echoRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the expression (remove @)
		expr := match[1:]
		return "{{ " + expr + " }}"
	})

	// Process go:: logic blocks AFTER echo expressions
	// Matches: go:: <statement>
	// Example: go:: if user != nil {
	goBlockRegex := regexp.MustCompile(`(?m)^[ \t]*go::\s*(.+?)[ \t]*$`)
	content = goBlockRegex.ReplaceAllString(content, "{{ $1 }}")

	// Process ::end tags
	// Matches: ::end
	endBlockRegex := regexp.MustCompile(`(?m)^[ \t]*::end[ \t]*$`)
	content = endBlockRegex.ReplaceAllString(content, "{{ end }}")

	return content
}

// handleLegacyPHPTags provides backward compatibility for old PHP-style syntax
// This can be removed in a future version
func (e *Engine) handleLegacyPHPTags(content string) string {
	// Only process if PHP tags are detected
	if !strings.Contains(content, "<?") {
		return content
	}

	// Log deprecation warning (in production, use proper logging)
	// fmt.Println("Warning: PHP-style template tags are deprecated. Use go:: / @ syntax instead.")

	// Replace <?= expr ?> with {{ expr }}
	re1 := regexp.MustCompile(`<\?=\s*(.*?)\s*\?>`)
	content = re1.ReplaceAllString(content, "{{ $1 }}")

	// Replace <? if expr ?> with {{ if expr }}
	re2 := regexp.MustCompile(`<\?\s*if\s+(.*?)\s*\?>`)
	content = re2.ReplaceAllString(content, "{{ if $1 }}")

	// Replace <? range expr ?> with {{ range expr }}
	re3 := regexp.MustCompile(`<\?\s*range\s+(.*?)\s*\?>`)
	content = re3.ReplaceAllString(content, "{{ range $1 }}")

	// Replace <? else ?> with {{ else }}
	re4 := regexp.MustCompile(`<\?\s*else\s*\?>`)
	content = re4.ReplaceAllString(content, "{{ else }}")

	// Replace <? end ?> with {{ end }}
	re5 := regexp.MustCompile(`<\?\s*end\s*\?>`)
	content = re5.ReplaceAllString(content, "{{ end }}")

	// Replace <? with expr ?> with {{ with expr }}
	re6 := regexp.MustCompile(`<\?\s*with\s+(.*?)\s*\?>`)
	content = re6.ReplaceAllString(content, "{{ with $1 }}")

	// Handle comments: <? /* comment */ ?> -> {{/* comment */}}
	re7 := regexp.MustCompile(`<\?\s*/\*(.*?)\*/\s*\?>`)
	content = re7.ReplaceAllString(content, "{{/* $1 */}}")

	return content
}

// AddFunc adds a custom template function
func (e *Engine) AddFunc(name string, fn any) {
	e.funcs[name] = fn
}

// RenderError renders an error page
func (e *Engine) RenderError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error %d</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 600px;
            margin: 100px auto;
            padding: 20px;
            text-align: center;
        }
        h1 { color: #e74c3c; }
        p { color: #7f8c8d; }
    </style>
</head>
<body>
    <h1>Error %d</h1>
    <p>%s</p>
</body>
</html>
`, statusCode, statusCode, template.HTMLEscapeString(message))

	w.Write([]byte(html))
}

// MustRender renders a template and panics on error (for use in handlers with recovery)
func (e *Engine) MustRender(w http.ResponseWriter, name string, data any) {
	if err := e.Render(w, name, data); err != nil {
		panic(err)
	}
}
