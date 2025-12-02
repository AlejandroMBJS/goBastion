package router

import (
	"net/http"
	"strings"
)

// Handler is the function signature for route handlers with path parameters
type Handler func(http.ResponseWriter, *http.Request, map[string]string)

// Middleware wraps a Handler with additional functionality
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
