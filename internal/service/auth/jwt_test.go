package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTTokenGenerator_GenerateToken(t *testing.T) {
	gen := NewJWTTokenGenerator("secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	token, err := gen.GenerateToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTTokenGenerator_ValidateToken(t *testing.T) {
	gen := NewJWTTokenGenerator("secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	token, _ := gen.GenerateToken(userID)
	resultID, err := gen.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, resultID)
}

func TestJWTTokenGenerator_InvalidToken(t *testing.T) {
	gen := NewJWTTokenGenerator("secret", time.Hour, 24*time.Hour)
	_, err := gen.ValidateToken("invalid")
	assert.Error(t, err)
}

func TestJWTTokenGenerator_WrongSecret(t *testing.T) {
	gen1 := NewJWTTokenGenerator("secret1", time.Hour, 24*time.Hour)
	gen2 := NewJWTTokenGenerator("secret2", time.Hour, 24*time.Hour)
	userID := uuid.New()
	token, _ := gen1.GenerateToken(userID)
	_, err := gen2.ValidateToken(token)
	assert.Error(t, err)
}

func TestJWTTokenGenerator_ExpiredToken(t *testing.T) {
	gen := NewJWTTokenGenerator("secret", time.Millisecond, 24*time.Hour)
	userID := uuid.New()
	token, _ := gen.GenerateToken(userID)
	time.Sleep(10 * time.Millisecond)
	_, err := gen.ValidateToken(token)
	assert.Error(t, err)
}
