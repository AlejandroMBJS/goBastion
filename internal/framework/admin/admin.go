package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AlejandroMBJS/goBastion/internal/app/models"
	"github.com/AlejandroMBJS/goBastion/internal/framework/db"
	"github.com/AlejandroMBJS/goBastion/internal/framework/middleware"
	frameworkrouter "github.com/AlejandroMBJS/goBastion/internal/framework/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/view"
)

// RegisterRoutes registers admin routes
func RegisterRoutes(r *frameworkrouter.Router, views *view.Engine) {
	// Admin routes require authentication and admin role
	adminAuth := middleware.RequireRole("admin")

	// GET /admin - Dashboard
	r.Handle("GET", "/admin", adminAuth(handleDashboard(views)))

	// GET /admin/users - List users
	r.Handle("GET", "/admin/users", adminAuth(handleUsersList(views)))

	// GET /admin/users/{id} - View/edit user
	r.Handle("GET", "/admin/users/{id}", adminAuth(handleUserDetail(views)))

	// POST /admin/users/{id} - Update user
	r.Handle("POST", "/admin/users/{id}", adminAuth(handleUserUpdate(views)))
}

// handleDashboard renders the admin dashboard
func handleDashboard(views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		claims := middleware.GetClaims(r.Context())
		if claims == nil {
			http.Redirect(w, r, "/api/v1/auth/login", http.StatusSeeOther)
			return
		}

		data := map[string]any{
			"Title":    "Admin Dashboard",
			"UserName": claims.Sub,
		}

		if err := views.Render(w, "admin/dashboard", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleUsersList renders the users list page
func handleUsersList(views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		users, err := db.ListUsers(r.Context())
		if err != nil {
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"Title": "Users",
			"Users": users,
		}

		if err := views.Render(w, "admin/users_list", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleUserDetail renders the user detail/edit page
func handleUserDetail(views *view.Engine) frameworkrouter.Handler {
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
			"Title": fmt.Sprintf("Edit User: %s", user.Name),
			"User":  user,
		}

		if err := views.Render(w, "admin/user_detail", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleUserUpdate processes the user update form
func handleUserUpdate(views *view.Engine) frameworkrouter.Handler {
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

		// Handle form submission
		name := r.FormValue("name")
		email := r.FormValue("email")
		role := r.FormValue("role")
		isStaff := r.FormValue("is_staff") == "on"
		isSuperuser := r.FormValue("is_superuser") == "on"

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

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
