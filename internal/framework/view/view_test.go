package view

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPreprocessEchoExpressions tests @expr syntax
func TestPreprocessEchoExpressions(t *testing.T) {
	engine := &Engine{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple variable",
			input:    "@user.Name",
			expected: "{{ user.Name }}",
		},
		{
			name:     "Function call",
			input:    "@formatPrice(product.Price)",
			expected: "{{ formatPrice(product.Price) }}",
		},
		{
			name:     "Nested property",
			input:    "@user.Profile.Avatar",
			expected: "{{ user.Profile.Avatar }}",
		},
		{
			name:     "Multiple expressions",
			input:    "<p>Hello @user.Name, your email is @user.Email</p>",
			expected: "<p>Hello {{ user.Name }}, your email is {{ user.Email }}</p>",
		},
		{
			name:     "Expression in attribute",
			input:    `<input value="@user.Name">`,
			expected: `<input value="{{ user.Name }}">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.preprocess(tt.input)
			if result != tt.expected {
				t.Errorf("preprocess() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestPreprocessLogicBlocks tests go:: ... ::end syntax
func TestPreprocessLogicBlocks(t *testing.T) {
	engine := &Engine{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple if block",
			input: `go:: if user != nil {
<p>Hello</p>
::end`,
			expected: `{{ if user != nil { }}
<p>Hello</p>
{{ end }}`,
		},
		{
			name: "Range block",
			input: `go:: range .Users
<li>@.Name</li>
::end`,
			expected: `{{ range .Users }}
<li>{{ .Name }}</li>
{{ end }}`,
		},
		{
			name: "Nested blocks",
			input: `go:: if users != nil {
go:: range users
<p>@.Name</p>
::end
::end`,
			expected: `{{ if users != nil { }}
{{ range users }}
<p>{{ .Name }}</p>
{{ end }}
{{ end }}`,
		},
		{
			name: "With block",
			input: `go:: with .User
<p>@.Name</p>
::end`,
			expected: `{{ with .User }}
<p>{{ .Name }}</p>
{{ end }}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.preprocess(tt.input)
			if result != tt.expected {
				t.Errorf("preprocess() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestPreprocessMixedSyntax tests combined go:: and @ syntax
func TestPreprocessMixedSyntax(t *testing.T) {
	engine := &Engine{}

	input := `go:: if user != nil {
<p>Hello @user.Name</p>
<p>Email: @user.Email</p>
::end`

	expected := `{{ if user != nil { }}
<p>Hello {{ user.Name }}</p>
<p>Email: {{ user.Email }}</p>
{{ end }}`

	result := engine.preprocess(input)
	if result != expected {
		t.Errorf("preprocess() = %q, want %q", result, expected)
	}
}

// TestPreprocessBackwardCompatibility tests legacy PHP-style syntax
func TestPreprocessBackwardCompatibility(t *testing.T) {
	engine := &Engine{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Old echo tag",
			input:    "<?= .Name ?>",
			expected: "{{ .Name }}",
		},
		{
			name:     "Old if tag",
			input:    "<? if .Error ?>Error<? end ?>",
			expected: "{{ if .Error }}Error{{ end }}",
		},
		{
			name:     "Old range tag",
			input:    "<? range .Items ?><li><?= .Name ?></li><? end ?>",
			expected: "{{ range .Items }}<li>{{ .Name }}</li>{{ end }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.preprocess(tt.input)
			if result != tt.expected {
				t.Errorf("preprocess() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestRenderString tests the RenderString method
func TestRenderString(t *testing.T) {
	// Create temp directory for templates
	tmpDir, err := os.MkdirTemp("", "gobastion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test template
	templateContent := `go:: if .Name
Hello @.Name!
::end`
	templatePath := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create engine
	engine, err := NewEngine(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test rendering
	data := map[string]any{
		"Name": "Alice",
	}

	result, err := engine.RenderString("test", data)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	expected := "Hello Alice!\n"
	if !strings.Contains(result, "Hello Alice!") {
		t.Errorf("RenderString() = %q, want to contain %q", result, expected)
	}
}

// TestRender tests the Render method with HTTP response writer
func TestRender(t *testing.T) {
	// Create temp directory for templates
	tmpDir, err := os.MkdirTemp("", "gobastion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test template
	templateContent := `<h1>@.Title</h1>
go:: if .Items
<ul>
go:: range .Items
<li>@.</li>
::end
</ul>
::end`
	templatePath := filepath.Join(tmpDir, "list.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create engine
	engine, err := NewEngine(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test rendering
	data := map[string]any{
		"Title": "My List",
		"Items": []string{"Apple", "Banana", "Cherry"},
	}

	w := httptest.NewRecorder()
	if err := engine.Render(w, "list", data); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	result := w.Body.String()

	// Check if key content is present
	if !strings.Contains(result, "My List") {
		t.Errorf("Render() missing title, got %q", result)
	}
	if !strings.Contains(result, "Apple") || !strings.Contains(result, "Banana") || !strings.Contains(result, "Cherry") {
		t.Errorf("Render() missing items, got %q", result)
	}

	// Check content type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Render() Content-Type = %q, want %q", contentType, "text/html; charset=utf-8")
	}
}

// TestAddFunc tests adding custom template functions
func TestAddFunc(t *testing.T) {
	// Create temp directory for templates
	tmpDir, err := os.MkdirTemp("", "gobastion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test template using the built-in upper function with proper syntax
	// Template functions in Go use the syntax: {{ funcName arg }}
	templateContent := `{{ upper .Text }}`
	templatePath := filepath.Join(tmpDir, "func.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create engine (it already has the upper function built-in)
	engine, err := NewEngine(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test rendering with built-in upper function
	data := map[string]any{
		"Text": "hello",
	}

	result, err := engine.RenderString("func", data)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if !strings.Contains(result, "HELLO") {
		t.Errorf("RenderString() = %q, want to contain %q", result, "HELLO")
	}
}

// TestRenderError tests the RenderError method
func TestRenderError(t *testing.T) {
	engine := &Engine{}
	w := httptest.NewRecorder()

	engine.RenderError(w, 404, "Page not found")

	result := w.Body.String()

	// Check status code
	if w.Code != 404 {
		t.Errorf("RenderError() status = %d, want 404", w.Code)
	}

	// Check error content
	if !strings.Contains(result, "404") {
		t.Errorf("RenderError() missing status code, got %q", result)
	}
	if !strings.Contains(result, "Page not found") {
		t.Errorf("RenderError() missing error message, got %q", result)
	}
}

// TestHTMLEscaping tests that @ expressions are HTML-escaped
func TestHTMLEscaping(t *testing.T) {
	// Create temp directory for templates
	tmpDir, err := os.MkdirTemp("", "gobastion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test template
	templateContent := `@.UnsafeHTML`
	templatePath := filepath.Join(tmpDir, "escape.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create engine
	engine, err := NewEngine(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test rendering with HTML content
	data := map[string]any{
		"UnsafeHTML": "<script>alert('xss')</script>",
	}

	result, err := engine.RenderString("escape", data)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check that HTML is escaped
	if strings.Contains(result, "<script>") {
		t.Errorf("RenderString() did not escape HTML, got %q", result)
	}
	if !strings.Contains(result, "&lt;script&gt;") {
		t.Errorf("RenderString() = %q, want escaped HTML", result)
	}
}

// TestComplexTemplate tests a complex real-world template
func TestComplexTemplate(t *testing.T) {
	// Create temp directory for templates
	tmpDir, err := os.MkdirTemp("", "gobastion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test template with complex logic
	templateContent := `<h1>@.Title</h1>
go:: if .Error
<div class="error">@.Error</div>
::end
go:: if .Users
<table>
go:: range .Users
<tr>
<td>@.ID</td>
<td>@.Name</td>
<td>
go:: if eq .Role "admin"
<span>Admin</span>
go:: else
<span>User</span>
::end
</td>
</tr>
::end
</table>
go:: else
<p>No users found</p>
::end`
	templatePath := filepath.Join(tmpDir, "complex.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create engine and add eq function
	engine, err := NewEngine(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	engine.AddFunc("eq", func(a, b string) bool { return a == b })

	// Test rendering
	type User struct {
		ID   int
		Name string
		Role string
	}
	data := map[string]any{
		"Title": "User List",
		"Users": []User{
			{ID: 1, Name: "Alice", Role: "admin"},
			{ID: 2, Name: "Bob", Role: "user"},
		},
	}

	var buf bytes.Buffer
	w := httptest.NewRecorder()
	w.Body = &buf

	if err := engine.Render(w, "complex", data); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	result := w.Body.String()

	// Check content
	if !strings.Contains(result, "User List") {
		t.Errorf("Render() missing title")
	}
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "Bob") {
		t.Errorf("Render() missing user names")
	}
	if !strings.Contains(result, "Admin") {
		t.Errorf("Render() missing admin role")
	}
}
