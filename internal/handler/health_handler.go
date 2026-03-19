package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	version string
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		version: version,
	}
}

// Health returns the health status
// GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"version":   h.version,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready returns the readiness status
// GET /ready
func (h *HealthHandler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
