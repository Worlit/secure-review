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
	"github.com/secure-review/internal/database"
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

	// Connect to database using GORM (аналог TypeORM DataSource)
	db, err := database.NewDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	log.Println("Connected to database via GORM")

	// Run auto migrations (аналог TypeORM synchronize: true)
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run auto migrations: %v", err)
	}

	log.Println("Auto migrations completed")

	// Initialize repositories with adapters (аналог getRepository() в TypeORM)
	userRepo := repository.NewUserRepositoryAdapter(db.DB)
	reviewRepo := repository.NewReviewRepositoryAdapter(db.DB)

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
