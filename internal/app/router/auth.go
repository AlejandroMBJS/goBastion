package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AlejandroMBJS/goBastion/internal/app/models"
	"github.com/AlejandroMBJS/goBastion/internal/framework/config"
	"github.com/AlejandroMBJS/goBastion/internal/framework/db"
	"github.com/AlejandroMBJS/goBastion/internal/framework/middleware"
	frameworkrouter "github.com/AlejandroMBJS/goBastion/internal/framework/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/security"

	"golang.org/x/crypto/bcrypt"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r *frameworkrouter.Router, cfg config.SecurityConfig) {
	// POST /api/v1/auth/register
	r.Handle("POST", "/api/v1/auth/register", handleRegister(cfg))

	// POST /api/v1/auth/login
	r.Handle("POST", "/api/v1/auth/login", handleLogin(cfg))

	// POST /api/v1/auth/refresh
	r.Handle("POST", "/api/v1/auth/refresh", handleRefresh(cfg))

	// GET /api/v1/auth/me (requires authentication)
	r.Handle("GET", "/api/v1/auth/me", handleMe())
}

// handleRegister handles user registration
func handleRegister(cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		var input models.RegisterInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}

		// Validate input
		if err := input.Validate(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		// Check if user already exists
		_, _, err := db.GetUserByEmail(r.Context(), input.Email)
		if err == nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "User with this email already exists"})
			return
		}

		// Hash password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
			return
		}

		// Create user
		user, err := db.CreateUser(r.Context(), input, string(passwordHash))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
			return
		}

		// Generate tokens
		accessToken, refreshToken, err := security.CreateTokenPair(
			cfg.JWTSecret,
			fmt.Sprintf("%d", user.ID),
			user.Role,
			cfg.AccessTokenMinutes,
			cfg.RefreshTokenMinutes,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate tokens"})
			return
		}

		// Return response
		writeJSON(w, http.StatusCreated, map[string]any{
			"user":          user,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
		})
	}
}

// handleLogin handles user login
func handleLogin(cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		var input models.LoginInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}

		// Validate input
		if err := input.Validate(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		// Get user by email
		user, passwordHash, err := db.GetUserByEmail(r.Context(), input.Email)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
			return
		}

		// Check if user is active
		if !user.IsActive {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "User account is inactive"})
			return
		}

		// Generate tokens
		accessToken, refreshToken, err := security.CreateTokenPair(
			cfg.JWTSecret,
			fmt.Sprintf("%d", user.ID),
			user.Role,
			cfg.AccessTokenMinutes,
			cfg.RefreshTokenMinutes,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate tokens"})
			return
		}

		// Return response
		writeJSON(w, http.StatusOK, map[string]any{
			"user":          user,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
		})
	}
}

// handleRefresh handles token refresh
func handleRefresh(cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		var input models.RefreshInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}

		// Validate input
		if err := input.Validate(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		// Parse and validate refresh token
		claims, err := security.ParseAndValidateToken(cfg.JWTSecret, input.RefreshToken)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired refresh token"})
			return
		}

		// Generate new access token
		accessToken, err := security.GenerateToken(
			cfg.JWTSecret,
			claims.Sub,
			claims.Role,
			cfg.AccessTokenMinutes,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
			return
		}

		// Return new access token
		writeJSON(w, http.StatusOK, map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
		})
	}
}

// handleMe returns the current user's profile
func handleMe() frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		// Get claims from context
		claims := middleware.GetClaims(r.Context())
		if claims == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
			return
		}

		// Parse user ID from claims
		userID, err := strconv.ParseInt(claims.Sub, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Invalid user ID"})
			return
		}

		// Get user from database
		user, err := db.GetUser(r.Context(), int(userID))
		if err != nil {
			if err == db.ErrNotFound {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
			return
		}

		// Return user profile
		writeJSON(w, http.StatusOK, user)
	}
}

// writeJSON is a helper function to write JSON responses
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
