package fakes

import (
	"context"

	"github.com/google/uuid"
	"github.com/secure-review/internal/domain"
)

type FakeGitHubAuthService struct{}

func NewFakeGitHubAuthService() *FakeGitHubAuthService {
	return &FakeGitHubAuthService{}
}

func (s *FakeGitHubAuthService) GetAuthURL(state string) string {
	return "http://fake-auth-url"
}

func (s *FakeGitHubAuthService) ExchangeCode(ctx context.Context, code string) (string, error) {
	return "fake-token", nil
}

func (s *FakeGitHubAuthService) GetUser(ctx context.Context, accessToken string) (*domain.GitHubUser, error) {
	return &domain.GitHubUser{Login: "testuser", ID: 123}, nil
}

func (s *FakeGitHubAuthService) AuthenticateOrCreate(ctx context.Context, code string) (*domain.AuthResponse, error) {
	return nil, nil
}

func (s *FakeGitHubAuthService) LinkAccount(ctx context.Context, userID uuid.UUID, code string) error {
	return nil
}

func (s *FakeGitHubAuthService) UnlinkAccount(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *FakeGitHubAuthService) ListRepositories(ctx context.Context, userID uuid.UUID) ([]domain.Repository, error) {
	return []domain.Repository{}, nil
}

func (s *FakeGitHubAuthService) ListBranches(ctx context.Context, userID uuid.UUID, owner, repo string) ([]string, error) {
	return []string{"main", "develop"}, nil
}

func (s *FakeGitHubAuthService) GetRepositoryContent(ctx context.Context, userID uuid.UUID, owner, repo, ref string) (string, error) {
	return "fake content", nil
}
