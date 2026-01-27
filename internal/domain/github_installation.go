package domain

import (
	"time"

	"github.com/google/uuid"
)

// GitHubInstallation represents a GitHub App installation in domain
type GitHubInstallation struct {
	ID             uuid.UUID
	InstallationID int64
	AccountID      int64
	AccountLogin   string
	AccountType    string
	UserID         *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
