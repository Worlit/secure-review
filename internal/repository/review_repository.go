package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
)

// PostgresReviewRepository implements domain.ReviewRepository for PostgreSQL
type PostgresReviewRepository struct {
	db *sql.DB
}

// NewPostgresReviewRepository creates a new PostgresReviewRepository
func NewPostgresReviewRepository(db *sql.DB) *PostgresReviewRepository {
	return &PostgresReviewRepository{db: db}
}

// Create creates a new code review in the database
func (r *PostgresReviewRepository) Create(ctx context.Context, review *domain.CodeReview) error {
	query := `
		INSERT INTO code_reviews (id, user_id, title, code, language, status, result, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	review.ID = uuid.New()
	review.CreatedAt = now
	review.UpdatedAt = now
	review.Status = domain.ReviewStatusPending

	_, err := r.db.ExecContext(
		ctx,
		query,
		review.ID,
		review.UserID,
		review.Title,
		review.Code,
		review.Language,
		review.Status,
		review.Result,
		review.CreatedAt,
		review.UpdatedAt,
		review.CompletedAt,
	)

	return err
}

// GetByID retrieves a code review by its ID
func (r *PostgresReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CodeReview, error) {
	query := `
		SELECT id, user_id, title, code, language, status, result, created_at, updated_at, completed_at
		FROM code_reviews
		WHERE id = $1
	`

	review := &domain.CodeReview{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&review.ID,
		&review.UserID,
		&review.Title,
		&review.Code,
		&review.Language,
		&review.Status,
		&review.Result,
		&review.CreatedAt,
		&review.UpdatedAt,
		&review.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrReviewNotFound
	}

	if err != nil {
		return nil, err
	}

	return review, nil
}

// GetByUserID retrieves all code reviews for a user with pagination
func (r *PostgresReviewRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.CodeReview, int, error) {
	countQuery := `SELECT COUNT(*) FROM code_reviews WHERE user_id = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, user_id, title, code, language, status, result, created_at, updated_at, completed_at
		FROM code_reviews
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []domain.CodeReview
	for rows.Next() {
		var review domain.CodeReview
		err := rows.Scan(
			&review.ID,
			&review.UserID,
			&review.Title,
			&review.Code,
			&review.Language,
			&review.Status,
			&review.Result,
			&review.CreatedAt,
			&review.UpdatedAt,
			&review.CompletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// Update updates an existing code review
func (r *PostgresReviewRepository) Update(ctx context.Context, review *domain.CodeReview) error {
	query := `
		UPDATE code_reviews
		SET title = $2, code = $3, language = $4, status = $5, result = $6, 
		    updated_at = $7, completed_at = $8
		WHERE id = $1
	`

	review.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		review.ID,
		review.Title,
		review.Code,
		review.Language,
		review.Status,
		review.Result,
		review.UpdatedAt,
		review.CompletedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrReviewNotFound
	}

	return nil
}

// Delete deletes a code review
func (r *PostgresReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.DeleteSecurityIssuesByReviewID(ctx, id)
	if err != nil {
		return err
	}

	query := `DELETE FROM code_reviews WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrReviewNotFound
	}

	return nil
}

// CreateSecurityIssue creates a new security issue for a review
func (r *PostgresReviewRepository) CreateSecurityIssue(ctx context.Context, issue *domain.SecurityIssue) error {
	query := `
		INSERT INTO security_issues (id, review_id, severity, title, description, line_start, line_end, suggestion, cwe, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	issue.ID = uuid.New()
	issue.CreatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		issue.ID,
		issue.ReviewID,
		issue.Severity,
		issue.Title,
		issue.Description,
		issue.LineStart,
		issue.LineEnd,
		issue.Suggestion,
		issue.CWE,
		issue.CreatedAt,
	)

	return err
}

// GetSecurityIssuesByReviewID retrieves all security issues for a review
func (r *PostgresReviewRepository) GetSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) ([]domain.SecurityIssue, error) {
	query := `
		SELECT id, review_id, severity, title, description, line_start, line_end, suggestion, cwe, created_at
		FROM security_issues
		WHERE review_id = $1
		ORDER BY 
			CASE severity 
				WHEN 'critical' THEN 1 
				WHEN 'high' THEN 2 
				WHEN 'medium' THEN 3 
				WHEN 'low' THEN 4 
				WHEN 'info' THEN 5 
			END
	`

	rows, err := r.db.QueryContext(ctx, query, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []domain.SecurityIssue
	for rows.Next() {
		var issue domain.SecurityIssue
		err := rows.Scan(
			&issue.ID,
			&issue.ReviewID,
			&issue.Severity,
			&issue.Title,
			&issue.Description,
			&issue.LineStart,
			&issue.LineEnd,
			&issue.Suggestion,
			&issue.CWE,
			&issue.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return issues, nil
}

// DeleteSecurityIssuesByReviewID deletes all security issues for a review
func (r *PostgresReviewRepository) DeleteSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) error {
	query := `DELETE FROM security_issues WHERE review_id = $1`
	_, err := r.db.ExecContext(ctx, query, reviewID)
	return err
}
