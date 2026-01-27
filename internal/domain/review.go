package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReviewStatus represents the status of a code review
type ReviewStatus string

const (
	ReviewStatusPending    ReviewStatus = "pending"
	ReviewStatusProcessing ReviewStatus = "processing"
	ReviewStatusCompleted  ReviewStatus = "completed"
	ReviewStatusFailed     ReviewStatus = "failed"
)

// SecuritySeverity represents the severity level of a security issue
type SecuritySeverity string

const (
	SeverityCritical SecuritySeverity = "critical"
	SeverityHigh     SecuritySeverity = "high"
	SeverityMedium   SecuritySeverity = "medium"
	SeverityLow      SecuritySeverity = "low"
	SeverityInfo     SecuritySeverity = "info"
)

// CodeReview represents a code review request and its results
type CodeReview struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	Title        string       `json:"title"`
	Code         string       `json:"code"`
	Language     string       `json:"language"`
	Status       ReviewStatus `json:"status"`
	Result       *string      `json:"result,omitempty"`
	CustomPrompt *string      `json:"custom_prompt,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}

// SecurityIssue represents a security vulnerability found in code
type SecurityIssue struct {
	ID          uuid.UUID        `json:"id"`
	ReviewID    uuid.UUID        `json:"review_id"`
	Severity    SecuritySeverity `json:"severity"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	LineStart   *int             `json:"line_start,omitempty"`
	LineEnd     *int             `json:"line_end,omitempty"`
	Suggestion  string           `json:"suggestion"`
	CWE         *string          `json:"cwe,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}

// CreateReviewInput represents input for creating a new code review
type CreateReviewInput struct {
	Title        string  `json:"title" binding:"required,min=1,max=255"`
	Code         *string `json:"code,omitempty"`
	Language     string  `json:"language"`
	RepoOwner    *string `json:"repo_owner,omitempty"`
	RepoName     *string `json:"repo_name,omitempty"`
	RepoBranch   *string `json:"repo_branch,omitempty"`
	CustomPrompt *string `json:"custom_prompt,omitempty"`
}

// ReviewResponse represents the response for a code review
type ReviewResponse struct {
	ID             uuid.UUID       `json:"id"`
	UserID         uuid.UUID       `json:"user_id"`
	Title          string          `json:"title"`
	Code           string          `json:"code"`
	Language       string          `json:"language"`
	Status         ReviewStatus    `json:"status"`
	Result         *string         `json:"result,omitempty"`
	CustomPrompt   *string         `json:"custom_prompt,omitempty"`
	SecurityIssues []SecurityIssue `json:"security_issues,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}

// ToResponse converts CodeReview to ReviewResponse
func (r *CodeReview) ToResponse(issues []SecurityIssue) *ReviewResponse {
	return &ReviewResponse{
		ID:             r.ID,
		UserID:         r.UserID,
		Title:          r.Title,
		Code:           r.Code,
		Language:       r.Language,
		Status:         r.Status,
		Result:         r.Result,
		CustomPrompt:   r.CustomPrompt,
		SecurityIssues: issues,
		CreatedAt:      r.CreatedAt,
		CompletedAt:    r.CompletedAt,
	}
}

// ReviewListResponse represents a paginated list of reviews
type ReviewListResponse struct {
	Reviews    []ReviewResponse `json:"reviews"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// AnalysisRequest represents the request to OpenAI for code analysis
type AnalysisRequest struct {
	Code         string  `json:"code"`
	Language     string  `json:"language"`
	CustomPrompt *string `json:"custom_prompt,omitempty"`
}

// AnalysisResult represents the result from OpenAI code analysis
type AnalysisResult struct {
	Summary        string               `json:"summary"`
	SecurityIssues []SecurityIssueInput `json:"security_issues"`
	Suggestions    []string             `json:"suggestions"`
	OverallScore   int                  `json:"overall_score"`
}

// SecurityIssueInput represents input for creating a security issue
type SecurityIssueInput struct {
	Severity    SecuritySeverity `json:"severity"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	LineStart   *int             `json:"line_start,omitempty"`
	LineEnd     *int             `json:"line_end,omitempty"`
	Suggestion  string           `json:"suggestion"`
	CWE         *string          `json:"cwe,omitempty"`
}
