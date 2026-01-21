package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                uuid.UUID `json:"id"`
	Email             string    `json:"email"`
	Username          string    `json:"username"`
	PasswordHash      string    `json:"-"`
	GitHubID          *int64    `json:"github_id,omitempty"`
	GitHubLogin       *string   `json:"github_login,omitempty"`
	AvatarURL         *string   `json:"avatar_url,omitempty"`
	GitHubAccessToken *string   `json:"-"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreateUserInput represents input for creating a new user
type CreateUserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginInput represents input for user login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserInput represents input for updating user profile
type UpdateUserInput struct {
	Username  *string `json:"username,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// LinkGitHubInput represents input for linking GitHub account
type LinkGitHubInput struct {
	GitHubID          int64  `json:"github_id"`
	GitHubLogin       string `json:"github_login"`
	AvatarURL         string `json:"avatar_url"`
	GitHubAccessToken string `json:"github_access_token"`
}

// UserResponse represents the response for user data
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	GitHubLogin *string   `json:"github_login,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		GitHubLogin: u.GitHubLogin,
		AvatarURL:   u.AvatarURL,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
	}
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

// GitHubUser represents GitHub user data from OAuth
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
}
