package fakes

import (
	"github.com/google/uuid"
)

type FakePasswordHasher struct{}

func NewFakePasswordHasher() *FakePasswordHasher {
	return &FakePasswordHasher{}
}

func (h *FakePasswordHasher) Hash(password string) (string, error) {
	return password, nil // No real hashing for tests
}

func (h *FakePasswordHasher) Compare(password, hash string) error {
	if password != hash {
		return &domainError{Message: "password mismatch"}
	}
	return nil
}

type domainError struct {
	Message string
}

func (e *domainError) Error() string {
	return e.Message
}

type FakeTokenGenerator struct{}

func NewFakeTokenGenerator() *FakeTokenGenerator {
	return &FakeTokenGenerator{}
}

func (t *FakeTokenGenerator) GenerateToken(userID uuid.UUID) (string, error) {
	return "test-token-" + userID.String(), nil
}

func (t *FakeTokenGenerator) ValidateToken(token string) (uuid.UUID, error) {
	// Extract simple user ID from mock token
	if len(token) > 11 && token[:11] == "test-token-" {
		idStr := token[11:]
		return uuid.Parse(idStr)
	}
	return uuid.Nil, &domainError{Message: "invalid token"}
}
