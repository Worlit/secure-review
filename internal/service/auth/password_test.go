package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcryptPasswordHasher_Hash(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	hash, err := hasher.Hash("password123")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "password123", hash)
}

func TestBcryptPasswordHasher_Compare_Success(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	hash, _ := hasher.Hash("password123")
	err := hasher.Compare("password123", hash)
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_Compare_WrongPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	hash, _ := hasher.Hash("password123")
	err := hasher.Compare("wrongpassword", hash)
	assert.Error(t, err)
}

func TestBcryptPasswordHasher_Compare_InvalidHash(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	err := hasher.Compare("password123", "invalid-hash")
	assert.Error(t, err)
}

func TestBcryptPasswordHasher_DifferentHashes(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	hash1, _ := hasher.Hash("password123")
	hash2, _ := hasher.Hash("password123")
	assert.NotEqual(t, hash1, hash2)
	assert.NoError(t, hasher.Compare("password123", hash1))
	assert.NoError(t, hasher.Compare("password123", hash2))
}
