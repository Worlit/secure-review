package router

import (
	"github.com/gin-gonic/gin"

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
	engine.Use(middleware.CORSWithConfig(r.config.Frontend.URL))

	// Health check routes (no auth required)
	engine.GET("/health", r.healthHandler.Health)
	engine.GET("/ready", r.healthHandler.Ready)

	// API routes
	api := engine.Group("/api")
	{
		// Auth routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)

			// GitHub OAuth
			auth.GET("/github", r.githubHandler.GetAuthURL)
			auth.GET("/github/callback", r.githubHandler.Callback)
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
		}

		// Review routes (auth required)
		reviews := api.Group("/reviews")
		reviews.Use(r.authMiddleware.RequireAuth())
		{
			reviews.POST("", r.reviewHandler.CreateReview)
			reviews.GET("", r.reviewHandler.ListReviews)
			reviews.GET("/:id", r.reviewHandler.GetReview)
			reviews.DELETE("/:id", r.reviewHandler.DeleteReview)
			reviews.POST("/:id/reanalyze", r.reviewHandler.ReanalyzeReview)
		}
	}

	return engine
}
