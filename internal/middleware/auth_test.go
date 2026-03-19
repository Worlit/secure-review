package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/secure-review/internal/domain"
)

type mockAuthService struct {
	validTokens map[string]uuid.UUID
}

func newMockAuthService() *mockAuthService {
	return &mockAuthService{
		validTokens: make(map[string]uuid.UUID),
	}
}

func (s *mockAuthService) Register(ctx context.Context, input *domain.CreateUserInput) (*domain.AuthResponse, error) {
	return nil, nil
}

func (s *mockAuthService) Login(ctx context.Context, input *domain.LoginInput) (*domain.AuthResponse, error) {
	return nil, nil
}

func (s *mockAuthService) ValidateToken(token string) (uuid.UUID, error) {
	if userID, ok := s.validTokens[token]; ok {
		return userID, nil
	}
	return uuid.Nil, errors.New("invalid token")
}

func (s *mockAuthService) RefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return "", nil
}

func (s *mockAuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	return nil
}

func (s *mockAuthService) AddToken(token string, userID uuid.UUID) {
	s.validTokens[token] = userID
}

func setupRouter(middleware gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware)
	r.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if exists {
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		} else {
			c.JSON(http.StatusOK, gin.H{"user_id": nil})
		}
	})
	return r
}

func TestRequireAuth(t *testing.T) {
	authService := newMockAuthService()
	userID := uuid.New()
	validToken := "valid-token-123"
	authService.AddToken(validToken, userID)

	authMiddleware := NewAuthMiddleware(authService)

	t.Run("valid token passes auth", func(t *testing.T) {
		r := setupRouter(authMiddleware.RequireAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), userID.String())
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		r := setupRouter(authMiddleware.RequireAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token format returns 401", func(t *testing.T) {
		r := setupRouter(authMiddleware.RequireAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		r := setupRouter(authMiddleware.RequireAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("empty bearer token returns 401", func(t *testing.T) {
		r := setupRouter(authMiddleware.RequireAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid token from cookie passes auth", func(t *testing.T) {
		r := setupRouter(authMiddleware.RequireAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: validToken})
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), userID.String())
	})
}

func TestOptionalAuth(t *testing.T) {
	authService := newMockAuthService()
	userID := uuid.New()
	validToken := "valid-token-456"
	authService.AddToken(validToken, userID)

	authMiddleware := NewAuthMiddleware(authService)

	t.Run("valid token sets userID", func(t *testing.T) {
		r := setupRouter(authMiddleware.OptionalAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), userID.String())
	})

	t.Run("missing header continues without userID", func(t *testing.T) {
		r := setupRouter(authMiddleware.OptionalAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "null")
	})

	t.Run("invalid token continues without userID", func(t *testing.T) {
		r := setupRouter(authMiddleware.OptionalAuth())

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetUserID(t *testing.T) {
	t.Run("returns userID if set", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		expectedID := uuid.New()
		c.Set("userID", expectedID)

		userID, exists := GetUserID(c)

		require.True(t, exists)
		assert.Equal(t, expectedID, userID)
	})

	t.Run("returns false if not set", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userID, exists := GetUserID(c)

		assert.False(t, exists)
		assert.Equal(t, uuid.Nil, userID)
	})

	t.Run("returns false if wrong type", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", "not-a-uuid")

		userID, exists := GetUserID(c)

		assert.False(t, exists)
		assert.Equal(t, uuid.Nil, userID)
	})
}
