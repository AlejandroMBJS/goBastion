package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlejandroMBJS/goBastion/internal/app/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/admin"
	"github.com/AlejandroMBJS/goBastion/internal/framework/config"
	"github.com/AlejandroMBJS/goBastion/internal/framework/db"
	"github.com/AlejandroMBJS/goBastion/internal/framework/docs"
	"github.com/AlejandroMBJS/goBastion/internal/framework/middleware"
	frameworkrouter "github.com/AlejandroMBJS/goBastion/internal/framework/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/view"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting server with configuration:")
	log.Printf("  - Port: %s", cfg.Server.Port)
	log.Printf("  - Database Driver: %s", cfg.Database.Driver)
	log.Printf("  - CSRF Enabled: %v", cfg.Security.EnableCSRF)
	log.Printf("  - JWT Enabled: %v", cfg.Security.EnableJWT)
	log.Printf("  - Rate Limiting Enabled: %v", cfg.RateLimit.Enabled)

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

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", cfg.Server.Port)
		log.Printf("  - API Documentation: http://localhost%s/docs", cfg.Server.Port)
		log.Printf("  - Admin Panel: http://localhost%s/admin", cfg.Server.Port)
		log.Printf("  - Health Check: http://localhost%s/api/v1/users", cfg.Server.Port)
		log.Println("\nPress Ctrl+C to shutdown")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\nShutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}
