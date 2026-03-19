package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/secure-review/docs" // Init swagger docs
	"github.com/secure-review/internal/config"
	"github.com/secure-review/internal/handler"
	"github.com/secure-review/internal/middleware"
)

// Router holds all handlers and middleware
type Router struct {
	config         *config.Config
	authHandler    *handler.AuthHandler
	githubHandler  *handler.GitHubHandler
	userHandler    *handler.UserHandler
	reviewHandler  *handler.ReviewHandler
	healthHandler  *handler.HealthHandler
	authMiddleware *middleware.AuthMiddleware
}

// NewRouter creates a new Router
func NewRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	githubHandler *handler.GitHubHandler,
	userHandler *handler.UserHandler,
	reviewHandler *handler.ReviewHandler,
	healthHandler *handler.HealthHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		config:         cfg,
		authHandler:    authHandler,
		githubHandler:  githubHandler,
		userHandler:    userHandler,
		reviewHandler:  reviewHandler,
		healthHandler:  healthHandler,
		authMiddleware: authMiddleware,
	}
}

// Setup sets up the router with all routes
func (r *Router) Setup() *gin.Engine {
	if r.config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Global middleware
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())

	// Swagger
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Redirect root to swagger
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// CORS setup
	allowOrigins := strings.Split(r.config.Frontend.URL, ",")
	var cleanOrigins []string
	for _, origin := range allowOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			cleanOrigins = append(cleanOrigins, trimmed)
		}
	}

	// Add localhost for development if not present
	hasLocalhost := false
	for _, origin := range cleanOrigins {
		if origin == "http://localhost:3000" {
			hasLocalhost = true
			break
		}
	}
	if !hasLocalhost {
		cleanOrigins = append(cleanOrigins, "http://localhost:3000")
	}

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     cleanOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check routes (no auth required)
	engine.GET("/health", r.healthHandler.Health)
	engine.GET("/ready", r.healthHandler.Ready)

	// API routes
	api := engine.Group("/api/v1")
	{
		// Auth routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)

			// GitHub OAuth
			auth.GET("/github", r.githubHandler.GetAuthURL)
			auth.POST("/github/callback", r.githubHandler.Callback)
			auth.GET("/github/callback", r.githubHandler.CallbackRedirect) // Обработка редиректа от GitHub (Plan B)
		}

		// GitHub Webhooks
		githubPublic := api.Group("/github")
		{
			githubPublic.POST("/webhook", r.githubHandler.Webhook)
		}

		// Protected auth routes
		authProtected := api.Group("/auth")
		authProtected.Use(r.authMiddleware.RequireAuth())
		{
			authProtected.POST("/refresh", r.authHandler.RefreshToken)
			authProtected.POST("/change-password", r.authHandler.ChangePassword)
			authProtected.POST("/github/link", r.githubHandler.LinkAccount)
			authProtected.DELETE("/github/link", r.githubHandler.UnlinkAccount)
		}

		// User routes (auth required)
		users := api.Group("/users")
		users.Use(r.authMiddleware.RequireAuth())
		{
			users.GET("/me", r.userHandler.GetProfile)
			users.PUT("/me", r.userHandler.UpdateProfile)
			users.DELETE("/me", r.userHandler.DeleteAccount)
			users.GET("/repos", r.githubHandler.ListRepositories)
		}

		// GitHub Data routes (auth required)
		gh := api.Group("/github")
		gh.Use(r.authMiddleware.RequireAuth())
		{
			gh.GET("/repos", r.githubHandler.ListRepositories)
			gh.GET("/repos/:owner/:repo/branches", r.githubHandler.ListBranches)
		}

		// Review routes (auth required)
		reviews := api.Group("/reviews")
		reviews.Use(r.authMiddleware.RequireAuth())
		{
			reviews.POST("", r.reviewHandler.CreateReview)
			reviews.GET("", r.reviewHandler.ListReviews)
			reviews.GET("/:id", r.reviewHandler.GetReview)
			reviews.GET("/:id/pdf", r.reviewHandler.GetReviewPDF)
			reviews.DELETE("/:id", r.reviewHandler.DeleteReview)
			reviews.POST("/:id/reanalyze", r.reviewHandler.ReanalyzeReview)
		}
	}

	return engine
}
