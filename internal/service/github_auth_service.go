package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	googleGithub "github.com/google/go-github/v69/github"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/secure-review/internal/domain"
)

var _ domain.GitHubAuthService = (*GitHubAuthServiceImpl)(nil)

// GitHubAuthServiceImpl implements the GitHubAuthService interface
type GitHubAuthServiceImpl struct {
	oauth2Config   *oauth2.Config
	userRepo       domain.UserRepository
	tokenGenerator *JWTTokenGenerator
}

// NewGitHubAuthService creates a new GitHubAuthServiceImpl
func NewGitHubAuthService(
	clientID, clientSecret, redirectURL string,
	userRepo domain.UserRepository,
	tokenGenerator *JWTTokenGenerator,
) *GitHubAuthServiceImpl {
	return &GitHubAuthServiceImpl{
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"user:email", "read:user", "repo"},
			Endpoint:     github.Endpoint,
		},
		userRepo:       userRepo,
		tokenGenerator: tokenGenerator,
	}
}

// GetAuthURL returns the GitHub OAuth authorization URL
func (s *GitHubAuthServiceImpl) GetAuthURL(state string) string {
	return s.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for an access token
func (s *GitHubAuthServiceImpl) ExchangeCode(ctx context.Context, code string) (string, error) {
	token, err := s.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	return token.AccessToken, nil
}

// GetUser fetches the GitHub user info using an access token
func (s *GitHubAuthServiceImpl) GetUser(ctx context.Context, accessToken string) (*domain.GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var ghUser domain.GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, err
	}

	// If email is empty, fetch from emails endpoint
	if ghUser.Email == "" {
		email, err := s.fetchPrimaryEmail(ctx, accessToken)
		if err == nil && email != "" {
			ghUser.Email = email
		}
	}

	return &ghUser, nil
}

func (s *GitHubAuthServiceImpl) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", nil
}

// AuthenticateOrCreate authenticates with GitHub and creates/updates user
func (s *GitHubAuthServiceImpl) AuthenticateOrCreate(ctx context.Context, code string) (*domain.AuthResponse, error) {
	accessToken, err := s.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	ghUser, err := s.GetUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// Try to find user by GitHub ID
	user, err := s.userRepo.GetByGitHubID(ctx, ghUser.ID)
	if err == nil && user != nil {
		// User exists, update token and generate session token
		// Always update key details on login
		user.GitHubLogin = &ghUser.Login
		user.GitHubAccessToken = &accessToken
		if ghUser.AvatarURL != "" {
			user.AvatarURL = &ghUser.AvatarURL
		}
		_ = s.userRepo.Update(ctx, user)

		token, err := s.tokenGenerator.GenerateToken(user.ID)
		if err != nil {
			return nil, err
		}
		return &domain.AuthResponse{
			Token: token,
			User:  user.ToResponse(),
		}, nil
	}

	// Try to find user by email
	if ghUser.Email != "" {
		user, err = s.userRepo.GetByEmail(ctx, ghUser.Email)
		if err == nil && user != nil {
			// Link GitHub account to existing user
			user.GitHubID = &ghUser.ID
			user.GitHubLogin = &ghUser.Login
			user.GitHubAccessToken = &accessToken
			if ghUser.AvatarURL != "" {
				user.AvatarURL = &ghUser.AvatarURL
			}
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, err
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
	}

	// Create new user
	username := ghUser.Login
	if ghUser.Name != "" {
		username = ghUser.Name
	}

	email := ghUser.Email
	if email == "" {
		email = fmt.Sprintf("%s@github.local", ghUser.Login)
	}

	user = &domain.User{
		ID:                uuid.New(),
		Email:             email,
		Username:          username,
		GitHubID:          &ghUser.ID,
		GitHubLogin:       &ghUser.Login,
		AvatarURL:         &ghUser.AvatarURL,
		GitHubAccessToken: &accessToken,
		IsActive:          true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
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

// LinkAccount links a GitHub account to an existing user
func (s *GitHubAuthServiceImpl) LinkAccount(ctx context.Context, userID uuid.UUID, code string) error {
	accessToken, err := s.ExchangeCode(ctx, code)
	if err != nil {
		return err
	}

	ghUser, err := s.GetUser(ctx, accessToken)
	if err != nil {
		return err
	}

	// Check if GitHub account is already linked to another user
	existingUser, err := s.userRepo.GetByGitHubID(ctx, ghUser.ID)
	if err == nil && existingUser != nil && existingUser.ID != userID {
		return domain.ErrGitHubAlreadyLinked
	}

	return s.userRepo.LinkGitHub(ctx, userID, &domain.LinkGitHubInput{
		GitHubID:          ghUser.ID,
		GitHubLogin:       ghUser.Login,
		AvatarURL:         ghUser.AvatarURL,
		GitHubAccessToken: accessToken,
	})
}

// UnlinkAccount removes the GitHub link from a user account
func (s *GitHubAuthServiceImpl) UnlinkAccount(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Ensure user has a password before unlinking
	if user.PasswordHash == "" {
		return fmt.Errorf("cannot unlink github: no password set")
	}

	return s.userRepo.UnlinkGitHub(ctx, userID)
}

// ListRepositories lists repositories for the authenticated user
func (s *GitHubAuthServiceImpl) ListRepositories(ctx context.Context, userID uuid.UUID) ([]domain.Repository, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if user.GitHubAccessToken == nil || *user.GitHubAccessToken == "" {
		return nil, fmt.Errorf("GitHub account not linked or token missing")
	}

	client := googleGithub.NewClient(nil).WithAuthToken(*user.GitHubAccessToken)

	opts := &googleGithub.RepositoryListByAuthenticatedUserOptions{
		ListOptions: googleGithub.ListOptions{PerPage: 100},
		Sort:        "updated",
	}

	var allRepos []*googleGithub.Repository
	for {
		repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repositories: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	result := make([]domain.Repository, len(allRepos))
	for i, r := range allRepos {
		result[i] = domain.Repository{
			ID:          r.GetID(),
			Name:        r.GetName(),
			FullName:    r.GetFullName(),
			HTMLURL:     r.GetHTMLURL(),
			Description: r.GetDescription(),
			Language:    r.GetLanguage(),
			Private:     r.GetPrivate(),
		}
	}

	return result, nil
}
