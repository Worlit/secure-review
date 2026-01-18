package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
)

var _ domain.UserService = (*UserServiceImpl)(nil)

// UserServiceImpl implements the UserService interface
type UserServiceImpl struct {
	userRepo domain.UserRepository
}

// NewUserService creates a new UserServiceImpl
func NewUserService(userRepo domain.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

// GetByID returns a user by ID
func (s *UserServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetByEmail returns a user by email
func (s *UserServiceImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// Update updates a user's profile
func (s *UserServiceImpl) Update(ctx context.Context, userID uuid.UUID, input *domain.UpdateUserInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete deletes a user account
func (s *UserServiceImpl) Delete(ctx context.Context, userID uuid.UUID) error {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, userID)
}
