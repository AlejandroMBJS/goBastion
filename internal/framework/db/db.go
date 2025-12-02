package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go-native-fastapi/internal/app/models"
	"go-native-fastapi/internal/framework/config"

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

// ListUsers retrieves all users
func ListUsers(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT id, name, email, role, is_active, is_staff, is_superuser
		FROM users
		ORDER BY id DESC
	`

	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var isActive, isStaff, isSuperuser int
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &isActive, &isStaff, &isSuperuser)
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

// CreateUser creates a new user with the given password hash
func CreateUser(ctx context.Context, in models.RegisterInput, passwordHash string) (models.User, error) {
	// Default role to "user" if not specified
	role := in.Role
	if role == "" {
		role = "user"
	}

	query := `
		INSERT INTO users (name, email, role, password_hash, is_active, is_staff, is_superuser)
		VALUES (?, ?, ?, ?, 1, 0, 0)
	`

	result, err := DB.ExecContext(ctx, query, in.Name, in.Email, role, passwordHash)
	if err != nil {
		return models.User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.User{}, err
	}

	return GetUser(ctx, int(id))
}

// GetUser retrieves a user by ID
func GetUser(ctx context.Context, id int) (models.User, error) {
	query := `
		SELECT id, name, email, role, is_active, is_staff, is_superuser
		FROM users
		WHERE id = ?
	`

	var u models.User
	var isActive, isStaff, isSuperuser int
	err := DB.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &isActive, &isStaff, &isSuperuser,
	)

	if err == sql.ErrNoRows {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, err
	}

	u.IsActive = isActive == 1
	u.IsStaff = isStaff == 1
	u.IsSuperuser = isSuperuser == 1

	return u, nil
}

// GetUserByEmail retrieves a user by email and returns the password hash
func GetUserByEmail(ctx context.Context, email string) (models.User, string, error) {
	query := `
		SELECT id, name, email, role, is_active, is_staff, is_superuser, password_hash
		FROM users
		WHERE email = ?
	`

	var u models.User
	var isActive, isStaff, isSuperuser int
	var passwordHash string

	err := DB.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &isActive, &isStaff, &isSuperuser, &passwordHash,
	)

	if err == sql.ErrNoRows {
		return models.User{}, "", ErrNotFound
	}
	if err != nil {
		return models.User{}, "", err
	}

	u.IsActive = isActive == 1
	u.IsStaff = isStaff == 1
	u.IsSuperuser = isSuperuser == 1

	return u, passwordHash, nil
}

// UpdateUser updates an existing user
func UpdateUser(ctx context.Context, id int, in models.UserInput) (models.User, error) {
	query := `
		UPDATE users
		SET name = ?, email = ?, role = ?
		WHERE id = ?
	`

	result, err := DB.ExecContext(ctx, query, in.Name, in.Email, in.Role, id)
	if err != nil {
		return models.User{}, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return models.User{}, err
	}

	if rows == 0 {
		return models.User{}, ErrNotFound
	}

	return GetUser(ctx, id)
}

// UpdateUserAdmin updates user admin fields (is_staff, is_superuser)
func UpdateUserAdmin(ctx context.Context, id int, isStaff, isSuperuser bool) error {
	query := `
		UPDATE users
		SET is_staff = ?, is_superuser = ?
		WHERE id = ?
	`

	isStaffInt := 0
	if isStaff {
		isStaffInt = 1
	}
	isSuperuserInt := 0
	if isSuperuser {
		isSuperuserInt = 1
	}

	result, err := DB.ExecContext(ctx, query, isStaffInt, isSuperuserInt, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteUser deletes a user by ID
func DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
