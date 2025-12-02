package router

import (
	"fmt"
	"net/http"

	"go-native-fastapi/internal/app/models"
	"go-native-fastapi/internal/framework/config"
	"go-native-fastapi/internal/framework/db"
	frameworkrouter "go-native-fastapi/internal/framework/router"
	"go-native-fastapi/internal/framework/security"
	"go-native-fastapi/internal/framework/view"

	"golang.org/x/crypto/bcrypt"
)

// RegisterAuthViewsRoutes registers HTML authentication routes
func RegisterAuthViewsRoutes(r *frameworkrouter.Router, cfg config.SecurityConfig, views *view.Engine) {
	// GET /login - show login page
	r.Handle("GET", "/login", handleLoginPage(cfg, views))

	// POST /login - process login form
	r.Handle("POST", "/login", handleLoginForm(cfg, views))

	// GET /register - show register page
	r.Handle("GET", "/register", handleRegisterPage(cfg, views))

	// POST /register - process register form
	r.Handle("POST", "/register", handleRegisterForm(cfg, views))
}

// handleLoginPage shows the login page
func handleLoginPage(cfg config.SecurityConfig, views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		data := map[string]any{
			"Error":     "",
			"CSRFToken": "",
		}

		// Generate CSRF token if CSRF is enabled
		if cfg.EnableCSRF {
			token, err := security.GenerateCSRFToken(cfg.JWTSecret)
			if err == nil {
				data["CSRFToken"] = token
				// Set CSRF cookie
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

		if err := views.Render(w, "auth/login", data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

// handleLoginForm processes the login form submission
func handleLoginForm(cfg config.SecurityConfig, views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		// Parse form
		if err := r.ParseForm(); err != nil {
			renderLoginError(w, views, cfg, "Invalid form data")
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Validate CSRF if enabled
		if cfg.EnableCSRF {
			csrfCookie, err := r.Cookie(cfg.CSRFCookieName)
			if err != nil || csrfCookie.Value == "" {
				renderLoginError(w, views, cfg, "CSRF token missing")
				return
			}
			csrfFormToken := r.FormValue("csrf_token")
			if !security.ValidateCSRFToken(cfg.JWTSecret, csrfCookie.Value, csrfFormToken) {
				renderLoginError(w, views, cfg, "CSRF token invalid")
				return
			}
		}

		// Validate input
		loginInput := models.LoginInput{
			Email:    email,
			Password: password,
		}
		if err := loginInput.Validate(); err != nil {
			renderLoginError(w, views, cfg, err.Error())
			return
		}

		// Get user by email
		user, passwordHash, err := db.GetUserByEmail(r.Context(), email)
		if err != nil {
			renderLoginError(w, views, cfg, "Invalid email or password")
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
			renderLoginError(w, views, cfg, "Invalid email or password")
			return
		}

		// Check if user is active
		if !user.IsActive {
			renderLoginError(w, views, cfg, "Account is inactive")
			return
		}

		// Generate JWT token
		accessToken, err := security.GenerateToken(
			cfg.JWTSecret,
			fmt.Sprintf("%d", user.ID),
			user.Role,
			cfg.AccessTokenMinutes,
		)
		if err != nil {
			renderLoginError(w, views, cfg, "Failed to generate token")
			return
		}

		// Set auth cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    accessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: http.SameSiteLaxMode,
			MaxAge:   cfg.AccessTokenMinutes * 60,
		})

		// Redirect based on role
		if user.Role == "admin" || user.IsStaff || user.IsSuperuser {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

// handleRegisterPage shows the registration page
func handleRegisterPage(cfg config.SecurityConfig, views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		data := map[string]any{
			"Error":     "",
			"Success":   "",
			"CSRFToken": "",
			"Name":      "",
			"Email":     "",
		}

		// Generate CSRF token if CSRF is enabled
		if cfg.EnableCSRF {
			token, err := security.GenerateCSRFToken(cfg.JWTSecret)
			if err == nil {
				data["CSRFToken"] = token
				// Set CSRF cookie
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

		if err := views.Render(w, "auth/register", data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

// handleRegisterForm processes the registration form submission
func handleRegisterForm(cfg config.SecurityConfig, views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		// Parse form
		if err := r.ParseForm(); err != nil {
			renderRegisterError(w, views, cfg, "Invalid form data", "", "")
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Validate CSRF if enabled
		if cfg.EnableCSRF {
			csrfCookie, err := r.Cookie(cfg.CSRFCookieName)
			if err != nil || csrfCookie.Value == "" {
				renderRegisterError(w, views, cfg, "CSRF token missing", name, email)
				return
			}
			csrfFormToken := r.FormValue("csrf_token")
			if !security.ValidateCSRFToken(cfg.JWTSecret, csrfCookie.Value, csrfFormToken) {
				renderRegisterError(w, views, cfg, "CSRF token invalid", name, email)
				return
			}
		}

		// Check password confirmation
		if password != confirmPassword {
			renderRegisterError(w, views, cfg, "Passwords do not match", name, email)
			return
		}

		// Validate input
		registerInput := models.RegisterInput{
			Name:     name,
			Email:    email,
			Password: password,
			Role:     "user", // Default role
		}
		if err := registerInput.Validate(); err != nil {
			renderRegisterError(w, views, cfg, err.Error(), name, email)
			return
		}

		// Check if user already exists
		_, _, err := db.GetUserByEmail(r.Context(), email)
		if err == nil {
			renderRegisterError(w, views, cfg, "User with this email already exists", name, email)
			return
		}

		// Hash password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			renderRegisterError(w, views, cfg, "Failed to hash password", name, email)
			return
		}

		// Create user
		user, err := db.CreateUser(r.Context(), registerInput, string(passwordHash))
		if err != nil {
			renderRegisterError(w, views, cfg, "Failed to create user", name, email)
			return
		}

		// Generate JWT token
		accessToken, err := security.GenerateToken(
			cfg.JWTSecret,
			fmt.Sprintf("%d", user.ID),
			user.Role,
			cfg.AccessTokenMinutes,
		)
		if err != nil {
			renderRegisterError(w, views, cfg, "Failed to generate token", name, email)
			return
		}

		// Set auth cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    accessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: http.SameSiteLaxMode,
			MaxAge:   cfg.AccessTokenMinutes * 60,
		})

		// Redirect to home or admin based on role
		if user.Role == "admin" || user.IsStaff || user.IsSuperuser {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

// Helper function to render login page with error
func renderLoginError(w http.ResponseWriter, views *view.Engine, cfg config.SecurityConfig, errorMsg string) {
	data := map[string]any{
		"Error":     errorMsg,
		"CSRFToken": "",
	}

	// Generate new CSRF token
	if cfg.EnableCSRF {
		token, err := security.GenerateCSRFToken(cfg.JWTSecret)
		if err == nil {
			data["CSRFToken"] = token
			http.SetCookie(w, &http.Cookie{
				Name:     cfg.CSRFCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   false,
				SameSite: http.SameSiteStrictMode,
			})
		}
	}

	w.WriteHeader(http.StatusBadRequest)
	if err := views.Render(w, "auth/login", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// Helper function to render register page with error
func renderRegisterError(w http.ResponseWriter, views *view.Engine, cfg config.SecurityConfig, errorMsg, name, email string) {
	data := map[string]any{
		"Error":     errorMsg,
		"Success":   "",
		"CSRFToken": "",
		"Name":      name,
		"Email":     email,
	}

	// Generate new CSRF token
	if cfg.EnableCSRF {
		token, err := security.GenerateCSRFToken(cfg.JWTSecret)
		if err == nil {
			data["CSRFToken"] = token
			http.SetCookie(w, &http.Cookie{
				Name:     cfg.CSRFCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   false,
				SameSite: http.SameSiteStrictMode,
			})
		}
	}

	w.WriteHeader(http.StatusBadRequest)
	if err := views.Render(w, "auth/register", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}
