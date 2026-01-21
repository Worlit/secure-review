package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/middleware"
)

// GitHubHandler handles GitHub OAuth endpoints
type GitHubHandler struct {
	githubAuthService domain.GitHubAuthService
}

// NewGitHubHandler creates a new GitHubHandler
func NewGitHubHandler(githubAuthService domain.GitHubAuthService) *GitHubHandler {
	return &GitHubHandler{
		githubAuthService: githubAuthService,
	}
}

// GetAuthURL returns the GitHub OAuth authorization URL
// GET /api/auth/github
func (h *GitHubHandler) GetAuthURL(c *gin.Context) {
	state := generateState()
	url := h.githubAuthService.GetAuthURL(state)
	c.JSON(http.StatusOK, gin.H{
		"url":   url,
		"state": state,
	})
}

// Callback handles the GitHub OAuth callback
// GET /api/auth/github/callback
func (h *GitHubHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization code is required",
		})
		return
	}

	response, err := h.githubAuthService.AuthenticateOrCreate(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to authenticate with GitHub: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// LinkAccount links GitHub account to existing user
// POST /api/auth/github/link
func (h *GitHubHandler) LinkAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := h.githubAuthService.LinkAccount(c.Request.Context(), userID, req.Code)
	if err != nil {
		if err == domain.ErrGitHubAlreadyLinked {
			c.JSON(http.StatusConflict, gin.H{
				"error": "GitHub account is already linked to another user",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to link GitHub account: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "GitHub account linked successfully",
	})
}

// UnlinkAccount unlinks GitHub account from user
// DELETE /api/auth/github/link
func (h *GitHubHandler) UnlinkAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	err := h.githubAuthService.UnlinkAccount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unlink GitHub account: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "GitHub account unlinked successfully",
	})
}

// ListRepositories lists GitHub repositories for the authenticated user
// GET /api/v1/github/repos
func (h *GitHubHandler) ListRepositories(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	repos, err := h.githubAuthService.ListRepositories(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list repositories: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, repos)
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
