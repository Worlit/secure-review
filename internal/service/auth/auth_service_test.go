package auth

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/secure-review/internal/domain"
)

type mockUserRepository struct {
	mu    sync.Mutex
	users map[uuid.UUID]*domain.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{users: make(map[uuid.UUID]*domain.User)}
}

func (r *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email == user.Email {
			return domain.ErrUserAlreadyExists
		}
	}
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *mockUserRepository) GetByGitHubID(ctx context.Context, id int64) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}

func (r *mockUserRepository) LinkGitHub(ctx context.Context, userID uuid.UUID, input *domain.LinkGitHubInput) error {
	return nil
}

func (r *mockUserRepository) UnlinkGitHub(ctx context.Context, userID uuid.UUID) error {
	return nil
}

type mockPasswordHasher struct {
	shouldFail bool
}

func newMockPasswordHasher() *mockPasswordHasher {
	return &mockPasswordHasher{}
}

func (h *mockPasswordHasher) Hash(password string) (string, error) {
	if h.shouldFail {
		return "", errors.New("hash error")
	}
	return "hashed_" + password, nil
}

func (h *mockPasswordHasher) Compare(password, hash string) error {
	if hash == "hashed_"+password {
		return nil
	}
	return errors.New("password mismatch")
}

type mockTokenGenerator struct{}

func newMockTokenGenerator() *mockTokenGenerator {
	return &mockTokenGenerator{}
}

func (t *mockTokenGenerator) GenerateToken(userID uuid.UUID) (string, error) {
	return "token_" + userID.String(), nil
}

func (t *mockTokenGenerator) ValidateToken(token string) (uuid.UUID, error) {
	if len(token) > 6 && token[:6] == "token_" {
		return uuid.Parse(token[6:])
	}
	return uuid.Nil, errors.New("invalid token")
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := newMockUserRepository()
	hasher := newMockPasswordHasher()
	tokenGen := newMockTokenGenerator()
	svc := NewAuthService(repo, hasher, tokenGen)

	input := &domain.CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}

	result, err := svc.Register(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "test@example.com", result.User.Email)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	repo := newMockUserRepository()
	hasher := newMockPasswordHasher()
	tokenGen := newMockTokenGenerator()

	existingUser := &domain.User{ID: uuid.New(), Email: "test@example.com"}
	repo.users[existingUser.ID] = existingUser

	svc := NewAuthService(repo, hasher, tokenGen)

	input := &domain.CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}

	result, err := svc.Register(context.Background(), input)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserAlreadyExists, err)
	assert.Nil(t, result)
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := newMockUserRepository()
	hasher := newMockPasswordHasher()
	tokenGen := newMockTokenGenerator()

	userID := uuid.New()
	existingUser := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "hashed_password123",
		IsActive:     true,
	}
	repo.users[userID] = existingUser

	svc := NewAuthService(repo, hasher, tokenGen)

	input := &domain.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	result, err := svc.Login(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	repo := newMockUserRepository()
	hasher := newMockPasswordHasher()
	tokenGen := newMockTokenGenerator()

	svc := NewAuthService(repo, hasher, tokenGen)

	input := &domain.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	result, err := svc.Login(context.Background(), input)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
	assert.Nil(t, result)
}

func TestAuthService_ValidateToken(t *testing.T) {
	repo := newMockUserRepository()
	hasher := newMockPasswordHasher()
	tokenGen := newMockTokenGenerator()

	svc := NewAuthService(repo, hasher, tokenGen)

	userID := uuid.New()
	token := "token_" + userID.String()

	resultID, err := svc.ValidateToken(token)

	require.NoError(t, err)
	assert.Equal(t, userID, resultID)
}

func TestAuthService_RefreshToken(t *testing.T) {
	repo := newMockUserRepository()
	hasher := newMockPasswordHasher()
	tokenGen := newMockTokenGenerator()

	userID := uuid.New()
	repo.users[userID] = &domain.User{ID: userID, Email: "test@example.com"}

	svc := NewAuthService(repo, hasher, tokenGen)

	newToken, err := svc.RefreshToken(context.Background(), userID)

	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
}

func TestAuthService_ChangePassword(t *testing.T) {
	repo := newMockUserRepository()
	hasher := newMockPasswordHasher()
	tokenGen := newMockTokenGenerator()

	userID := uuid.New()
	repo.users[userID] = &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "hashed_oldpassword",
	}

	svc := NewAuthService(repo, hasher, tokenGen)

	err := svc.ChangePassword(context.Background(), userID, "oldpassword", "newpassword")

	require.NoError(t, err)
	updatedUser, _ := repo.GetByID(context.Background(), userID)
	assert.Equal(t, "hashed_newpassword", updatedUser.PasswordHash)
}
