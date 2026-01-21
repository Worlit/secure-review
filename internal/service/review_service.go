package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
)

var _ domain.ReviewService = (*ReviewServiceImpl)(nil)

// ReviewServiceImpl implements the ReviewService interface
type ReviewServiceImpl struct {
	reviewRepo   domain.ReviewRepository
	codeAnalyzer domain.CodeAnalyzer
}

// NewReviewService creates a new ReviewServiceImpl
func NewReviewService(reviewRepo domain.ReviewRepository, codeAnalyzer domain.CodeAnalyzer) *ReviewServiceImpl {
	return &ReviewServiceImpl{
		reviewRepo:   reviewRepo,
		codeAnalyzer: codeAnalyzer,
	}
}

// Create creates a new code review
func (s *ReviewServiceImpl) Create(ctx context.Context, userID uuid.UUID, input *domain.CreateReviewInput) (*domain.ReviewResponse, error) {
	review := &domain.CodeReview{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     input.Title,
		Code:      input.Code,
		Language:  input.Language,
		Status:    domain.ReviewStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	// Start async analysis
	go s.analyzeCode(context.Background(), review)

	return review.ToResponse(nil), nil
}

// GetByID returns a review by ID
func (s *ReviewServiceImpl) GetByID(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) (*domain.ReviewResponse, error) {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, domain.ErrReviewNotFound
	}

	if review.UserID != userID {
		return nil, domain.ErrReviewAccessDenied
	}

	issues, err := s.reviewRepo.GetSecurityIssuesByReviewID(ctx, reviewID)
	if err != nil {
		issues = nil
	}

	return review.ToResponse(issues), nil
}

// GetUserReviews returns paginated reviews for a user
func (s *ReviewServiceImpl) GetUserReviews(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.ReviewListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reviews, total, err := s.reviewRepo.GetByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.ReviewResponse, len(reviews))
	for i, review := range reviews {
		issues, _ := s.reviewRepo.GetSecurityIssuesByReviewID(ctx, review.ID)
		responses[i] = *review.ToResponse(issues)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &domain.ReviewListResponse{
		Reviews:    responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Delete deletes a review
func (s *ReviewServiceImpl) Delete(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) error {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return domain.ErrReviewNotFound
	}

	if review.UserID != userID {
		return domain.ErrReviewAccessDenied
	}

	// Delete security issues first
	if err := s.reviewRepo.DeleteSecurityIssuesByReviewID(ctx, reviewID); err != nil {
		return err
	}

	return s.reviewRepo.Delete(ctx, reviewID)
}

// ReanalyzeReview re-runs analysis on an existing review
func (s *ReviewServiceImpl) ReanalyzeReview(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) (*domain.ReviewResponse, error) {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, domain.ErrReviewNotFound
	}

	if review.UserID != userID {
		return nil, domain.ErrReviewAccessDenied
	}

	// Delete old issues
	if err := s.reviewRepo.DeleteSecurityIssuesByReviewID(ctx, reviewID); err != nil {
		return nil, err
	}

	// Reset status
	review.Status = domain.ReviewStatusPending
	review.Result = nil
	review.CompletedAt = nil
	review.UpdatedAt = time.Now()

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	// Start async analysis
	go s.analyzeCode(context.Background(), review)

	return review.ToResponse(nil), nil
}

func (s *ReviewServiceImpl) analyzeCode(ctx context.Context, review *domain.CodeReview) {
	review.Status = domain.ReviewStatusProcessing
	review.UpdatedAt = time.Now()
	_ = s.reviewRepo.Update(ctx, review)

	result, err := s.codeAnalyzer.AnalyzeCode(ctx, &domain.AnalysisRequest{
		Code:     review.Code,
		Language: review.Language,
	})

	if err != nil {
		review.Status = domain.ReviewStatusFailed
		errorMsg := err.Error()
		review.Result = &errorMsg
		review.UpdatedAt = time.Now()
		_ = s.reviewRepo.Update(ctx, review)
		return
	}

	// Save security issues
	for _, issue := range result.SecurityIssues {
		securityIssue := &domain.SecurityIssue{
			ID:          uuid.New(),
			ReviewID:    review.ID,
			Severity:    issue.Severity,
			Title:       issue.Title,
			Description: issue.Description,
			LineStart:   issue.LineStart,
			LineEnd:     issue.LineEnd,
			Suggestion:  issue.Suggestion,
			CWE:         issue.CWE,
			CreatedAt:   time.Now(),
		}
		_ = s.reviewRepo.CreateSecurityIssue(ctx, securityIssue)
	}

	review.Status = domain.ReviewStatusCompleted

	// Format the result to include score, summary and suggestions
	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("# Analysis Result\n\n"))
	resultBuilder.WriteString(fmt.Sprintf("**Overall Safe Score:** %d/100\n\n", result.OverallScore))

	resultBuilder.WriteString("## Summary\n")
	resultBuilder.WriteString(result.Summary)
	resultBuilder.WriteString("\n\n")

	if len(result.Suggestions) > 0 {
		resultBuilder.WriteString("## Code Quality Suggestions\n")
		for _, suggestion := range result.Suggestions {
			resultBuilder.WriteString(fmt.Sprintf("- %s\n", suggestion))
		}
	}

	resultString := resultBuilder.String()
	review.Result = &resultString

	now := time.Now()
	review.CompletedAt = &now
	review.UpdatedAt = now
	_ = s.reviewRepo.Update(ctx, review)
}
