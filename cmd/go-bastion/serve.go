package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AlejandroMBJS/goBastion/internal/app/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/admin"
	"github.com/AlejandroMBJS/goBastion/internal/framework/config"
	"github.com/AlejandroMBJS/goBastion/internal/framework/db"
	"github.com/AlejandroMBJS/goBastion/internal/framework/docs"
	"github.com/AlejandroMBJS/goBastion/internal/framework/middleware"
	frameworkrouter "github.com/AlejandroMBJS/goBastion/internal/framework/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/view"
)

type configModel struct {
    cfg      *config.Config
    env      string
    quitting bool
    ready    bool
}

func (m configModel) Init() tea.Cmd {
    return func() tea.Msg {
        return readyMsg{}
    }
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" || msg.String() == "ctrl+c" {
            m.quitting = true
            return m, tea.Quit
        }
    case readyMsg:
        m.ready = true
        return m, tea.Quit
    }
    return m, nil
}

type readyMsg struct{}


func (m configModel) View() string {
	if m.quitting {
		return ""
	}

	var s string

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 2).
		Width(70).
		Align(lipgloss.Center)

	s += headerStyle.Render("GO-BASTION WEB SERVER") + "\n\n"

	// Configuration Section
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Underline(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")).
		Bold(true).
		Width(25)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	enabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	disabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6347")).
		Bold(true)

	s += titleStyle.Render("SERVER CONFIGURATION") + "\n\n"

	// Server Info
	s += labelStyle.Render("Port:") + valueStyle.Render(m.cfg.Server.Port) + "\n"
	s += labelStyle.Render("Environment:") + valueStyle.Render(m.env) + "\n\n"

	// Database Info
	s += titleStyle.Render("DATABASE CONFIGURATION") + "\n\n"
	s += labelStyle.Render("Driver:") + valueStyle.Render(m.cfg.Database.Driver) + "\n"
	s += labelStyle.Render("DSN:") + valueStyle.Render(m.cfg.Database.DSN) + "\n\n"

	// Security Features
	s += titleStyle.Render("SECURITY FEATURES") + "\n\n"

	csrfStatus := disabledStyle.Render("DISABLED")
	if m.cfg.Security.EnableCSRF {
		csrfStatus = enabledStyle.Render("ENABLED")
	}
	s += labelStyle.Render("CSRF Protection:") + csrfStatus + "\n"

	jwtStatus := disabledStyle.Render("DISABLED")
	if m.cfg.Security.EnableJWT {
		jwtStatus = enabledStyle.Render("ENABLED")
	}
	s += labelStyle.Render("JWT Authentication:") + jwtStatus + "\n"
	s += labelStyle.Render("JWT Access TTL:") + valueStyle.Render(fmt.Sprintf("%d minutes", m.cfg.Security.AccessTokenMinutes)) + "\n"
	s += labelStyle.Render("JWT Refresh TTL:") + valueStyle.Render(fmt.Sprintf("%d minutes", m.cfg.Security.RefreshTokenMinutes)) + "\n\n"

	// Rate Limiting
	s += titleStyle.Render("RATE LIMITING") + "\n\n"

	rateLimitStatus := disabledStyle.Render("DISABLED")
	if m.cfg.RateLimit.Enabled {
		rateLimitStatus = enabledStyle.Render("ENABLED")
	}
	s += labelStyle.Render("Status:") + rateLimitStatus + "\n"
	s += labelStyle.Render("Max Requests:") + valueStyle.Render(fmt.Sprintf("%d per minute", m.cfg.RateLimit.RequestsPerMinute)) + "\n\n"

	// Routes
	s += titleStyle.Render("AVAILABLE ROUTES") + "\n\n"

	routeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB"))

	s += routeStyle.Render("  • API Documentation:    http://localhost"+m.cfg.Server.Port+"/docs") + "\n"
	s += routeStyle.Render("  • Admin Panel:          http://localhost"+m.cfg.Server.Port+"/admin") + "\n"
	s += routeStyle.Render("  • Login Page:           http://localhost"+m.cfg.Server.Port+"/login") + "\n"
	s += routeStyle.Render("  • Register Page:        http://localhost"+m.cfg.Server.Port+"/register") + "\n"
	s += routeStyle.Render("  • Health Check:         http://localhost"+m.cfg.Server.Port+"/api/v1/users") + "\n"
	s += routeStyle.Render("  • Auth Register API:    POST http://localhost"+m.cfg.Server.Port+"/api/v1/auth/register") + "\n"
	s += routeStyle.Render("  • Auth Login API:       POST http://localhost"+m.cfg.Server.Port+"/api/v1/auth/login") + "\n\n"

	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080")).
		Italic(true)

	if !m.ready {
		s += footerStyle.Render("Press Ctrl+C to shutdown") + "\n"
	}

	return s
}

func displayConfigAndServe(env string) {
	// Load configuration
	cfg, err := config.Load("config/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if env == "" {
		env = "development"
	}

	// Create and run the Bubble Tea program to display config
	p := tea.NewProgram(configModel{cfg: &cfg, env: env})

	// Display the config
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error displaying config: %v\n", err)
		os.Exit(1)
	}

	// Now start the actual server
	startServer(&cfg)
}

func startServer(cfg *config.Config) {
	// Initialize database
	if err := db.Init(cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	// Initialize template engine
	tmplEngine, err := view.NewEngine("templates")
	if err != nil {
		log.Fatalf("Failed to initialize template engine: %v", err)
	}
	log.Println("Template engine initialized successfully")

	// Create router
	r := frameworkrouter.New()

	// Register global middlewares (order matters!)
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging)
	r.Use(middleware.Recover)
	r.Use(middleware.WithTimeout(5 * time.Second))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.MaxBodySize(cfg.Security.MaxBodyBytes))
	r.Use(middleware.CORSMiddleware(cfg.Server.AllowedOrigins))

	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimit(cfg.RateLimit))
		log.Println("Rate limiting enabled")
	}

	if cfg.Security.EnableCSRF {
		r.Use(middleware.CSRFMiddleware(cfg.Security))
		log.Println("CSRF protection enabled")
	}

	if cfg.Security.EnableJWT {
		r.Use(middleware.JWTAuthMiddleware(cfg.Security))
		log.Println("JWT authentication enabled")
	}

	// Register application routes
	log.Println("Registering application routes...")
	router.RegisterAuthRoutes(r, cfg.Security)
	router.RegisterUserRoutes(r)

	// Register HTML authentication routes
	log.Println("Registering HTML authentication routes...")
	router.RegisterAuthViewsRoutes(r, cfg.Security, tmplEngine)

	// Register admin routes
	log.Println("Registering admin routes...")
	admin.RegisterRoutes(r, tmplEngine)

	// Register documentation routes
	log.Println("Registering documentation routes...")
	docs.RegisterRoutes(r)

	// Serve static files (CSS)
	log.Println("Registering static file routes...")
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	r.Handle("GET", "/static/css/output.css", frameworkrouter.WrapHandler(staticHandler))

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.GetReadTimeout(),
		WriteTimeout: cfg.Server.GetWriteTimeout(),
		IdleTimeout:  cfg.Server.GetIdleTimeout(),
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("\nShutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server
	log.Printf("Server starting on %s", cfg.Server.Port)
	log.Println("\nPress Ctrl+C to shutdown")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped")
}
