package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/entity"
)

// UserRepositoryAdapter adapts UserRepository to domain.UserRepository interface
type UserRepositoryAdapter struct {
	repo *UserRepository
}

// NewUserRepositoryAdapter creates a new adapter
func NewUserRepositoryAdapter(db *gorm.DB) domain.UserRepository {
	return &UserRepositoryAdapter{
		repo: NewUserRepository(db),
	}
}

// Create creates a new user
func (a *UserRepositoryAdapter) Create(ctx context.Context, user *domain.User) error {
	entityUser := domainUserToEntity(user)
	if err := a.repo.Create(ctx, entityUser); err != nil {
		return err
	}
	user.ID = entityUser.ID
	user.CreatedAt = entityUser.CreatedAt
	user.UpdatedAt = entityUser.UpdatedAt
	return nil
}

// GetByID returns a user by ID
func (a *UserRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	entityUser, err := a.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return entityUserToDomain(entityUser), nil
}

// GetByEmail returns a user by email
func (a *UserRepositoryAdapter) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	entityUser, err := a.repo.FindByEmail(ctx, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return entityUserToDomain(entityUser), nil
}

// GetByGitHubID returns a user by GitHub ID
func (a *UserRepositoryAdapter) GetByGitHubID(ctx context.Context, githubID int64) (*domain.User, error) {
	entityUser, err := a.repo.FindByGitHubID(ctx, githubID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return entityUserToDomain(entityUser), nil
}

// Update updates a user
func (a *UserRepositoryAdapter) Update(ctx context.Context, user *domain.User) error {
	entityUser := domainUserToEntity(user)
	return a.repo.Update(ctx, entityUser)
}

// Delete soft-deletes a user
func (a *UserRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.repo.Delete(ctx, id)
}

// LinkGitHub links GitHub account to user
func (a *UserRepositoryAdapter) LinkGitHub(ctx context.Context, userID uuid.UUID, input *domain.LinkGitHubInput) error {
	entityInput := &entity.LinkGitHubInput{
		GitHubID:          input.GitHubID,
		GitHubLogin:       input.GitHubLogin,
		AvatarURL:         input.AvatarURL,
		GitHubAccessToken: input.GitHubAccessToken,
	}
	return a.repo.LinkGitHub(ctx, userID, entityInput)
}

// UnlinkGitHub unlinks GitHub account from user
func (a *UserRepositoryAdapter) UnlinkGitHub(ctx context.Context, userID uuid.UUID) error {
	return a.repo.UnlinkGitHub(ctx, userID)
}

func domainUserToEntity(user *domain.User) *entity.User {
	return &entity.User{
		ID:                user.ID,
		Email:             user.Email,
		Username:          user.Username,
		PasswordHash:      user.PasswordHash,
		GitHubID:          user.GitHubID,
		GitHubLogin:       user.GitHubLogin,
		AvatarURL:         user.AvatarURL,
		GitHubAccessToken: user.GitHubAccessToken,
		IsActive:          user.IsActive,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}

func entityUserToDomain(user *entity.User) *domain.User {
	return &domain.User{
		ID:                user.ID,
		Email:             user.Email,
		Username:          user.Username,
		PasswordHash:      user.PasswordHash,
		GitHubID:          user.GitHubID,
		GitHubLogin:       user.GitHubLogin,
		AvatarURL:         user.AvatarURL,
		GitHubAccessToken: user.GitHubAccessToken,
		IsActive:          user.IsActive,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}
