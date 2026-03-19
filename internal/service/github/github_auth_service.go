package github

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	googleGithub "github.com/google/go-github/v69/github"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/logger"
)

var _ domain.GitHubAuthService = (*GitHubAuthServiceImpl)(nil)

// GitHubAuthServiceImpl implements the GitHubAuthService interface
type GitHubAuthServiceImpl struct {
	oauth2Config   *oauth2.Config
	userRepo       domain.UserRepository
	tokenGenerator domain.TokenGenerator
	appService     domain.GitHubAppService
}

// NewGitHubAuthService creates a new GitHubAuthServiceImpl
func NewGitHubAuthService(
	clientID, clientSecret, redirectURL string,
	userRepo domain.UserRepository,
	tokenGenerator domain.TokenGenerator,
	appService domain.GitHubAppService,
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
		appService:     appService,
	}
}

// GetAuthURL returns the GitHub OAuth authorization URL
func (s *GitHubAuthServiceImpl) GetAuthURL(state string) string {
	return s.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for an access token
func (s *GitHubAuthServiceImpl) ExchangeCode(ctx context.Context, code string) (string, error) {
	logger.Log.Info("Exchanging GitHub code for token")
	token, err := s.oauth2Config.Exchange(ctx, code)
	if err != nil {
		logger.Log.Error("Failed to exchange GitHub code", "error", err)
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	return token.AccessToken, nil
}

// GetUser fetches the GitHub user info using an access token
func (s *GitHubAuthServiceImpl) GetUser(ctx context.Context, accessToken string) (*domain.GitHubUser, error) {
	logger.Log.Info("Fetching GitHub user info")
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("Failed to fetch GitHub user", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error("GitHub API error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("github api error: %s", resp.Status)
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
	var allRepos []*googleGithub.Repository

	// 1. Try App Installation
	appClient, err := s.appService.GetClient(ctx, userID)
	if err == nil {
		// Use App API
		opts := &googleGithub.ListOptions{PerPage: 100}
		for {
			repos, resp, err := appClient.Apps.ListRepos(ctx, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch installation repositories: %w", err)
			}
			allRepos = append(allRepos, repos.Repositories...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	} else {
		// 2. Fallback to OAuth
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

// ListBranches lists branches for a repository
func (s *GitHubAuthServiceImpl) ListBranches(ctx context.Context, userID uuid.UUID, owner, repo string) ([]string, error) {
	var client *googleGithub.Client

	appClient, err := s.appService.GetClient(ctx, userID)
	if err == nil {
		client = appClient
	} else {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, domain.ErrUserNotFound
		}

		if user.GitHubAccessToken == nil || *user.GitHubAccessToken == "" {
			return nil, fmt.Errorf("GitHub account not linked or token missing")
		}

		client = googleGithub.NewClient(nil).WithAuthToken(*user.GitHubAccessToken)
	}

	opts := &googleGithub.BranchListOptions{
		ListOptions: googleGithub.ListOptions{PerPage: 100},
	}

	var allBranches []*googleGithub.Branch
	for {
		branches, resp, err := client.Repositories.ListBranches(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch branches: %w", err)
		}
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	var result []string
	for _, b := range allBranches {
		if b.Name != nil {
			result = append(result, *b.Name)
		}
	}

	return result, nil
}

// GetRepositoryContent fetches the repository content as a single string
func (s *GitHubAuthServiceImpl) GetRepositoryContent(ctx context.Context, userID uuid.UUID, owner, repo, ref string) (string, error) {
	var client *googleGithub.Client

	appClient, err := s.appService.GetClient(ctx, userID)
	if err == nil {
		client = appClient
	} else {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return "", domain.ErrUserNotFound
		}

		if user.GitHubAccessToken == nil || *user.GitHubAccessToken == "" {
			return "", fmt.Errorf("GitHub account not linked or token missing")
		}
		client = googleGithub.NewClient(nil).WithAuthToken(*user.GitHubAccessToken)
	}

	// Get archive link
	opts := &googleGithub.RepositoryContentGetOptions{
		Ref: ref,
	}
	url, resp, err := client.Repositories.GetArchiveLink(ctx, owner, repo, googleGithub.Zipball, opts, 5)
	if err != nil {
		// handle redirect manually if needed, but GetArchiveLink with true should follow or return url
		return "", fmt.Errorf("failed to get archive link: %w", err)
	}
	// resp.Location is usually where it redirects, but client might return it in url
	downloadURL := url.String()
	if downloadURL == "" && resp != nil {
		downloadURL = resp.Header.Get("Location")
	}
	if downloadURL == "" {
		return "", fmt.Errorf("failed to determine download URL")
	}

	// Download zip
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{}
	dlResp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download archive: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download archive: status %d", dlResp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(dlResp.Body)
	if err != nil {
		return "", err
	}

	// Unzip and filter
	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to open zip archive: %w", err)
	}

	var sb strings.Builder
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		// Skip unwanted files and directories
		if shouldSkipFile(file.Name) {
			continue
		}

		// Limit file size (skip large files, e.g. > 100KB)
		if file.FileInfo().Size() > 100*1024 {
			continue
		}

		f, err := file.Open()
		if err != nil {
			continue
		}

		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			continue
		}

		// Basic check if file is text
		if !isText(content) {
			continue
		}

		sb.WriteString(fmt.Sprintf("\n--- File: %s ---\n", file.Name))
		sb.WriteString(string(content))
		sb.WriteString("\n")
	}

	if sb.Len() == 0 {
		return "", fmt.Errorf("no suitable source files found in repository")
	}

	return sb.String(), nil
}

func shouldSkipFile(path string) bool {
	// Simple filters
	if strings.Contains(path, "node_modules/") ||
		strings.Contains(path, ".git/") ||
		strings.Contains(path, "vendor/") ||
		strings.Contains(path, ".idea/") ||
		strings.Contains(path, ".vscode/") ||
		strings.Contains(path, "dist/") ||
		strings.Contains(path, "build/") ||
		strings.Contains(path, "coverage/") ||
		strings.Contains(path, "tmp/") ||
		strings.Contains(path, "__pycache__/") {
		return true
	}

	// Skip specific large or non-source files
	fileName := strings.ToLower(filepath.Base(path))
	if fileName == "package-lock.json" ||
		fileName == "yarn.lock" ||
		fileName == "pnpm-lock.yaml" ||
		fileName == "go.sum" ||
		fileName == "cargo.lock" ||
		strings.HasSuffix(fileName, ".map") ||
		strings.HasSuffix(fileName, ".min.js") ||
		strings.HasSuffix(fileName, ".min.css") {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))
	allowedExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".py": true, ".java": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true, ".rb": true,
		".php": true, ".cs": true, ".rs": true, ".swift": true, ".kt": true,
		".html": true, ".css": true, ".json": true, ".yaml": true, ".yml": true,
		".sql": true, ".md": true,
	}
	return !allowedExts[ext]
}

func isText(b []byte) bool {
	// Simple heuristic: check for null bytes
	if len(b) > 1024 {
		b = b[:1024]
	}
	for _, c := range b {
		if c == 0 {
			return false
		}
	}
	return true
}
