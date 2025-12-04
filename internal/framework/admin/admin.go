package admin

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
	"github.com/AlejandroMBJS/goBastion/internal/framework/view"

	"golang.org/x/crypto/bcrypt"
)

// Store full config for dashboard metrics
var fullConfig *config.Config

// RegisterRoutes registers admin routes with CSRF protection
func RegisterRoutes(r *frameworkrouter.Router, views *view.Engine, cfg config.SecurityConfig) {
	// Admin routes require authentication and admin role
	adminAuth := middleware.RequireRole("admin")

	// GET /admin - Dashboard
	r.Handle("GET", "/admin", adminAuth(handleDashboard(views, cfg)))

	// GET /admin/users - List users
	r.Handle("GET", "/admin/users", adminAuth(handleUsersList(views, cfg)))

	// GET /admin/users/new - Create new user form
	r.Handle("GET", "/admin/users/new", adminAuth(handleUserNew(views, cfg)))

	// POST /admin/users/new - Create new user
	r.Handle("POST", "/admin/users/new", adminAuth(handleUserCreate(views, cfg)))

	// GET /admin/users/{id} - View/edit user
	r.Handle("GET", "/admin/users/{id}", adminAuth(handleUserDetail(views, cfg)))

	// POST /admin/users/{id} - Update user
	r.Handle("POST", "/admin/users/{id}", adminAuth(handleUserUpdate(views, cfg)))

	// POST /admin/users/{id}/delete - Delete user
	r.Handle("POST", "/admin/users/{id}/delete", adminAuth(handleUserDelete(views, cfg)))
}

// SetFullConfig stores the full configuration for metrics display
func SetFullConfig(cfg *config.Config) {
	fullConfig = cfg
}

