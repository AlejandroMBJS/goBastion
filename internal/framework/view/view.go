// Package view implements goBastion's custom template engine with clean, Go-like syntax.
//
// ⚠️ FRAMEWORK CORE - DO NOT MODIFY
//
// This package contains the core template preprocessing engine that powers goBastion's
// clean template syntax (go:: / @ constructs). It is security-critical code that:
//   - Preprocesses custom syntax into Go's html/template format
//   - Ensures automatic HTML escaping for XSS prevention
//   - Maintains type safety through Go's template compilation
//
// WHEN TO MODIFY:
//   - ❌ NEVER modify this file for application-level changes
//   - ❌ DO NOT bypass template preprocessing (security risk!)
//   - ❌ DO NOT add raw {{ }} syntax handling (breaks preprocessing)
//   - ✅ OK to extend template.FuncMap in your app code via AddFunctions()
//   - ⚠️  ONLY modify preprocessor if you're extending framework template syntax itself
//
// TEMPLATE SYNTAX SUPPORTED:
//   1. Echo expressions: @variable, @.Prop, @obj.Field, @func(arg)
//      - Automatically converted to {{ ... }} with HTML escaping
//      - First char must be [a-zA-Z_.], NOT $ or special chars
//      - Examples: @.Title, @user.Name, @len(items)
//
//   2. Logic blocks: go:: ... ::end
//      - Go template control flow with clean syntax
//      - Converted to {{ ... }} {{ end }}
//      - Examples: go:: if .User, go:: range .Items, go:: else
//
// SECURITY NOTES:
//   - All @expr outputs are HTML-escaped automatically (XSS prevention)
//   - Never use raw {{ }} in templates - always use @ or go:: syntax
//   - Template compilation errors indicate syntax issues - check template rules
//
// EXTENSION POINTS:
//   - Add custom template functions via Engine.AddFunctions()
//   - Register custom helpers in your app's init code
//   - See README.md "Extending the Framework" section for examples
//
// For template syntax documentation, see: TEMPLATE_SYNTAX.md
// For configuration options, see: config/config.json
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

// Engine is the template rendering engine that preprocesses goBastion's custom syntax
// (go:: / @ constructs) into Go's html/template format.
//
// The engine maintains a base directory for templates and a function map for custom
// template helpers. All templates are preprocessed at render time to convert clean
// syntax into secure, HTML-escaped Go templates.
type Engine struct {
	baseDir string
	funcs   template.FuncMap
	verbose bool // Enable verbose template debugging (controlled by config.logging.verbose)
}

// NewEngine creates a new template engine instance.
//
// This initializes the template engine with a base directory for template files and
// a default set of template helper functions (upper, lower, title).
//
// Parameters:
//   - baseDir: Absolute path to the directory containing template files (e.g., "/app/templates")
//
// Returns:
//   - *Engine: Configured template engine ready to render templates
//   - error: Returns error if baseDir doesn't exist
//
// USAGE:
//
//	views, err := view.NewEngine("templates")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// EXTENSION:
// To add custom template functions, use Engine.AddFunctions() after creation:
//
//	views.AddFunctions(template.FuncMap{
//	    "myHelper": func(s string) string { return strings.ToUpper(s) },
//	})
func NewEngine(baseDir string) (*Engine, error) {
	// Check if base directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("template directory does not exist: %s", baseDir)
	}

	// Initialize default template functions
	// These are available in all templates via @ syntax
	funcs := template.FuncMap{
		"upper": strings.ToUpper, // @upper(.Name) -> JOHN
		"lower": strings.ToLower, // @lower(.Name) -> john
		"title": strings.Title,   // @title(.Name) -> John
	}

	return &Engine{
		baseDir: baseDir,
		funcs:   funcs,
		verbose: false,
	}, nil
}

// SetVerbose enables or disables verbose template debugging output.
//
// When enabled, the engine will print detailed debug information including:
//   - Template parse errors with full preprocessed content
//   - Template execution errors
//   - Syntax transformation details
//
// This should be controlled by config.logging.verbose in production.
//
// USAGE:
//
//	views, _ := view.NewEngine("templates")
//	views.SetVerbose(cfg.Logging.Verbose)
func (e *Engine) SetVerbose(verbose bool) {
	e.verbose = verbose
}

// Render renders a template with the given data and writes the output to the HTTP response.
//
// This is the main entry point for template rendering. It:
//  1. Locates the template file by name (adds .html extension)
//  2. Reads the template content
//  3. Preprocesses custom syntax (go:: / @) into Go template syntax ({{ }})
//  4. Compiles the preprocessed template
//  5. Executes the template with the provided data
//  6. Writes the resulting HTML to the HTTP response
//
// Parameters:
//   - w: HTTP response writer where rendered HTML will be written
//   - name: Template name without extension (e.g., "admin/dashboard" for "templates/admin/dashboard.html")
//   - data: Data to pass to the template (typically map[string]any or a struct)
//
// Returns:
//   - error: Returns error if template not found, preprocessing fails, or compilation fails
//
// USAGE IN HANDLERS:
//
//	func handleDashboard(views *view.Engine) http.HandlerFunc {
//	    return func(w http.ResponseWriter, r *http.Request) {
//	        data := map[string]any{
//	            "Title": "Dashboard",
//	            "User": getCurrentUser(r),
//	        }
//	        if err := views.Render(w, "admin/dashboard", data); err != nil {
//	            http.Error(w, "Failed to render", 500)
//	        }
//	    }
//	}
//
// SECURITY:
//   - All @ expressions are automatically HTML-escaped (XSS prevention)
//   - Template compilation validates type safety
//   - Never bypass this method to render raw HTML
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
		if e.verbose {
			fmt.Printf("[VERBOSE] Template parse error for %s: %v\n", name, err)
			fmt.Printf("[VERBOSE] Processed content:\n%s\n", processed)
		}
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
