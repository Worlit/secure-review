package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
)

var _ domain.AuthService = (*AuthServiceImpl)(nil)

// AuthServiceImpl implements the AuthService interface
type AuthServiceImpl struct {
	userRepo       domain.UserRepository
	passwordHasher domain.PasswordHasher
	tokenGenerator domain.TokenGenerator
}

// NewAuthService creates a new AuthServiceImpl
func NewAuthService(
	userRepo domain.UserRepository,
	passwordHasher domain.PasswordHasher,
	tokenGenerator domain.TokenGenerator,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
	}
}

// Register creates a new user account
func (s *AuthServiceImpl) Register(ctx context.Context, input *domain.CreateUserInput) (*domain.AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.tokenGenerator.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// Login authenticates a user and returns a token
func (s *AuthServiceImpl) Login(ctx context.Context, input *domain.LoginInput) (*domain.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := s.passwordHasher.Compare(input.Password, user.PasswordHash); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := s.tokenGenerator.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// ValidateToken validates a token and returns the user ID
func (s *AuthServiceImpl) ValidateToken(token string) (uuid.UUID, error) {
	return s.tokenGenerator.ValidateToken(token)
}

// RefreshToken generates a new token for a user
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", domain.ErrUserNotFound
	}

	return s.tokenGenerator.GenerateToken(userID)
}

// ChangePassword changes the user's password
func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	if err := s.passwordHasher.Compare(oldPassword, user.PasswordHash); err != nil {
		return domain.ErrInvalidCredentials
	}

	hashedPassword, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	return s.userRepo.Update(ctx, user)
}
