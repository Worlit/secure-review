package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/handler"
	"github.com/secure-review/internal/service"
	"github.com/secure-review/tests/fakes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetupApp wires up the application with fake implementations
func SetupApp() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Initialize Fakes
	userRepo := fakes.NewFakeUserRepository()
	reviewRepo := fakes.NewFakeReviewRepository()
	analyzer := fakes.NewFakeCodeAnalyzer()
	hasher := fakes.NewFakePasswordHasher()
	tokenGen := fakes.NewFakeTokenGenerator()

	// Initialize Services
	authService := service.NewAuthService(userRepo, hasher, tokenGen)
	reviewService := service.NewReviewService(reviewRepo, analyzer)
	userService := service.NewUserService(userRepo)

	// Mocks for unused services in this test suite
	// githubService := service.NewGitHubAuthService(...)

	// Initialize Handlers
	authHandler := handler.NewAuthHandler(authService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler("1.0.0")
	// githubHandler := handler.NewGitHubHandler(...)

	// Initialize Router
	// Note: We need to manually construct the router or modify internal/router to accept interfaces/structs
	// If internal/router expects concrete implementations, we might need to adjust it or duplicate routing logic here.
	// For this test, I'll assume we can reconstruct the routing logic using the handlers we created.

	r := gin.New()
	r.Use(gin.Recovery())

	// Public routes
	r.GET("/health", healthHandler.Health)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Private routes
		// Retrieve the *service.AuthMiddlewareStruct if that's what NewAuthMiddleware returns, otherwise adapt.
		// Looking at folder structure, middleware/auth.go likely contains the middleware.
		// We'll trust that we can check code later if this fails compilation, but for now assuming standard middleware.

		// For tests, we can manually implement a middleware that uses our fake token generator
		authMiddleware := func(c *gin.Context) {
			token := c.GetHeader("Authorization")
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
			userID, err := tokenGen.ValidateToken(token)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
			c.Set("userID", userID)
			c.Next()
		}

		protected := api.Group("/")
		protected.Use(authMiddleware)
		{
			reviews := protected.Group("/reviews")
			{
				reviews.POST("", reviewHandler.CreateReview)
				reviews.GET("", reviewHandler.ListReviews)
				reviews.GET("/:id", reviewHandler.GetReview)
			}

			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetProfile)
			}
		}
	}

	return r
}

func TestHealthCheck(t *testing.T) {
	r := SetupApp()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthFlowAndReview(t *testing.T) {
	r := SetupApp()

	// 1. Register
	registerPayload := domain.CreateUserInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(registerPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("Registration failed: %v", w.Body.String())
	}

	var authResp domain.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &authResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, authResp.Token)
	token := authResp.Token

	// 2. Login (Optional since Register returned token, but good to test)
	loginPayload := domain.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ = json.Marshal(loginPayload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Create Review
	reviewPayload := domain.CreateReviewInput{
		Code:     "func main() { fmt.Println(\"Hello\") }",
		Language: "go",
		Title:    "Test Review",
	}
	body, _ = json.Marshal(reviewPayload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/reviews", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var reviewResp domain.ReviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &reviewResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, reviewResp.ID)
	assert.Equal(t, reviewPayload.Code, reviewResp.Code)

	// Since analysis is async, we need to wait and fetch again
	time.Sleep(100 * time.Millisecond)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/reviews/"+reviewResp.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &reviewResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, reviewResp.SecurityIssues) // e.g. SQL Injection found by fake analyzer

	// 4. Get User Profile
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
