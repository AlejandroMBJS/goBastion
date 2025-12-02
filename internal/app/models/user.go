package models

import (
	"errors"
	"strings"
)

// User represents a user in the system
type User struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
	IsStaff     bool   `json:"is_staff"`
	IsSuperuser bool   `json:"is_superuser"`
}

// UserInput represents input for creating or updating a user
type UserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// RegisterInput represents input for user registration
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// LoginInput represents input for user login
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshInput represents input for token refresh
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// Validate validates UserInput
func (u UserInput) Validate() error {
	if len(strings.TrimSpace(u.Name)) < 2 {
		return errors.New("name must be at least 2 characters long")
	}

	if len(u.Name) > 100 {
		return errors.New("name must not exceed 100 characters")
	}

	if !strings.Contains(u.Email, "@") {
		return errors.New("email must be a valid email address")
	}

	if len(u.Email) > 255 {
		return errors.New("email must not exceed 255 characters")
	}

	if u.Role != "" && u.Role != "user" && u.Role != "admin" {
		return errors.New("role must be either 'user' or 'admin'")
	}

	return nil
}

// Validate validates RegisterInput
func (r RegisterInput) Validate() error {
	if len(strings.TrimSpace(r.Name)) < 2 {
		return errors.New("name must be at least 2 characters long")
	}

	if len(r.Name) > 100 {
		return errors.New("name must not exceed 100 characters")
	}

	if !strings.Contains(r.Email, "@") {
		return errors.New("email must be a valid email address")
	}

	if len(r.Email) > 255 {
		return errors.New("email must not exceed 255 characters")
	}

	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(r.Password) > 100 {
		return errors.New("password must not exceed 100 characters")
	}

	if r.Role != "" && r.Role != "user" && r.Role != "admin" {
		return errors.New("role must be either 'user' or 'admin'")
	}

	return nil
}

// Validate validates LoginInput
func (l LoginInput) Validate() error {
	if !strings.Contains(l.Email, "@") {
		return errors.New("email must be a valid email address")
	}

	if len(l.Password) < 1 {
		return errors.New("password is required")
	}

	return nil
}

// Validate validates RefreshInput
func (r RefreshInput) Validate() error {
	if len(strings.TrimSpace(r.RefreshToken)) == 0 {
		return errors.New("refresh token is required")
	}

	return nil
}

// ToUserInput converts a User to UserInput
func (u User) ToUserInput() UserInput {
	return UserInput{
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
	}
}

// PublicUser returns a user without sensitive information
func (u User) PublicUser() map[string]any {
	return map[string]any{
		"id":           u.ID,
		"name":         u.Name,
		"email":        u.Email,
		"role":         u.Role,
		"is_active":    u.IsActive,
		"is_staff":     u.IsStaff,
		"is_superuser": u.IsSuperuser,
	}
}
