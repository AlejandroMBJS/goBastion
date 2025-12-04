// Package router provides a lightweight HTTP router with middleware support and path parameter extraction.
//
// ⚠️ FRAMEWORK CORE - DO NOT MODIFY (unless extending framework routing)
//
// This package implements the core HTTP routing functionality for goBastion:
//   - Route registration with HTTP methods (GET, POST, PUT, DELETE, etc.)
//   - Path parameter extraction (e.g., /users/{id})
//   - Global middleware chain
//   - Route matching and dispatching
//
// WHEN TO MODIFY:
//   - ❌ DO NOT modify route matching logic (breaks routing)
//   - ❌ DO NOT modify middleware chain execution (breaks middleware order)
//   - ❌ DO NOT change Handler or Middleware function signatures (breaking change)
//   - ⚠️  EXTEND with caution if you need advanced routing features
//   - ✅ REGISTER your application routes via Handle() (normal usage)
//   - ✅ CREATE middleware in your app code (internal/app/middleware/)
//
// ROUTER CONCEPTS:
//
// 1. Handler: Function that processes HTTP requests
//    - Signature: func(w http.ResponseWriter, r *http.Request, params map[string]string)
//    - params contains path parameters extracted from URL (e.g., {"id": "123"})
//
// 2. Middleware: Function that wraps a Handler to add functionality
//    - Signature: func(Handler) Handler
//    - Examples: logging, authentication, CSRF protection
//    - Executed in REVERSE order of registration (last registered runs first)
//
// 3. Route: Combination of HTTP method, URL pattern, and handler
//    - Supports path parameters with {name} syntax
//    - Example: GET /users/{id} matches GET /users/123
//
// USAGE IN APPLICATION CODE:
//
//	import "github.com/AlejandroMBJS/goBastion/internal/framework/router"
//
//	r := router.New()
//
//	// Register global middleware
//	r.Use(middleware.Logging)
//	r.Use(middleware.CSRF(cfg.Security))
//
//	// Register routes
//	r.Handle("GET", "/users", handleListUsers)
//	r.Handle("GET", "/users/{id}", handleGetUser)
//	r.Handle("POST", "/users", handleCreateUser)
//
//	// Start server
//	http.ListenAndServe(":8080", r)
//
// MIDDLEWARE EXECUTION ORDER:
// Middleware is applied in REVERSE order of registration:
//
//	r.Use(mw1)  // Runs THIRD
//	r.Use(mw2)  // Runs SECOND
//	r.Use(mw3)  // Runs FIRST
//
// This allows outer middleware to wrap inner middleware correctly.
//
// ROUTE PATTERNS:
//   - Static: /users, /admin/dashboard
//   - With params: /users/{id}, /posts/{postID}/comments/{commentID}
//   - Wildcard: Not yet supported (would need modification)
//
// PATH PARAMETER EXTRACTION:
//
//	r.Handle("GET", "/users/{id}", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	    userID := params["id"]  // Extract "id" from URL
//	    // ... handle request
//	})
//
// COMMON PATTERNS:
//
// 1. Protecting routes with authentication:
//
//	authMiddleware := middleware.RequireAuth()
//	r.Handle("GET", "/protected", authMiddleware(handleProtected))
//
// 2. Role-based access control:
//
//	adminOnly := middleware.RequireRole("admin")
//	r.Handle("GET", "/admin", adminOnly(handleAdmin))
//
// 3. Route groups (convention, not built-in):
//
//	func registerAdminRoutes(r *router.Router) {
//	    adminAuth := middleware.RequireRole("admin")
//	    r.Handle("GET", "/admin/users", adminAuth(handleUsers))
//	    r.Handle("GET", "/admin/settings", adminAuth(handleSettings))
//	}
//
// For examples, see: cmd/server/main.go (route registration)
// For middleware examples, see: internal/framework/middleware/middleware.go
package router

import (
	"net/http"
	"strings"
)

