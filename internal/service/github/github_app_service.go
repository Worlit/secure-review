package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	googleGithub "github.com/google/go-github/v69/github"
	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
)

type GitHubAppServiceImpl struct {
	appID            int64
	privateKey       []byte
	webhookSecret    []byte
	installationRepo domain.GitHubInstallationRepository
	userRepo         domain.UserRepository
}

func NewGitHubAppService(
	appID int64,
	privateKey string,
	webhookSecret string,
	installationRepo domain.GitHubInstallationRepository,
	userRepo domain.UserRepository,
) *GitHubAppServiceImpl {
	return &GitHubAppServiceImpl{
		appID:            appID,
		privateKey:       []byte(privateKey),
		webhookSecret:    []byte(webhookSecret),
		installationRepo: installationRepo,
		userRepo:         userRepo,
	}
}

func (s *GitHubAppServiceImpl) HandleWebhook(ctx context.Context, payload []byte, eventType string) error {
	event, err := googleGithub.ParseWebHook(eventType, payload)
	if err != nil {
		return err
	}

	switch e := event.(type) {
	case *googleGithub.InstallationEvent:
		return s.handleInstallationEvent(ctx, e)
	}

	return nil
}

func (s *GitHubAppServiceImpl) handleInstallationEvent(ctx context.Context, event *googleGithub.InstallationEvent) error {
	action := event.GetAction()
	installationID := event.GetInstallation().GetID()

	switch action {
	case "created", "unsuspend":
		account := event.GetInstallation().GetAccount()
		sender := event.GetSender()

		// Try to find user by sender ID (the user who installed the app)
		user, err := s.userRepo.GetByGitHubID(ctx, sender.GetID())
		var userID *uuid.UUID
		if err == nil && user != nil {
			userID = &user.ID
		}

		installation := &domain.GitHubInstallation{
			InstallationID: installationID,
			AccountID:      account.GetID(),
			AccountLogin:   account.GetLogin(),
			AccountType:    account.GetType(),
			UserID:         userID,
		}

		// Check if exists
		existing, err := s.installationRepo.GetByInstallationID(ctx, installationID)
		if err == nil && existing != nil {
			installation.ID = existing.ID
			return s.installationRepo.Update(ctx, installation)
		}
		return s.installationRepo.Create(ctx, installation)

	case "deleted", "suspend":
		return s.installationRepo.DeleteByInstallationID(ctx, installationID)
	}

	return nil
}

func (s *GitHubAppServiceImpl) GetInstallationToken(ctx context.Context, installationID int64) (string, error) {
	jwtToken, err := s.generateJWT()
	if err != nil {
		return "", err
	}

	client := googleGithub.NewClient(nil)
	client.WithAuthToken(jwtToken)

	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return "", err
	}

	return token.GetToken(), nil
}

func (s *GitHubAppServiceImpl) GetClient(ctx context.Context, userID uuid.UUID) (*googleGithub.Client, error) {
	// Find installation for user
	// First check if user has direct installation
	installation, err := s.installationRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Fallback: check if user is same as account ID (GitHubID)
		// This handles cases where installation isn't explicitly linked in DB but matches GitHub ID
		user, uErr := s.userRepo.GetByID(ctx, userID)
		if uErr != nil || user.GitHubID == nil {
			return nil, fmt.Errorf("no linked GitHub installation found")
		}

		installation, err = s.installationRepo.GetByAccountID(ctx, *user.GitHubID)
		if err != nil {
			return nil, fmt.Errorf("no GitHub installation found for user")
		}
	}

	token, err := s.GetInstallationToken(ctx, installation.InstallationID)
	if err != nil {
		return nil, err
	}

	return googleGithub.NewClient(nil).WithAuthToken(token), nil
}

func (s *GitHubAppServiceImpl) generateJWT() (string, error) {
	// Parse PEM block
	block, _ := pem.Decode(s.privateKey)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing private key")
	}

	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return "", fmt.Errorf("failed to parse private key: %v (PKCS1), %v (PKCS8)", err, err2)
		}
		// Assert to RSA
		var ok bool
		parsedKey, ok = key.(*rsa.PrivateKey) // This needs RSA import, wait...
		if !ok {
			return "", fmt.Errorf("private key is not RSA")
		}
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": s.appID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(parsedKey)
}
