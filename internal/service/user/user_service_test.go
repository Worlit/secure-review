package user

import (
	"context"
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

func TestUserService_GetByID_Success(t *testing.T) {
	repo := newMockUserRepository()
	userID := uuid.New()
	repo.users[userID] = &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: true,
	}

	svc := NewUserService(repo)

	result, err := svc.GetByID(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "test@example.com", result.Email)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	result, err := svc.GetByID(context.Background(), uuid.New())

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
	assert.Nil(t, result)
}

func TestUserService_GetByEmail_Success(t *testing.T) {
	repo := newMockUserRepository()
	userID := uuid.New()
	repo.users[userID] = &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Username: "testuser",
	}

	svc := NewUserService(repo)

	result, err := svc.GetByEmail(context.Background(), "test@example.com")

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", result.Email)
}

func TestUserService_Update_Success(t *testing.T) {
	repo := newMockUserRepository()
	userID := uuid.New()
	repo.users[userID] = &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Username: "oldusername",
	}

	svc := NewUserService(repo)

	newUsername := "newusername"
	input := &domain.UpdateUserInput{Username: &newUsername}

	result, err := svc.Update(context.Background(), userID, input)

	require.NoError(t, err)
	assert.Equal(t, "newusername", result.Username)
}

func TestUserService_Update_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	newUsername := "newusername"
	input := &domain.UpdateUserInput{Username: &newUsername}

	result, err := svc.Update(context.Background(), uuid.New(), input)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
	assert.Nil(t, result)
}

func TestUserService_Delete_Success(t *testing.T) {
	repo := newMockUserRepository()
	userID := uuid.New()
	repo.users[userID] = &domain.User{ID: userID, Email: "test@example.com"}

	svc := NewUserService(repo)

	err := svc.Delete(context.Background(), userID)

	require.NoError(t, err)
	_, err = repo.GetByID(context.Background(), userID)
	assert.Error(t, err)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	err := svc.Delete(context.Background(), uuid.New())

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
}
