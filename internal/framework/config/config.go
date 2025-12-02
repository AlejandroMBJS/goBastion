package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type ServerConfig struct {
	Port                string   `json:"port"`
	ReadTimeoutSeconds  int      `json:"read_timeout_seconds"`
	WriteTimeoutSeconds int      `json:"write_timeout_seconds"`
	IdleTimeoutSeconds  int      `json:"idle_timeout_seconds"`
	AllowedOrigins      []string `json:"allowed_origins"`
}

type DatabaseConfig struct {
	Driver                 string `json:"driver"`
	DSN                    string `json:"dsn"`
	MaxOpenConns           int    `json:"max_open_conns"`
	MaxIdleConns           int    `json:"max_idle_conns"`
	ConnMaxLifetimeMinutes int    `json:"conn_max_lifetime_minutes"`
}

type SecurityConfig struct {
	EnableCSRF          bool   `json:"enable_csrf"`
	CSRFHeaderName      string `json:"csrf_header_name"`
	CSRFCookieName      string `json:"csrf_cookie_name"`
	EnableJWT           bool   `json:"enable_jwt"`
	JWTSecret           string `json:"jwt_secret"`
	AccessTokenMinutes  int    `json:"access_token_minutes"`
	RefreshTokenMinutes int    `json:"refresh_token_minutes"`
	MaxBodyBytes        int64  `json:"max_body_bytes"`
}

type RateLimitConfig struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
}

type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	Security  SecurityConfig  `json:"security"`
	RateLimit RateLimitConfig `json:"rate_limit"`
}

// Load reads configuration from a file and applies environment variable overrides
func Load(path string) (Config, error) {
	// Start with sane defaults
	cfg := Config{
		Server: ServerConfig{
			Port:                ":8080",
			ReadTimeoutSeconds:  10,
			WriteTimeoutSeconds: 10,
			IdleTimeoutSeconds:  60,
			AllowedOrigins:      []string{"http://localhost:3000"},
		},
		Database: DatabaseConfig{
			Driver:                 "sqlite3",
			DSN:                    "file:api.db?_foreign_keys=on",
			MaxOpenConns:           10,
			MaxIdleConns:           5,
			ConnMaxLifetimeMinutes: 30,
		},
		Security: SecurityConfig{
			EnableCSRF:          true,
			CSRFHeaderName:      "X-CSRF-Token",
			CSRFCookieName:      "csrf_token",
			EnableJWT:           true,
			JWTSecret:           "change-me-in-prod",
			AccessTokenMinutes:  15,
			RefreshTokenMinutes: 4320,
			MaxBodyBytes:        1048576,
		},
		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 60,
		},
	}

	// Load from file if it exists
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, err
		}
	}

	// Apply environment variable overrides
	if port := os.Getenv("APP_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if driver := os.Getenv("APP_DB_DRIVER"); driver != "" {
		cfg.Database.Driver = driver
	}
	if dsn := os.Getenv("APP_DB_DSN"); dsn != "" {
		cfg.Database.DSN = dsn
	}
	if secret := os.Getenv("APP_JWT_SECRET"); secret != "" {
		cfg.Security.JWTSecret = secret
	}
	if maxBody := os.Getenv("APP_MAX_BODY_BYTES"); maxBody != "" {
		if val, err := strconv.ParseInt(maxBody, 10, 64); err == nil {
			cfg.Security.MaxBodyBytes = val
		}
	}
	if origins := os.Getenv("APP_ALLOWED_ORIGINS"); origins != "" {
		cfg.Server.AllowedOrigins = strings.Split(origins, ",")
	}
	if readTimeout := os.Getenv("APP_READ_TIMEOUT_SECONDS"); readTimeout != "" {
		if val, err := strconv.Atoi(readTimeout); err == nil {
			cfg.Server.ReadTimeoutSeconds = val
		}
	}
	if writeTimeout := os.Getenv("APP_WRITE_TIMEOUT_SECONDS"); writeTimeout != "" {
		if val, err := strconv.Atoi(writeTimeout); err == nil {
			cfg.Server.WriteTimeoutSeconds = val
		}
	}

	return cfg, nil
}

// GetReadTimeout returns the configured read timeout as time.Duration
func (s ServerConfig) GetReadTimeout() time.Duration {
	return time.Duration(s.ReadTimeoutSeconds) * time.Second
}

// GetWriteTimeout returns the configured write timeout as time.Duration
func (s ServerConfig) GetWriteTimeout() time.Duration {
	return time.Duration(s.WriteTimeoutSeconds) * time.Second
}

// GetIdleTimeout returns the configured idle timeout as time.Duration
func (s ServerConfig) GetIdleTimeout() time.Duration {
	return time.Duration(s.IdleTimeoutSeconds) * time.Second
}

// GetConnMaxLifetime returns the configured connection max lifetime as time.Duration
func (d DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(d.ConnMaxLifetimeMinutes) * time.Minute
}
