package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/entity"
)

// ReviewRepositoryAdapter adapts ReviewRepository to domain.ReviewRepository interface
type ReviewRepositoryAdapter struct {
	repo *ReviewRepository
}

// NewReviewRepositoryAdapter creates a new adapter
func NewReviewRepositoryAdapter(db *gorm.DB) domain.ReviewRepository {
	return &ReviewRepositoryAdapter{
		repo: NewReviewRepository(db),
	}
}

// Create creates a new code review
func (a *ReviewRepositoryAdapter) Create(ctx context.Context, review *domain.CodeReview) error {
	entityReview := domainReviewToEntity(review)
	if err := a.repo.Create(ctx, entityReview); err != nil {
		return err
	}
	review.ID = entityReview.ID
	review.CreatedAt = entityReview.CreatedAt
	review.UpdatedAt = entityReview.UpdatedAt
	return nil
}

// GetByID returns a review by ID
func (a *ReviewRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.CodeReview, error) {
	entityReview, err := a.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrReviewNotFound
		}
		return nil, err
	}
	return entityReviewToDomain(entityReview), nil
}

// GetByUserID returns paginated reviews by user ID
func (a *ReviewRepositoryAdapter) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.CodeReview, int, error) {
	entityReviews, total, err := a.repo.FindByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	reviews := make([]domain.CodeReview, len(entityReviews))
	for i, entityReview := range entityReviews {
		reviews[i] = *entityReviewToDomain(&entityReview)
	}

	return reviews, int(total), nil
}

// Update updates a review
func (a *ReviewRepositoryAdapter) Update(ctx context.Context, review *domain.CodeReview) error {
	entityReview := domainReviewToEntity(review)
	return a.repo.Update(ctx, entityReview)
}

// Delete soft-deletes a review
func (a *ReviewRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.repo.Delete(ctx, id)
}

// CreateSecurityIssue creates a security issue
func (a *ReviewRepositoryAdapter) CreateSecurityIssue(ctx context.Context, issue *domain.SecurityIssue) error {
	entityIssue := domainIssueToEntity(issue)
	if err := a.repo.CreateSecurityIssue(ctx, entityIssue); err != nil {
		return err
	}
	issue.ID = entityIssue.ID
	issue.CreatedAt = entityIssue.CreatedAt
	return nil
}

// GetSecurityIssuesByReviewID returns all security issues for a review
func (a *ReviewRepositoryAdapter) GetSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) ([]domain.SecurityIssue, error) {
	entityIssues, err := a.repo.FindSecurityIssuesByReviewID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	issues := make([]domain.SecurityIssue, len(entityIssues))
	for i, entityIssue := range entityIssues {
		issues[i] = *entityIssueToDomain(&entityIssue)
	}

	return issues, nil
}

// DeleteSecurityIssuesByReviewID deletes all security issues for a review
func (a *ReviewRepositoryAdapter) DeleteSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) error {
	return a.repo.DeleteSecurityIssuesByReviewID(ctx, reviewID)
}

func domainReviewToEntity(review *domain.CodeReview) *entity.CodeReview {
	return &entity.CodeReview{
		ID:          review.ID,
		UserID:      review.UserID,
		Title:       review.Title,
		Code:        review.Code,
		Language:    review.Language,
		Status:      entity.ReviewStatus(review.Status),
		Result:      review.Result,
		CreatedAt:   review.CreatedAt,
		UpdatedAt:   review.UpdatedAt,
		CompletedAt: review.CompletedAt,
	}
}

func entityReviewToDomain(review *entity.CodeReview) *domain.CodeReview {
	return &domain.CodeReview{
		ID:          review.ID,
		UserID:      review.UserID,
		Title:       review.Title,
		Code:        review.Code,
		Language:    review.Language,
		Status:      domain.ReviewStatus(review.Status),
		Result:      review.Result,
		CreatedAt:   review.CreatedAt,
		UpdatedAt:   review.UpdatedAt,
		CompletedAt: review.CompletedAt,
	}
}

func domainIssueToEntity(issue *domain.SecurityIssue) *entity.SecurityIssue {
	return &entity.SecurityIssue{
		ID:          issue.ID,
		ReviewID:    issue.ReviewID,
		Severity:    entity.SecuritySeverity(issue.Severity),
		Title:       issue.Title,
		Description: issue.Description,
		LineStart:   issue.LineStart,
		LineEnd:     issue.LineEnd,
		Suggestion:  issue.Suggestion,
		CWE:         issue.CWE,
		CreatedAt:   issue.CreatedAt,
	}
}

func entityIssueToDomain(issue *entity.SecurityIssue) *domain.SecurityIssue {
	return &domain.SecurityIssue{
		ID:          issue.ID,
		ReviewID:    issue.ReviewID,
		Severity:    domain.SecuritySeverity(issue.Severity),
		Title:       issue.Title,
		Description: issue.Description,
		LineStart:   issue.LineStart,
		LineEnd:     issue.LineEnd,
		Suggestion:  issue.Suggestion,
		CWE:         issue.CWE,
		CreatedAt:   issue.CreatedAt,
	}
}