// Handler is the function signature for route handlers with path parameters.
//
// Parameters:
//   - w: HTTP response writer for sending responses
//   - r: HTTP request containing method, headers, body, etc.
//   - params: Map of path parameters extracted from URL pattern
//
// Example:
//
//	func handleUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	    userID := params["id"]  // From pattern /users/{id}
//	    fmt.Fprintf(w, "User ID: %s", userID)
//	}
type Handler func(http.ResponseWriter, *http.Request, map[string]string)

// Middleware wraps a Handler with additional functionality.
//
// Middleware can execute code before and after the wrapped handler:
//   - Before: authentication, validation, logging
//   - After: response modification, cleanup
//
// Example:
//
//	func LoggingMiddleware(next Handler) Handler {
//	    return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
//	        log.Printf("Request: %s %s", r.Method, r.URL.Path)
//	        next(w, r, params)  // Call the wrapped handler
//	        log.Printf("Response sent")
//	    }
//	}
type Middleware func(Handler) Handler

// Route represents a single API route
type Route struct {
	Method  string
	Pattern string
	Handler Handler
}

// Router is a minimal HTTP router built on net/http
type Router struct {
	routes      []Route
	middlewares []Middleware
}

// New creates a new Router instance
func New() *Router {
	return &Router{
		routes:      make([]Route, 0),
		middlewares: make([]Middleware, 0),
	}
}

// Use registers a global middleware
func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

// Handle registers a new route with the given method, pattern, and handler
func (r *Router) Handle(method, pattern string, h Handler) {
	// Apply all middlewares to the handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}

	r.routes = append(r.routes, Route{
		Method:  method,
		Pattern: pattern,
		Handler: h,
	})
}

// GET is a convenience method for registering GET routes
func (r *Router) GET(pattern string, h Handler) {
	r.Handle("GET", pattern, h)
}

// POST is a convenience method for registering POST routes
func (r *Router) POST(pattern string, h Handler) {
	r.Handle("POST", pattern, h)
}

// PUT is a convenience method for registering PUT routes
func (r *Router) PUT(pattern string, h Handler) {
	r.Handle("PUT", pattern, h)
}

// DELETE is a convenience method for registering DELETE routes
func (r *Router) DELETE(pattern string, h Handler) {
	r.Handle("DELETE", pattern, h)
}

// PATCH is a convenience method for registering PATCH routes
func (r *Router) PATCH(pattern string, h Handler) {
	r.Handle("PATCH", pattern, h)
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if route.Method != req.Method {
			continue
		}

		params, ok := match(route.Pattern, req.URL.Path)
		if ok {
			route.Handler(w, req, params)
			return
		}
	}

	http.NotFound(w, req)
}

// match checks if a pattern matches a path and extracts parameters
func match(pattern, path string) (map[string]string, bool) {
	// Trim trailing slashes for consistent matching
	pattern = strings.TrimSuffix(pattern, "/")
	path = strings.TrimSuffix(path, "/")

	// Handle exact root match
	if pattern == "" && path == "" {
		return make(map[string]string), true
	}

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Must have same number of parts
	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)

	for i := 0; i < len(patternParts); i++ {
		patternPart := patternParts[i]
		pathPart := pathParts[i]

		// Check if this is a parameter (enclosed in braces)
		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			// Extract parameter name
			paramName := strings.TrimPrefix(strings.TrimSuffix(patternPart, "}"), "{")
			params[paramName] = pathPart
		} else if patternPart != pathPart {
			// Static parts must match exactly
			return nil, false
		}
	}

	return params, true
}

// WrapStdHandler wraps a standard http.HandlerFunc into our Handler type
func WrapStdHandler(h http.HandlerFunc) Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		h(w, r)
	}
}

// WrapHandler wraps a standard http.Handler into our Handler type
func WrapHandler(h http.Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		h.ServeHTTP(w, r)
	}
}
