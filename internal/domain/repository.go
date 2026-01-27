package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByGitHubID(ctx context.Context, githubID int64) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	LinkGitHub(ctx context.Context, userID uuid.UUID, input *LinkGitHubInput) error
	UnlinkGitHub(ctx context.Context, userID uuid.UUID) error
}

// ReviewRepository defines the interface for code review data access
type ReviewRepository interface {
	Create(ctx context.Context, review *CodeReview) error
	GetByID(ctx context.Context, id uuid.UUID) (*CodeReview, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]CodeReview, int, error)
	Update(ctx context.Context, review *CodeReview) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateSecurityIssue(ctx context.Context, issue *SecurityIssue) error
	GetSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) ([]SecurityIssue, error)
	DeleteSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) error
}

// GitHubInstallationRepository defines the interface for GitHub installation data access
type GitHubInstallationRepository interface {
	Create(ctx context.Context, installation *GitHubInstallation) error
	GetByInstallationID(ctx context.Context, installationID int64) (*GitHubInstallation, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*GitHubInstallation, error)
	GetByAccountID(ctx context.Context, accountID int64) (*GitHubInstallation, error)
	Update(ctx context.Context, installation *GitHubInstallation) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByInstallationID(ctx context.Context, installationID int64) error
}
