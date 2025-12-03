package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworkrouter "github.com/AlejandroMBJS/goBastion/internal/framework/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/view"
)

func TestHomeRoute(t *testing.T) {
	// Create temp directory for templates
	tmpDir, err := os.MkdirTemp("", "gobastion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple home template
	homeTemplate := `<!DOCTYPE html>
<html>
<head><title>Welcome to goBastion</title></head>
<body>
<h1>Welcome to goBastion</h1>
<p>Your opinionated Go backend framework</p>
</body>
</html>`

	templatePath := filepath.Join(tmpDir, "home.html")
	if err := os.WriteFile(templatePath, []byte(homeTemplate), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create view engine
	views, err := view.NewEngine(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create view engine: %v", err)
	}

	// Create router and register home route
	r := frameworkrouter.New()
	RegisterHomeRoutes(r, views)

	// Create test request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	body := rr.Body.String()
	if !strings.Contains(body, "Welcome to goBastion") {
		t.Errorf("handler returned unexpected body: got %v", body)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/html; charset=utf-8")
	}
}

func TestHomeRouteNotFound(t *testing.T) {
	// Create temp directory for templates
	tmpDir, err := os.MkdirTemp("", "gobastion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create view engine (without home template)
	views, err := view.NewEngine(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create view engine: %v", err)
	}

	// Create router and register home route
	r := frameworkrouter.New()
	RegisterHomeRoutes(r, views)

	// Test that other routes return 404
	req, err := http.NewRequest("GET", "/nonexistent", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code for nonexistent route: got %v want %v", status, http.StatusNotFound)
	}
}
