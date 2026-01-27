package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	googleGithub "github.com/google/go-github/v69/github"
	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/logger"
	"github.com/secure-review/internal/middleware"
)

// GitHubHandler handles GitHub OAuth endpoints
type GitHubHandler struct {
	githubAuthService domain.GitHubAuthService
	githubAppService  domain.GitHubAppService
	tokenGenerator    domain.TokenGenerator
	frontendURL       string
	webhookSecret     []byte
	isProduction      bool
}

// NewGitHubHandler creates a new GitHubHandler
func NewGitHubHandler(
	githubAuthService domain.GitHubAuthService,
	githubAppService domain.GitHubAppService,
	tokenGenerator domain.TokenGenerator,
	frontendURL string,
	webhookSecret string,
	isProduction bool,
) *GitHubHandler {
	return &GitHubHandler{
		githubAuthService: githubAuthService,
		githubAppService:  githubAppService,
		tokenGenerator:    tokenGenerator,
		frontendURL:       frontendURL,
		webhookSecret:     []byte(webhookSecret),
		isProduction:      isProduction,
	}
}

// GetAuthURL returns the GitHub OAuth authorization URL
// GET /api/auth/github
func (h *GitHubHandler) GetAuthURL(c *gin.Context) {
	state := generateState()

	// Check if user is already authenticated to enable linking
	var tokenString string
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	} else {
		// Try cookie
		cookie, err := c.Cookie("access_token")
		if err == nil {
			tokenString = cookie
		}
	}

	if tokenString != "" {
		userID, err := h.tokenGenerator.ValidateToken(tokenString)
		if err == nil {
			// Set cookie to identify user during callback which happens on the same domain
			c.SetCookie("github_link_user", userID.String(), 300, "/", "", false, true)
		}
	}

	url := h.githubAuthService.GetAuthURL(state)
	c.JSON(http.StatusOK, gin.H{
		"url":   url,
		"state": state,
	})
}

// Callback handles the GitHub OAuth callback
// POST /api/auth/github/callback
func (h *GitHubHandler) Callback(c *gin.Context) {
	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.githubAuthService.AuthenticateOrCreate(c.Request.Context(), req.Code)
	if err != nil {
		logger.Log.Error("GitHub authentication failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication failed: " + err.Error(),
		})
		return
	}

	// Set auth cookie
	if h.isProduction {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("access_token", response.Token, 3600*24, "/", "", h.isProduction, true)

	c.JSON(http.StatusOK, response)
}

// CallbackRedirect handles the GitHub OAuth callback via browser redirect (GET)
// GET /api/auth/github/callback
func (h *GitHubHandler) CallbackRedirect(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=no_code")
		return
	}

	// 1. ПРОВЕРКА НА ПРИВЯЗКУ (LINKING)
	// Читаем cookie, которую поставили в GetAuthURL
	linkUserID, err := c.Cookie("github_link_user")
	isLinking := err == nil && linkUserID != ""

	if isLinking {
		// Парсим ID пользователя
		userID, err := uuid.Parse(linkUserID)
		if err == nil {
			// Вызываем явную ПРИВЯЗКУ, передавая userID и code
			// Обратите внимание: метод LinkAccount должен быть доступен в интерфейсе сервиса
			err = h.githubAuthService.LinkAccount(c.Request.Context(), userID, code)

			// Очищаем cookie намерения
			c.SetCookie("github_link_user", "", -1, "/", "", false, true)

			if err != nil {
				// Если ошибка (например, этот GitHub уже занят)
				logger.Log.Error("GitHub link failed", "error", err)
				c.Redirect(http.StatusFound, h.frontendURL+"/profile?error=link_failed")
				return
			}

			// Успех -> Редирект в профиль со статусом
			c.Redirect(http.StatusFound, h.frontendURL+"/profile?status=github_linked")
			return
		}
	}

	// 2. ОБЫЧНЫЙ ВХОД (LOGIN / REGISTRATION)
	response, err := h.githubAuthService.AuthenticateOrCreate(c.Request.Context(), code)
	if err != nil {
		logger.Log.Error("GitHub authentication failed (GET)", "error", err)
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=auth_failed")
		return
	}

	// Сохраняем токен и редиректим на логин
	if h.isProduction {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("access_token", response.Token, 3600*24, "/", "", h.isProduction, true)

	// Важно передать token в URL для фронтенда
	c.Redirect(http.StatusFound, h.frontendURL+"/login?token="+response.Token)
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

// ListBranches lists branches for a repository
// GET /api/v1/github/repos/:owner/:repo/branches
func (h *GitHubHandler) ListBranches(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	branches, err := h.githubAuthService.ListBranches(c.Request.Context(), userID, owner, repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list branches: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, branches)
}

// Webhook handles GitHub App webhooks
// POST /api/v1/github/webhook
func (h *GitHubHandler) Webhook(c *gin.Context) {
	payload, err := googleGithub.ValidatePayload(c.Request, h.webhookSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	event := googleGithub.WebHookType(c.Request)
	if err := h.githubAppService.HandleWebhook(c.Request.Context(), payload, event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed"})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
