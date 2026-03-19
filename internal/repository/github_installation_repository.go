package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/entity"
)

type GitHubInstallationRepositoryAdapter struct {
	db *gorm.DB
}

func NewGitHubInstallationRepositoryAdapter(db *gorm.DB) domain.GitHubInstallationRepository {
	return &GitHubInstallationRepositoryAdapter{db: db}
}

func (r *GitHubInstallationRepositoryAdapter) Create(ctx context.Context, domainInstallation *domain.GitHubInstallation) error {
	installation := &entity.GitHubInstallation{
		ID:             domainInstallation.ID,
		InstallationID: domainInstallation.InstallationID,
		AccountID:      domainInstallation.AccountID,
		AccountLogin:   domainInstallation.AccountLogin,
		AccountType:    domainInstallation.AccountType,
		UserID:         domainInstallation.UserID,
	}
	if installation.ID == uuid.Nil {
		installation.ID = uuid.New()
	}

	err := r.db.WithContext(ctx).Create(installation).Error
	if err == nil {
		domainInstallation.ID = installation.ID
		domainInstallation.CreatedAt = installation.CreatedAt
		domainInstallation.UpdatedAt = installation.UpdatedAt
	}
	return err
}

func (r *GitHubInstallationRepositoryAdapter) GetByInstallationID(ctx context.Context, installationID int64) (*domain.GitHubInstallation, error) {
	var installation entity.GitHubInstallation
	err := r.db.WithContext(ctx).Where("installation_id = ?", installationID).First(&installation).Error
	if err != nil {
		return nil, err
	}
	return toDomainInstallation(&installation), nil
}

func (r *GitHubInstallationRepositoryAdapter) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.GitHubInstallation, error) {
	var installation entity.GitHubInstallation
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&installation).Error
	if err != nil {
		return nil, err
	}
	return toDomainInstallation(&installation), nil
}

func (r *GitHubInstallationRepositoryAdapter) GetByAccountID(ctx context.Context, accountID int64) (*domain.GitHubInstallation, error) {
	var installation entity.GitHubInstallation
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&installation).Error
	if err != nil {
		return nil, err
	}
	return toDomainInstallation(&installation), nil
}

func (r *GitHubInstallationRepositoryAdapter) Update(ctx context.Context, domainInstallation *domain.GitHubInstallation) error {
	installation := &entity.GitHubInstallation{
		ID:             domainInstallation.ID,
		InstallationID: domainInstallation.InstallationID,
		AccountID:      domainInstallation.AccountID,
		AccountLogin:   domainInstallation.AccountLogin,
		AccountType:    domainInstallation.AccountType,
		UserID:         domainInstallation.UserID,
	}
	return r.db.WithContext(ctx).Model(&entity.GitHubInstallation{}).Where("id = ?", installation.ID).Updates(installation).Error
}

func (r *GitHubInstallationRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.GitHubInstallation{}, "id = ?", id).Error
}

func (r *GitHubInstallationRepositoryAdapter) DeleteByInstallationID(ctx context.Context, installationID int64) error {
	return r.db.WithContext(ctx).Where("installation_id = ?", installationID).Delete(&entity.GitHubInstallation{}).Error
}

func toDomainInstallation(entity *entity.GitHubInstallation) *domain.GitHubInstallation {
	return &domain.GitHubInstallation{
		ID:             entity.ID,
		InstallationID: entity.InstallationID,
		AccountID:      entity.AccountID,
		AccountLogin:   entity.AccountLogin,
		AccountType:    entity.AccountType,
		UserID:         entity.UserID,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}
}
