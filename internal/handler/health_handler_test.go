package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler("1.0.0")

	r := gin.New()
	r.GET("/health", handler.Health)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestHealthHandler_Ready(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler("1.0.0")

	r := gin.New()
	r.GET("/ready", handler.Ready)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
}

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler("2.0.0")

	assert.NotNil(t, handler)
	assert.Equal(t, "2.0.0", handler.version)
}
