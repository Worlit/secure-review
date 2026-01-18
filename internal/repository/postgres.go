package repository

import (
	"database/sql"
	"fmt"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// RunMigrations runs the database migrations
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(50) NOT NULL,
			password_hash VARCHAR(255),
			github_id BIGINT UNIQUE,
			github_login VARCHAR(255),
			avatar_url TEXT,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_github_id ON users(github_id)`,
		`CREATE TABLE IF NOT EXISTS code_reviews (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			code TEXT NOT NULL,
			language VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			result TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			completed_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_code_reviews_user_id ON code_reviews(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_code_reviews_status ON code_reviews(status)`,
		`CREATE TABLE IF NOT EXISTS security_issues (
			id UUID PRIMARY KEY,
			review_id UUID NOT NULL REFERENCES code_reviews(id) ON DELETE CASCADE,
			severity VARCHAR(20) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			line_start INTEGER,
			line_end INTEGER,
			suggestion TEXT NOT NULL,
			cwe VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_security_issues_review_id ON security_issues(review_id)`,
		`CREATE INDEX IF NOT EXISTS idx_security_issues_severity ON security_issues(severity)`,
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			return fmt.Errorf("failed to run migration: %w\nQuery: %s", err, migration)
		}
	}

	return nil
}
