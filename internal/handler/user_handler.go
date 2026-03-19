package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/middleware"
)

// UserHandler handles user-related endpoints
type UserHandler struct {
	userService domain.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile returns the current user's profile
// GET /api/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user profile",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// UpdateProfile updates the current user's profile
// PUT /api/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req domain.UpdateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	user, err := h.userService.Update(c.Request.Context(), userID, &req)
	if err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// DeleteAccount deletes the current user's account
// DELETE /api/users/me
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	err := h.userService.Delete(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete account",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account deleted successfully",
	})
}
