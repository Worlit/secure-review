package fakes

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/secure-review/internal/domain"
)

type FakeReviewRepository struct {
	mu      sync.Mutex
	reviews map[uuid.UUID]*domain.CodeReview
	issues  map[uuid.UUID][]domain.SecurityIssue
}

func NewFakeReviewRepository() *FakeReviewRepository {
	return &FakeReviewRepository{
		reviews: make(map[uuid.UUID]*domain.CodeReview),
		issues:  make(map[uuid.UUID][]domain.SecurityIssue),
	}
}

func (r *FakeReviewRepository) Create(ctx context.Context, review *domain.CodeReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	if review.CreatedAt.IsZero() {
		review.CreatedAt = time.Now()
	}
	r.reviews[review.ID] = review
	return nil
}

func (r *FakeReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CodeReview, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	review, ok := r.reviews[id]
	if !ok {
		return nil, errors.New("review not found")
	}
	return review, nil
}

func (r *FakeReviewRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.CodeReview, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []domain.CodeReview
	for _, review := range r.reviews {
		if review.UserID == userID {
			result = append(result, *review)
		}
	}
	// Pagination logic omitted for simplicity in fake, returning all
	return result, len(result), nil
}

func (r *FakeReviewRepository) Update(ctx context.Context, review *domain.CodeReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.reviews[review.ID]; !ok {
		return errors.New("review not found")
	}
	review.UpdatedAt = time.Now()
	r.reviews[review.ID] = review
	return nil
}

func (r *FakeReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.reviews, id)
	return nil
}

func (r *FakeReviewRepository) CreateSecurityIssue(ctx context.Context, issue *domain.SecurityIssue) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if issue.ID == uuid.Nil {
		issue.ID = uuid.New()
	}
	r.issues[issue.ReviewID] = append(r.issues[issue.ReviewID], *issue)
	return nil
}

func (r *FakeReviewRepository) GetSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) ([]domain.SecurityIssue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.issues[reviewID], nil
}

func (r *FakeReviewRepository) DeleteSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.issues, reviewID)
	return nil
}
