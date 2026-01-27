package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GitHubInstallation represents a GitHub App installation
type GitHubInstallation struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Installation ID from GitHub
	InstallationID int64 `gorm:"index;unique" json:"installation_id"`

	// GitHub Account ID (User or Org)
	AccountID    int64  `gorm:"index" json:"account_id"`
	AccountLogin string `gorm:"size:255" json:"account_login"`
	AccountType  string `gorm:"size:50" json:"account_type"` // "User" or "Organization"

	// Linked internal User (if mapped)
	UserID *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User   *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (GitHubInstallation) TableName() string {
	return "github_installations"
}
