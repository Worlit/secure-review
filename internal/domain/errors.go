package domain

import "errors"

// Common errors used across the application
var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is inactive")

	// GitHub errors
	ErrGitHubAlreadyLinked = errors.New("github account already linked to another user")
	ErrGitHubNotLinked     = errors.New("github account not linked to any user")

	// Review errors
	ErrReviewNotFound     = errors.New("code review not found")
	ErrReviewAccessDenied = errors.New("access denied to this review")

	// Authentication errors
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrTokenRequired = errors.New("authentication token required")
	ErrUnauthorized  = errors.New("unauthorized access")

	// Validation errors
	ErrInvalidInput = errors.New("invalid input data")

	// OpenAI errors
	ErrOpenAIUnavailable = errors.New("openai service unavailable")
	ErrAnalysisFailed    = errors.New("code analysis failed")
)
