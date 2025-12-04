// Package config provides centralized configuration management for goBastion.
//
// ⚠️ FRAMEWORK CORE - MODIFY CAREFULLY
//
// This package defines the configuration structure and loading logic for the entire
// goBastion framework. It provides:
//   - Type-safe configuration structs
//   - JSON file loading with defaults
//   - Environment variable overrides
//   - Validation of critical settings
//
// WHEN TO MODIFY:
//   - ✅ ADD new configuration fields to existing structs (extend)
//   - ✅ ADD new top-level config sections (e.g., EmailConfig)
//   - ✅ ADD new environment variable overrides in Load()
//   - ⚠️  MODIFY existing field types carefully (may break apps)
//   - ❌ DO NOT remove existing configuration fields (breaking change)
//   - ❌ DO NOT change JSON field names (breaks config files)
//
// CONFIGURATION FLOW:
//  1. Load() reads config/config.json
//  2. Applies sensible defaults for missing values
//  3. Overrides with environment variables (if set)
//  4. Returns fully populated Config struct
//
// USAGE IN YOUR APP:
//
//	import "github.com/AlejandroMBJS/goBastion/internal/framework/config"
//
//	cfg, err := config.Load("config/config.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access configuration
//	port := cfg.Server.Port
//	dbDriver := cfg.Database.Driver
//	appName := cfg.App.Name
//
// EXTENDING:
// To add a new config section:
//
//  1. Define struct with json tags:
//       type EmailConfig struct {
//           SMTPHost string `json:"smtp_host"`
//           SMTPPort int    `json:"smtp_port"`
//       }
//
//  2. Add to Config struct:
//       type Config struct {
//           ...
//           Email EmailConfig `json:"email"`
//       }
//
//  3. Add defaults in Load():
//       Email: EmailConfig{
//           SMTPHost: "localhost",
//           SMTPPort: 587,
//       },
//
//  4. Add env overrides in Load() if needed:
//       if host := os.Getenv("APP_SMTP_HOST"); host != "" {
//           cfg.Email.SMTPHost = host
//       }
//
// For documentation on config fields, see: README.md "Configuration" section
package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

// AppConfig holds application-level settings like name, environment, and branding.
//
// These settings control application-wide behavior and are typically displayed
// in the UI or used for environment-specific logic.
type AppConfig struct {
	Name        string `json:"name"`         // Application name (e.g., "goBastion")
	Environment string `json:"environment"`  // "development" or "production"
	BaseURL     string `json:"base_url"`     // Base URL for the application
	Locale      string `json:"locale"`       // Default locale (e.g., "en-US")
	Description string `json:"description"`  // App description for meta tags
}

type ServerConfig struct {
	Host                string   `json:"host"`                 // Server host (e.g., "0.0.0.0")
	Port                string   `json:"port"`                 // Server port (e.g., ":8080")
	ReadTimeoutSeconds  int      `json:"read_timeout_seconds"` // Read timeout in seconds
	WriteTimeoutSeconds int      `json:"write_timeout_seconds"` // Write timeout in seconds
	IdleTimeoutSeconds  int      `json:"idle_timeout_seconds"`  // Idle timeout in seconds
	AllowedOrigins      []string `json:"allowed_origins"`       // CORS allowed origins
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
	Enabled           bool `json:"enabled"`             // Enable rate limiting
	RequestsPerMinute int  `json:"requests_per_minute"` // Max requests per minute per IP
}

// FrontendConfig holds frontend/theme settings
type FrontendConfig struct {
	Theme ThemeConfig `json:"theme"` // Theme configuration
}

// ThemeConfig holds UI theme settings
type ThemeConfig struct {
	PrimaryColor   string `json:"primary_color"`    // Primary color (e.g., "indigo", "blue", "purple")
	SecondaryColor string `json:"secondary_color"`  // Secondary color
	DarkMode       bool   `json:"dark_mode"`        // Enable dark mode
	LogoPath       string `json:"logo_path"`        // Path to logo image
	FaviconPath    string `json:"favicon_path"`     // Path to favicon
}

// AdminConfig holds admin panel settings
type AdminConfig struct {
	EnableDashboardMetrics bool   `json:"enable_dashboard_metrics"` // Show tech metrics on dashboard
	DefaultAdminEmail      string `json:"default_admin_email"`      // Default admin email
	RegistrationOpen       bool   `json:"registration_open"`        // Allow new user registration
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level       string `json:"level"`         // Log level: "debug", "info", "warn", "error"
	Format      string `json:"format"`        // Log format: "json" or "text"
	RequestID   bool   `json:"request_id"`    // Log request IDs
	SQLQueries  bool   `json:"sql_queries"`   // Log SQL queries (development only)
	Debug       bool   `json:"debug"`         // Enable debug mode (verbose logging like NextJS)
	Verbose     bool   `json:"verbose"`       // Enable verbose template/framework debugging
}

// FeaturesConfig holds feature flags
type FeaturesConfig struct {
	EnableChat          bool `json:"enable_chat"`           // Enable real-time chat feature
	EnableNotifications bool `json:"enable_notifications"`  // Enable notifications demo
	EnableMetrics       bool `json:"enable_metrics"`        // Enable metrics collection
}

type Config struct {
	App       AppConfig       `json:"app"`        // Application settings
	Server    ServerConfig    `json:"server"`     // Server settings
	Database  DatabaseConfig  `json:"database"`   // Database settings
	Security  SecurityConfig  `json:"security"`   // Security settings
	RateLimit RateLimitConfig `json:"rate_limit"` // Rate limiting settings
	Frontend  FrontendConfig  `json:"frontend"`   // Frontend/theme settings
	Admin     AdminConfig     `json:"admin"`      // Admin panel settings
	Logging   LoggingConfig   `json:"logging"`    // Logging settings
	Features  FeaturesConfig  `json:"features"`   // Feature flags
}

// Load reads configuration from a file and applies environment variable overrides
func Load(path string) (Config, error) {
	// Start with sane defaults
	cfg := Config{
		App: AppConfig{
			Name:        "goBastion",
			Environment: "development",
			BaseURL:     "http://localhost:8080",
			Locale:      "en-US",
			Description: "Modern Go web framework with security and performance built-in",
		},
		Server: ServerConfig{
			Host:                "0.0.0.0",
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
		Frontend: FrontendConfig{
			Theme: ThemeConfig{
				PrimaryColor:   "indigo",
				SecondaryColor: "purple",
				DarkMode:       false,
				LogoPath:       "/static/logo.svg",
				FaviconPath:    "/static/favicon.ico",
			},
		},
		Admin: AdminConfig{
			EnableDashboardMetrics: true,
			DefaultAdminEmail:      "admin@example.com",
			RegistrationOpen:       true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			RequestID:  true,
			SQLQueries: false,
			Debug:      false,
			Verbose:    false,
		},
		Features: FeaturesConfig{
			EnableChat:          true,
			EnableNotifications: false,
			EnableMetrics:       true,
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
