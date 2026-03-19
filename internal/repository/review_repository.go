package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/secure-review/internal/entity"
)

// ReviewRepository - аналог Repository<CodeReview> в TypeORM
type ReviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository creates a new ReviewRepository - аналог getRepository(CodeReview)
func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create creates a new code review - аналог repository.save()
func (r *ReviewRepository) Create(ctx context.Context, review *entity.CodeReview) error {
	return r.db.WithContext(ctx).Create(review).Error
}

// FindByID finds a review by ID - аналог repository.findOne({ where: { id } })
func (r *ReviewRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.CodeReview, error) {
	var review entity.CodeReview
	err := r.db.WithContext(ctx).First(&review, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// FindByIDWithIssues finds a review with security issues preloaded
// Аналог { relations: ['securityIssues'] } в TypeORM
func (r *ReviewRepository) FindByIDWithIssues(ctx context.Context, id uuid.UUID) (*entity.CodeReview, error) {
	var review entity.CodeReview
	err := r.db.WithContext(ctx).
		Preload("SecurityIssues").
		First(&review, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// FindByIDWithUserAndIssues finds a review with user and security issues
// Аналог { relations: ['user', 'securityIssues'] } в TypeORM
func (r *ReviewRepository) FindByIDWithUserAndIssues(ctx context.Context, id uuid.UUID) (*entity.CodeReview, error) {
	var review entity.CodeReview
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("SecurityIssues").
		First(&review, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// FindByUserID finds all reviews by user ID with pagination
// Аналог repository.findAndCount({ where: { userId }, skip, take, order })
func (r *ReviewRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]entity.CodeReview, int64, error) {
	var reviews []entity.CodeReview
	var total int64

	// Count total - аналог findAndCount
	err := r.db.WithContext(ctx).
		Model(&entity.CodeReview{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results with preload - аналог { relations, skip, take, order }
	offset := (page - 1) * pageSize
	err = r.db.WithContext(ctx).
		Preload("SecurityIssues").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&reviews).Error

	return reviews, total, err
}

// FindByUserIDAndStatus finds reviews by user ID and status
// Аналог repository.find({ where: { userId, status } })
func (r *ReviewRepository) FindByUserIDAndStatus(ctx context.Context, userID uuid.UUID, status entity.ReviewStatus) ([]entity.CodeReview, error) {
	var reviews []entity.CodeReview
	err := r.db.WithContext(ctx).
		Preload("SecurityIssues").
		Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

// Update updates a review - аналог repository.save()
func (r *ReviewRepository) Update(ctx context.Context, review *entity.CodeReview) error {
	return r.db.WithContext(ctx).Save(review).Error
}

// UpdateStatus updates only the status field - аналог repository.update(id, { status })
func (r *ReviewRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReviewStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.CodeReview{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateFields updates specific fields - аналог repository.update(id, { ...fields })
func (r *ReviewRepository) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&entity.CodeReview{}).
		Where("id = ?", id).
		Updates(fields).Error
}

// Delete soft-deletes a review - аналог repository.softDelete(id)
func (r *ReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.CodeReview{}, "id = ?", id).Error
}

// HardDelete permanently deletes a review - аналог repository.delete(id)
func (r *ReviewRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.CodeReview{}, "id = ?", id).Error
}

// CountByUserID counts reviews for a user - аналог repository.count({ where: { userId } })
func (r *ReviewRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.CodeReview{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CreateSecurityIssue creates a security issue - аналог repository.save() для SecurityIssue
func (r *ReviewRepository) CreateSecurityIssue(ctx context.Context, issue *entity.SecurityIssue) error {
	return r.db.WithContext(ctx).Create(issue).Error
}

// CreateSecurityIssues creates multiple security issues - аналог repository.save([...])
func (r *ReviewRepository) CreateSecurityIssues(ctx context.Context, issues []entity.SecurityIssue) error {
	if len(issues) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&issues).Error
}

// FindSecurityIssuesByReviewID finds all security issues for a review
// Аналог repository.find({ where: { reviewId } })
func (r *ReviewRepository) FindSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) ([]entity.SecurityIssue, error) {
	var issues []entity.SecurityIssue
	err := r.db.WithContext(ctx).
		Where("review_id = ?", reviewID).
		Order("severity ASC, created_at ASC").
		Find(&issues).Error
	return issues, err
}

// DeleteSecurityIssuesByReviewID deletes all security issues for a review
// Аналог repository.delete({ reviewId })
func (r *ReviewRepository) DeleteSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("review_id = ?", reviewID).
		Delete(&entity.SecurityIssue{}).Error
}

// FindRecentByUserID finds recent reviews for a user
// Аналог repository.find({ where: { userId }, take: limit, order: { createdAt: 'DESC' } })
func (r *ReviewRepository) FindRecentByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]entity.CodeReview, error) {
	var reviews []entity.CodeReview
	err := r.db.WithContext(ctx).
		Preload("SecurityIssues").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&reviews).Error
	return reviews, err
}

// FindPendingReviews finds all pending reviews (for background processing)
// Аналог repository.find({ where: { status: 'pending' } })
func (r *ReviewRepository) FindPendingReviews(ctx context.Context) ([]entity.CodeReview, error) {
	var reviews []entity.CodeReview
	err := r.db.WithContext(ctx).
		Where("status = ?", entity.ReviewStatusPending).
		Order("created_at ASC").
		Find(&reviews).Error
	return reviews, err
}
