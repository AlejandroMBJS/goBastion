package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AlejandroMBJS/goBastion/internal/framework/config"
	"github.com/AlejandroMBJS/goBastion/internal/framework/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/security"
)

// Context keys for storing values
type contextKey string

const (
	requestIDKey contextKey = "requestID"
	claimsKey    contextKey = "claims"
)

// 1. RequestID generates and attaches a unique request ID
func RequestID() router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			// Generate request ID
			id := generateRequestID()

			// Add to context
			ctx := context.WithValue(r.Context(), requestIDKey, id)
			r = r.WithContext(ctx)

			// Add to response header
			w.Header().Set("X-Request-ID", id)

			next(w, r, params)
		}
	}
}

// 2. Logging logs HTTP requests with method, path, status, and duration
func Logging(next router.Handler) router.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapped, r, params)

		duration := time.Since(start)
		requestID := GetRequestID(r.Context())

		log.Printf("[%s] %s %s - %d - %v\n",
			requestID,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	}
}

// 3. Recover recovers from panics and returns 500
func Recover(next router.Handler) router.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())
				log.Printf("[%s] PANIC: %v\n", requestID, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":"Internal server error","request_id":"%s"}`, requestID)
			}
		}()

		next(w, r, params)
	}
}

// 4. WithTimeout adds a timeout to the request context
func WithTimeout(timeout time.Duration) router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			next(w, r, params)
		}
	}
}

// 5. CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(allowedOrigins []string) router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(w, r, params)
		}
	}
}

// 6. CSRFMiddleware implements CSRF protection using double-submit cookie
func CSRFMiddleware(cfg config.SecurityConfig) router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			if !cfg.EnableCSRF {
				next(w, r, params)
				return
			}

			// Skip CSRF for auth endpoints and API routes (JWT provides protection)
			if shouldSkipAuth(r.URL.Path) || strings.HasPrefix(r.URL.Path, "/api/") {
				next(w, r, params)
				return
			}

			method := r.Method

			// Safe methods don't need CSRF validation
			if method == "GET" || method == "HEAD" || method == "OPTIONS" {
				// Generate token if cookie is missing
				cookie, err := r.Cookie(cfg.CSRFCookieName)
				if err != nil || cookie.Value == "" {
					token, err := security.GenerateCSRFToken(cfg.JWTSecret)
					if err == nil {
						http.SetCookie(w, &http.Cookie{
							Name:     cfg.CSRFCookieName,
							Value:    token,
							Path:     "/",
							HttpOnly: true,
							Secure:   false, // Set to true in production with HTTPS
							SameSite: http.SameSiteStrictMode,
						})
					}
				}
				next(w, r, params)
				return
			}

			// Mutating methods need validation
			cookie, err := r.Cookie(cfg.CSRFCookieName)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"CSRF token missing"}`))
				return
			}

			headerToken := r.Header.Get(cfg.CSRFHeaderName)
			if !security.ValidateCSRFToken(cfg.JWTSecret, cookie.Value, headerToken) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"CSRF token invalid"}`))
				return
			}

			next(w, r, params)
		}
	}
}

// 7. JWTAuthMiddleware validates JWT tokens from Authorization header or auth_token cookie
func JWTAuthMiddleware(cfg config.SecurityConfig) router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			if !cfg.EnableJWT {
				next(w, r, params)
				return
			}

			// Skip auth for certain paths
			if shouldSkipAuth(r.URL.Path) {
				next(w, r, params)
				return
			}

			var token string
			var err error

			// First, try to get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				token, err = security.ExtractBearerToken(authHeader)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error":"Invalid authorization header"}`))
					return
				}
			} else {
				// If no header, try to get token from cookie
				cookie, err := r.Cookie("auth_token")
				if err != nil || cookie.Value == "" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error":"Missing authorization header or cookie"}`))
					return
				}
				token = cookie.Value
			}

			// Validate the token
			claims, err := security.ParseAndValidateToken(cfg.JWTSecret, token)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Invalid or expired token"}`))
				return
			}

			// Store claims in context
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			r = r.WithContext(ctx)

			next(w, r, params)
		}
	}
}

// 8. RequireRole checks if the user has the required role
func RequireRole(role string) router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			claims := GetClaims(r.Context())
			if claims == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Authentication required"}`))
				return
			}

			// Check role
			if claims.Role != role && claims.Role != "admin" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"Insufficient permissions"}`))
				return
			}

			next(w, r, params)
		}
	}
}

// 9. SecurityHeaders sets security-related HTTP headers
func SecurityHeaders() router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net")

			next(w, r, params)
		}
	}
}

// 10. MaxBodySize limits the request body size
func MaxBodySize(limitBytes int64) router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			r.Body = http.MaxBytesReader(w, r.Body, limitBytes)
			next(w, r, params)
		}
	}
}

// 11. RateLimit implements a simple rate limiter
func RateLimit(cfg config.RateLimitConfig) router.Middleware {
	limiter := newRateLimiter(cfg.RequestsPerMinute)

	return func(next router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			if !cfg.Enabled {
				next(w, r, params)
				return
			}

			ip := getIP(r)
			if !limiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Too many requests"}`))
				return
			}

			next(w, r, params)
		}
	}
}

// Helper functions

func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

func GetClaims(ctx context.Context) *security.Claims {
	if claims, ok := ctx.Value(claimsKey).(security.Claims); ok {
		return &claims
	}
	return nil
}

func shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/docs",
		"/docs/openapi.json",
		"/login",
		"/register",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func getIP(r *http.Request) string {
	// Try X-Forwarded-For first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Try X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Simple rate limiter implementation
type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type visitor struct {
	requests []time.Time
	mu       sync.Mutex
}

func newRateLimiter(requestsPerMinute int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}

	// Cleanup old visitors every 5 minutes
	go rl.cleanup()

	return rl
}

func (rl *rateLimiter) Allow(ip string) bool {
	rl.mu.RLock()
	v, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if !exists {
		v = &visitor{requests: make([]time.Time, 0)}
		rl.mu.Lock()
		rl.visitors[ip] = v
		rl.mu.Unlock()
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old requests
	valid := 0
	for i, t := range v.requests {
		if t.After(cutoff) {
			valid = i
			break
		}
	}
	v.requests = v.requests[valid:]

	// Check limit
	if len(v.requests) >= rl.limit {
		return false
	}

	v.requests = append(v.requests, now)
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if len(v.requests) == 0 || time.Since(v.requests[len(v.requests)-1]) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
