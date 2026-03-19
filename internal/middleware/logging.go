package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/secure-review/internal/logger"
)

// Logger returns a logging middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Log.Info("Request",
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.Duration("latency", latency),
			slog.String("client_ip", clientIP),
			slog.String("method", method),
		)

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Log.Error("Request error",
					slog.String("error", e.Error()),
					slog.String("path", path),
				)
			}
		}
	}
}

// Recovery returns a recovery middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Log.Error("Panic recovered",
					slog.Any("error", err),
				)
				c.AbortWithStatusJSON(500, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}
