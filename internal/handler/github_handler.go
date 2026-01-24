package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/middleware"
)

// GitHubHandler handles GitHub OAuth endpoints
type GitHubHandler struct {
	githubAuthService domain.GitHubAuthService
	tokenGenerator    domain.TokenGenerator
	frontendURL       string
}

// NewGitHubHandler creates a new GitHubHandler
func NewGitHubHandler(githubAuthService domain.GitHubAuthService, tokenGenerator domain.TokenGenerator, frontendURL string) *GitHubHandler {
	return &GitHubHandler{
		githubAuthService: githubAuthService,
		tokenGenerator:    tokenGenerator,
		frontendURL:       frontendURL,
	}
}

// GetAuthURL returns the GitHub OAuth authorization URL
// GET /api/auth/github
func (h *GitHubHandler) GetAuthURL(c *gin.Context) {
	state := generateState()

	// Check if user is already authenticated to enable linking
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			userID, err := h.tokenGenerator.ValidateToken(parts[1])
			if err == nil {
				// Set cookie to identify user during callback which happens on the same domain
				c.SetCookie("github_link_user", userID.String(), 300, "/", "", true, true)
			}
		}
	}

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
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=no_code")
		return
	}

	var token string

	// Check if we are in linking mode
	linkUserIDStr, err := c.Cookie("github_link_user")
	if err == nil && linkUserIDStr != "" {
		// Linking mode
		userID, err := uuid.Parse(linkUserIDStr)
		if err == nil {
			err = h.githubAuthService.LinkAccount(c.Request.Context(), userID, code)
			if err != nil {
				c.Redirect(http.StatusFound, h.frontendURL+"/login?error=link_failed")
				return
			}
			// Clean up linking cookie
			c.SetCookie("github_link_user", "", -1, "/", "", false, true)

			// Generate a new token for the user to refresh session
			token, err = h.tokenGenerator.GenerateToken(userID)
			if err != nil {
				c.Redirect(http.StatusFound, h.frontendURL+"/login?error=token_generation_failed")
				return
			}
		} else {
			// Invalid user ID in cookie
			c.Redirect(http.StatusFound, h.frontendURL+"/login?error=invalid_link_state")
			return
		}
	} else {
		// Login/Register mode
		response, err := h.githubAuthService.AuthenticateOrCreate(c.Request.Context(), code)
		if err != nil {
			c.Redirect(http.StatusFound, h.frontendURL+"/login?error=auth_failed")
			return
		}
		token = response.Token
	}

	// Set auth cookie
	c.SetCookie("access_token", token, 3600*24, "/", "", false, false)

	// Redirect to frontend with token (keeping it in URL for compatibility, but cookie is primary now)
	c.Redirect(http.StatusFound, h.frontendURL+"/login?token="+token)
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
