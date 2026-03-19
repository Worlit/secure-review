package auth

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/secure-review/internal/domain"
)

var _ domain.PasswordHasher = (*BcryptPasswordHasher)(nil)

// BcryptPasswordHasher implements PasswordHasher using bcrypt
type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher creates a new BcryptPasswordHasher
func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// Hash creates a bcrypt hash of a password
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare compares a password with a hash
func (h *BcryptPasswordHasher) Compare(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
