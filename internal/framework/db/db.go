package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlejandroMBJS/goBastion/internal/app/models"
	"github.com/AlejandroMBJS/goBastion/internal/framework/config"

	_ "github.com/mattn/go-sqlite3"
	// Uncomment the driver you need:
	// _ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
	// _ "github.com/godror/godror"
)

var (
	DB          *sql.DB
	ErrNotFound = errors.New("not found")
)

// Init initializes the database connection
func Init(cfg config.DatabaseConfig) error {
	var err error
	DB, err = sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetConnMaxLifetime(cfg.GetConnMaxLifetime())

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// migrate creates the necessary tables
func migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		role TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		is_staff INTEGER NOT NULL DEFAULT 0,
		is_superuser INTEGER NOT NULL DEFAULT 0,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	`

	_, err := DB.Exec(schema)
	return err
}

// ListUsers retrieves all users using the query builder
func ListUsers(ctx context.Context) ([]models.User, error) {
	// Use the framework's FindAll helper
	rows, err := FindAll(ctx, "users", "id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var isActive, isStaff, isSuperuser int
		var passwordHash string
		var createdAt sql.NullTime

		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &isActive, &isStaff, &isSuperuser, &passwordHash, &createdAt)
		if err != nil {
			return nil, err
		}
		u.IsActive = isActive == 1
		u.IsStaff = isStaff == 1
		u.IsSuperuser = isSuperuser == 1
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CreateUser creates a new user with the given password hash using query builder
func CreateUser(ctx context.Context, in models.RegisterInput, passwordHash string) (models.User, error) {
	// Default role to "user" if not specified
	role := in.Role
	if role == "" {
		role = "user"
	}

	// Use the framework's Insert helper
	data := map[string]any{
		"name":          in.Name,
		"email":         in.Email,
		"role":          role,
		"password_hash": passwordHash,
		"is_active":     1,
		"is_staff":      0,
		"is_superuser":  0,
	}

	id, err := Insert(ctx, "users", data)
	if err != nil {
		return models.User{}, err
	}

	return GetUser(ctx, int(id))
}

// GetUser retrieves a user by ID using query builder
func GetUser(ctx context.Context, id int) (models.User, error) {
	var u models.User
	var isActive, isStaff, isSuperuser int
	var passwordHash string
	var createdAt sql.NullTime

	// Use the framework's FindByID helper
	err := FindByID(ctx, "users", id, &u.ID, &u.Name, &u.Email, &u.Role, &isActive, &isStaff, &isSuperuser, &passwordHash, &createdAt)
	if err != nil {
		return models.User{}, err
	}

	u.IsActive = isActive == 1
	u.IsStaff = isStaff == 1
	u.IsSuperuser = isSuperuser == 1

	return u, nil
}

// GetUserByEmail retrieves a user by email and returns the password hash using query builder
func GetUserByEmail(ctx context.Context, email string) (models.User, string, error) {
	var u models.User
	var isActive, isStaff, isSuperuser int
	var passwordHash string
	var createdAt sql.NullTime

	// Use the framework's FindOneBy helper
	conditions := map[string]any{"email": email}
	err := FindOneBy(ctx, "users", conditions, &u.ID, &u.Name, &u.Email, &u.Role, &isActive, &isStaff, &isSuperuser, &passwordHash, &createdAt)
	if err != nil {
		return models.User{}, "", err
	}

	u.IsActive = isActive == 1
	u.IsStaff = isStaff == 1
	u.IsSuperuser = isSuperuser == 1

	return u, passwordHash, nil
}

// UpdateUser updates an existing user using query builder
func UpdateUser(ctx context.Context, id int, in models.UserInput) (models.User, error) {
	// Use the framework's UpdateByID helper
	data := map[string]any{
		"name":  in.Name,
		"email": in.Email,
		"role":  in.Role,
	}

	err := UpdateByID(ctx, "users", id, data)
	if err != nil {
		return models.User{}, err
	}

	return GetUser(ctx, id)
}

// UpdateUserAdmin updates user admin fields (is_staff, is_superuser) using query builder
func UpdateUserAdmin(ctx context.Context, id int, isStaff, isSuperuser bool) error {
	isStaffInt := 0
	if isStaff {
		isStaffInt = 1
	}
	isSuperuserInt := 0
	if isSuperuser {
		isSuperuserInt = 1
	}

	// Use the framework's UpdateByID helper
	data := map[string]any{
		"is_staff":     isStaffInt,
		"is_superuser": isSuperuserInt,
	}

	return UpdateByID(ctx, "users", id, data)
}

// UpdateUserActive updates user active status using query builder
func UpdateUserActive(ctx context.Context, id int, isActive bool) error {
	isActiveInt := 0
	if isActive {
		isActiveInt = 1
	}

	// Use the framework's UpdateByID helper
	data := map[string]any{
		"is_active": isActiveInt,
	}

	return UpdateByID(ctx, "users", id, data)
}

// DeleteUser deletes a user by ID using query builder
func DeleteUser(ctx context.Context, id int) error {
	// Use the framework's DeleteByID helper
	return DeleteByID(ctx, "users", id)
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
