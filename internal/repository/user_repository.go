package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/secure-review/internal/domain"
)

// PostgresUserRepository implements domain.UserRepository for PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create creates a new user in the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, github_id, github_login, avatar_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.GitHubID,
		user.GitHubLogin,
		user.AvatarURL,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetByID retrieves a user by their ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, github_id, github_login, avatar_url, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.GitHubID,
		&user.GitHubLogin,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by their email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, github_id, github_login, avatar_url, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.GitHubID,
		&user.GitHubLogin,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByGitHubID retrieves a user by their GitHub ID
func (r *PostgresUserRepository) GetByGitHubID(ctx context.Context, githubID int64) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, github_id, github_login, avatar_url, is_active, created_at, updated_at
		FROM users
		WHERE github_id = $1 AND is_active = true
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, githubID).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.GitHubID,
		&user.GitHubLogin,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, username = $3, password_hash = $4, github_id = $5, 
		    github_login = $6, avatar_url = $7, is_active = $8, updated_at = $9
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.GitHubID,
		user.GitHubLogin,
		user.AvatarURL,
		user.IsActive,
		user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete soft deletes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET is_active = false, updated_at = $2 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// LinkGitHub links a GitHub account to a user
func (r *PostgresUserRepository) LinkGitHub(ctx context.Context, userID uuid.UUID, input *domain.LinkGitHubInput) error {
	existingUser, err := r.GetByGitHubID(ctx, input.GitHubID)
	if err == nil && existingUser.ID != userID {
		return domain.ErrGitHubAlreadyLinked
	}

	query := `
		UPDATE users
		SET github_id = $2, github_login = $3, avatar_url = $4, updated_at = $5
		WHERE id = $1 AND is_active = true
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		userID,
		input.GitHubID,
		input.GitHubLogin,
		input.AvatarURL,
		time.Now(),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// UnlinkGitHub removes GitHub account link from user
func (r *PostgresUserRepository) UnlinkGitHub(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET github_id = NULL, github_login = NULL, updated_at = $2 WHERE id = $1 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
