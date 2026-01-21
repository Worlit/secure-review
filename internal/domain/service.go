package domain

import (
	"context"

	"github.com/google/uuid"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, input *CreateUserInput) (*AuthResponse, error)
	Login(ctx context.Context, input *LoginInput) (*AuthResponse, error)
	ValidateToken(token string) (uuid.UUID, error)
	RefreshToken(ctx context.Context, userID uuid.UUID) (string, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}

// GitHubAuthService defines the interface for GitHub OAuth operations
type GitHubAuthService interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUser(ctx context.Context, accessToken string) (*GitHubUser, error)
	AuthenticateOrCreate(ctx context.Context, code string) (*AuthResponse, error)
	LinkAccount(ctx context.Context, userID uuid.UUID, code string) error
	UnlinkAccount(ctx context.Context, userID uuid.UUID) error
	ListRepositories(ctx context.Context, userID uuid.UUID) ([]Repository, error)
}

// Repository represents a GitHub repository
type Repository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Private     bool   `json:"private"`
}

// UserService defines the interface for user operations
type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, userID uuid.UUID, input *UpdateUserInput) (*User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

// ReviewService defines the interface for code review operations
type ReviewService interface {
	Create(ctx context.Context, userID uuid.UUID, input *CreateReviewInput) (*ReviewResponse, error)
	GetByID(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) (*ReviewResponse, error)
	GetUserReviews(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ReviewListResponse, error)
	Delete(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) error
	ReanalyzeReview(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) (*ReviewResponse, error)
}

// CodeAnalyzer defines the interface for code analysis (OpenAI)
type CodeAnalyzer interface {
	AnalyzeCode(ctx context.Context, request *AnalysisRequest) (*AnalysisResult, error)
	AnalyzeSecurity(ctx context.Context, request *AnalysisRequest) ([]SecurityIssueInput, error)
}

// TokenGenerator defines the interface for JWT token generation
type TokenGenerator interface {
	GenerateToken(userID uuid.UUID) (string, error)
	ValidateToken(token string) (uuid.UUID, error)
}

// PasswordHasher defines the interface for password hashing
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password, hash string) error
}