// handleDashboard renders the admin dashboard
func handleDashboard(views *view.Engine, cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		claims := middleware.GetClaims(r.Context())
		if claims == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Compute metrics
		totalUsers, _ := db.CountWhere(r.Context(), "users", map[string]any{})
		adminUsers, _ := db.CountWhere(r.Context(), "users", map[string]any{"role": "admin"})
		activeUsers, _ := db.CountWhere(r.Context(), "users", map[string]any{"is_active": true})

		// Get environment from config
		environment := "production"
		if fullConfig != nil {
			environment = fullConfig.App.Environment
		}

		// Build metrics data
		metrics := map[string]any{
			"TotalUsers":   totalUsers,
			"AdminUsers":   adminUsers,
			"ActiveUsers":  activeUsers,
			"RegularUsers": totalUsers - adminUsers,
		}

		// Build system config data
		systemConfig := map[string]any{
			"CSRFEnabled":     cfg.EnableCSRF,
			"JWTEnabled":      cfg.EnableJWT,
			"Environment":     environment,
			"DatabaseDriver":  "SQLite", // Default
			"RateLimiting":    false,
			"RequestsPerMin":  0,
		}

		if fullConfig != nil {
			systemConfig["DatabaseDriver"] = fullConfig.Database.Driver
			systemConfig["RateLimiting"] = fullConfig.RateLimit.Enabled
			systemConfig["RequestsPerMin"] = fullConfig.RateLimit.RequestsPerMinute
		}

		data := map[string]any{
			"Title":        "Admin Dashboard",
			"UserName":     claims.Sub,
			"CSRFToken":    generateAndSetCSRFToken(w, cfg),
			"Metrics":      metrics,
			"SystemConfig": systemConfig,
		}

		if err := views.Render(w, "admin/dashboard", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// UserRow wraps a user with CSRF token for template rendering
type UserRow struct {
	ID          int64
	Email       string
	Name        string
	Role        string
	IsActive    bool
	IsStaff     bool
	IsSuperuser bool
	CSRFToken   string
}

// handleUsersList renders the users list page
func handleUsersList(views *view.Engine, cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		users, err := db.ListUsers(r.Context())
		if err != nil {
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}

		csrfToken := generateAndSetCSRFToken(w, cfg)

		// Convert users to UserRow with CSRF token
		userRows := make([]UserRow, len(users))
		for i, user := range users {
			userRows[i] = UserRow{
				ID:          user.ID,
				Email:       user.Email,
				Name:        user.Name,
				Role:        user.Role,
				IsActive:    user.IsActive,
				IsStaff:     user.IsStaff,
				IsSuperuser: user.IsSuperuser,
				CSRFToken:   csrfToken,
			}
		}

		data := map[string]any{
			"Title":     "Users",
			"Users":     userRows,
			"CSRFToken": csrfToken,
		}

		if err := views.Render(w, "admin/users_list", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleUserDetail renders the user detail/edit page
func handleUserDetail(views *view.Engine, cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		user, err := db.GetUser(r.Context(), id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to load user", http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"Title":     fmt.Sprintf("Edit User: %s", user.Name),
			"User":      user,
			"CSRFToken": generateAndSetCSRFToken(w, cfg),
		}

		if err := views.Render(w, "admin/user_detail", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleUserUpdate processes the user update form
func handleUserUpdate(views *view.Engine, cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Check if it's a JSON request (from API)
		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" {
			handleUserUpdateJSON(w, r, id)
			return
		}

		// Validate CSRF token for form submissions
		if !validateCSRFFromForm(r, cfg) {
			http.Error(w, "CSRF token invalid", http.StatusForbidden)
			return
		}

		// Handle form submission
		name := r.FormValue("name")
		email := r.FormValue("email")
		role := r.FormValue("role")
		isStaff := r.FormValue("is_staff") == "on"
		isSuperuser := r.FormValue("is_superuser") == "on"
		isActive := r.FormValue("is_active") == "on"

		// Update basic user info
		input := models.UserInput{
			Name:  name,
			Email: email,
			Role:  role,
		}

		if err := input.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = db.UpdateUser(r.Context(), id, input)
		if err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}

		// Update admin fields
		if err := db.UpdateUserAdmin(r.Context(), id, isStaff, isSuperuser); err != nil {
			http.Error(w, "Failed to update admin fields", http.StatusInternalServerError)
			return
		}

		// Update active status
		if err := db.UpdateUserActive(r.Context(), id, isActive); err != nil {
			http.Error(w, "Failed to update active status", http.StatusInternalServerError)
			return
		}

		// Redirect back to users list
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// handleUserUpdateJSON handles JSON API updates
func handleUserUpdateJSON(w http.ResponseWriter, r *http.Request, id int) {
	var input struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Role        string `json:"role"`
		IsStaff     bool   `json:"is_staff"`
		IsSuperuser bool   `json:"is_superuser"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	// Update basic user info
	userInput := models.UserInput{
		Name:  input.Name,
		Email: input.Email,
		Role:  input.Role,
	}

	if err := userInput.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user, err := db.UpdateUser(r.Context(), id, userInput)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
		return
	}

	// Update admin fields
	if err := db.UpdateUserAdmin(r.Context(), id, input.IsStaff, input.IsSuperuser); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update admin fields"})
		return
	}

	// Get updated user
	user, err = db.GetUser(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get updated user"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// handleUserNew shows the create user form
func handleUserNew(views *view.Engine, cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		data := map[string]any{
			"Title":     "Create New User",
			"Error":     "",
			"CSRFToken": generateAndSetCSRFToken(w, cfg),
		}

		if err := views.Render(w, "admin/user_new", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleUserCreate creates a new user
func handleUserCreate(views *view.Engine, cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Validate CSRF token
		if !validateCSRFFromForm(r, cfg) {
			http.Error(w, "CSRF token invalid", http.StatusForbidden)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		role := r.FormValue("role")
		isStaff := r.FormValue("is_staff") == "on"
		isSuperuser := r.FormValue("is_superuser") == "on"
		isActive := r.FormValue("is_active") == "on"

		// Validate input
		input := models.RegisterInput{
			Name:     name,
			Email:    email,
			Password: password,
			Role:     role,
		}

		if err := input.Validate(); err != nil {
			renderUserNewError(w, views, cfg, err.Error(), name, email, role)
			return
		}

		// Check if user already exists
		_, _, err := db.GetUserByEmail(r.Context(), email)
		if err == nil {
			renderUserNewError(w, views, cfg, "User with this email already exists", name, email, role)
			return
		}

		// Hash password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			renderUserNewError(w, views, cfg, "Failed to hash password", name, email, role)
			return
		}

		// Create user
		user, err := db.CreateUser(r.Context(), input, string(passwordHash))
		if err != nil {
			renderUserNewError(w, views, cfg, "Failed to create user", name, email, role)
			return
		}

		// Update admin fields
		if err := db.UpdateUserAdmin(r.Context(), int(user.ID), isStaff, isSuperuser); err != nil {
			renderUserNewError(w, views, cfg, "Failed to set admin permissions", name, email, role)
			return
		}

		// Update active status if needed
		if !isActive {
			if err := db.UpdateUserActive(r.Context(), int(user.ID), isActive); err != nil {
				renderUserNewError(w, views, cfg, "Failed to set active status", name, email, role)
				return
			}
		}

		// Redirect to users list
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// handleUserDelete deletes a user
func handleUserDelete(views *view.Engine, cfg config.SecurityConfig) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Parse form to get CSRF token
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Validate CSRF token
		if !validateCSRFFromForm(r, cfg) {
			http.Error(w, "CSRF token invalid", http.StatusForbidden)
			return
		}

		// Delete user
		err = db.DeleteUser(r.Context(), id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}

		// Redirect to users list
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// renderUserNewError renders the create user form with an error
func renderUserNewError(w http.ResponseWriter, views *view.Engine, cfg config.SecurityConfig, errorMsg, name, email, role string) {
	data := map[string]any{
		"Title":     "Create New User",
		"Error":     errorMsg,
		"Name":      name,
		"Email":     email,
		"Role":      role,
		"CSRFToken": generateAndSetCSRFToken(w, cfg),
	}

	w.WriteHeader(http.StatusBadRequest)
	if err := views.Render(w, "admin/user_new", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// generateAndSetCSRFToken generates a CSRF token and sets the cookie
func generateAndSetCSRFToken(w http.ResponseWriter, cfg config.SecurityConfig) string {
	if !cfg.EnableCSRF {
		return ""
	}

	token, err := security.GenerateCSRFToken(cfg.JWTSecret)
	if err != nil {
		return ""
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	return token
}

// validateCSRFFromForm validates CSRF token from form submission
func validateCSRFFromForm(r *http.Request, cfg config.SecurityConfig) bool {
	if !cfg.EnableCSRF {
		return true
	}

	csrfCookie, err := r.Cookie(cfg.CSRFCookieName)
	if err != nil || csrfCookie.Value == "" {
		return false
	}

	csrfFormToken := r.FormValue("csrf_token")
	return security.ValidateCSRFToken(cfg.JWTSecret, csrfCookie.Value, csrfFormToken)
}
