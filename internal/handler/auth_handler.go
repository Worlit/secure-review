package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/middleware"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService domain.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService domain.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.CreateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "User with this email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register user",
		})
		return
	}

	// Set access token in cookie
	c.SetCookie("access_token", response.Token, 3600*24, "/", "", false, false)

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to login",
		})
		return
	}

	// Set access token in cookie
	c.SetCookie("access_token", response.Token, 3600*24, "/", "", false, false)

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	token, err := h.authService.RefreshToken(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed to refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// ChangePassword handles password change
// POST /api/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid old password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to change password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}
