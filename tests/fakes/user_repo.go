package fakes

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/secure-review/internal/domain"
)

type FakeUserRepository struct {
	mu    sync.Mutex
	users map[uuid.UUID]*domain.User
}

func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *FakeUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	for _, u := range r.users {
		if u.Email == user.Email {
			return errors.New("user already exists")
		}
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	r.users[user.ID] = user
	return nil
}

func (r *FakeUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *FakeUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *FakeUserRepository) GetByGitHubID(ctx context.Context, githubID int64) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.GitHubID != nil && *u.GitHubID == githubID {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *FakeUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	r.users[user.ID] = user
	return nil
}

func (r *FakeUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.users, id)
	return nil
}

func (r *FakeUserRepository) LinkGitHub(ctx context.Context, userID uuid.UUID, input *domain.LinkGitHubInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	user.GitHubID = &input.GitHubID
	user.GitHubLogin = &input.GitHubLogin
	user.GitHubAccessToken = &input.GitHubAccessToken

	return nil
}

func (r *FakeUserRepository) UnlinkGitHub(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	user.GitHubID = nil
	user.GitHubLogin = nil
	user.GitHubAccessToken = nil
	return nil
}
