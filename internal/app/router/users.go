package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-native-fastapi/internal/app/models"
	"go-native-fastapi/internal/framework/db"
	frameworkrouter "go-native-fastapi/internal/framework/router"

	"golang.org/x/crypto/bcrypt"
)

// RegisterUserRoutes registers user CRUD routes
func RegisterUserRoutes(r *frameworkrouter.Router) {
	// GET /api/v1/users - List all users
	r.Handle("GET", "/api/v1/users", handleListUsers)

	// POST /api/v1/users - Create a new user
	r.Handle("POST", "/api/v1/users", handleCreateUser)

	// GET /api/v1/users/{id} - Get a user by ID
	r.Handle("GET", "/api/v1/users/{id}", handleGetUser)

	// PUT /api/v1/users/{id} - Update a user
	r.Handle("PUT", "/api/v1/users/{id}", handleUpdateUser)

	// DELETE /api/v1/users/{id} - Delete a user
	r.Handle("DELETE", "/api/v1/users/{id}", handleDeleteUser)
}

// handleListUsers lists all users
func handleListUsers(w http.ResponseWriter, r *http.Request, params map[string]string) {
	users, err := db.ListUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list users"})
		return
	}

	// If no users, return empty array instead of null
	if users == nil {
		users = []models.User{}
	}

	writeJSON(w, http.StatusOK, users)
}

// handleCreateUser creates a new user
func handleCreateUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
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

	// Hash password (using bcrypt)
	// Note: In production, you might want to use a stronger password policy
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

	writeJSON(w, http.StatusCreated, user)
}

// handleGetUser gets a user by ID
func handleGetUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	user, err := db.GetUser(r.Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// handleUpdateUser updates a user
func handleUpdateUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	var input models.UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate input
	if err := input.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Update user
	user, err := db.UpdateUser(r.Context(), id, input)
	if err != nil {
		if err == db.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// handleDeleteUser deletes a user
func handleDeleteUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	err = db.DeleteUser(r.Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
