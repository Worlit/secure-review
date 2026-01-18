package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/secure-review/internal/config"
	"github.com/secure-review/internal/handler"
	"github.com/secure-review/internal/middleware"
	"github.com/secure-review/internal/repository"
	"github.com/secure-review/internal/router"
	"github.com/secure-review/internal/service"
)

const version = "1.0.0"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := repository.NewPostgresDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Run migrations
	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed")

	// Initialize repositories
	userRepo := repository.NewPostgresUserRepository(db)
	reviewRepo := repository.NewPostgresReviewRepository(db)

	// Initialize services
	passwordHasher := service.NewBcryptPasswordHasher()
	tokenGenerator := service.NewJWTTokenGenerator(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.ExpirationHours)*time.Hour,
		time.Duration(cfg.JWT.ExpirationHours*7)*time.Hour,
	)
	codeAnalyzer := service.NewOpenAICodeAnalyzer(cfg.OpenAI.APIKey)

	authService := service.NewAuthService(userRepo, passwordHasher, tokenGenerator)
	userService := service.NewUserService(userRepo)
	githubAuthService := service.NewGitHubAuthService(
		cfg.GitHub.ClientID,
		cfg.GitHub.ClientSecret,
		cfg.GitHub.RedirectURL,
		userRepo,
		tokenGenerator,
	)
	reviewService := service.NewReviewService(reviewRepo, codeAnalyzer)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	githubHandler := handler.NewGitHubHandler(githubAuthService)
	userHandler := handler.NewUserHandler(userService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	healthHandler := handler.NewHealthHandler(version)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup router
	r := router.NewRouter(
		cfg,
		authHandler,
		githubHandler,
		userHandler,
		reviewHandler,
		healthHandler,
		authMiddleware,
	)

	engine := r.Setup()

	// Create server
	srv := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.GetServerAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
