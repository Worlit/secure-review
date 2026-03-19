package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/secure-review/internal/entity"
)

// UserRepository - аналог Repository<User> в TypeORM
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository - аналог getRepository(User) в TypeORM
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user - аналог repository.save() в TypeORM
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID finds a user by ID - аналог repository.findOne({ where: { id } }) в TypeORM
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByIDWithReviews finds a user with reviews preloaded - аналог { relations: ['reviews'] }
func (r *UserRepository) FindByIDWithReviews(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Reviews").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email - аналог repository.findOne({ where: { email } })
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGitHubID finds a user by GitHub ID - аналог repository.findOne({ where: { githubId } })
func (r *UserRepository) FindByGitHubID(ctx context.Context, githubID int64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("github_id = ?", githubID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user - аналог repository.save() для существующей entity
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateFields updates specific fields - аналог repository.update(id, { ...fields })
func (r *UserRepository) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Updates(fields).Error
}

// Delete soft-deletes a user - аналог repository.softDelete(id)
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

// HardDelete permanently deletes a user - аналог repository.delete(id)
func (r *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.User{}, "id = ?", id).Error
}

// FindAll returns all users - аналог repository.find()
func (r *UserRepository) FindAll(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

// FindAllActive returns all active users - аналог repository.find({ where: { isActive: true } })
func (r *UserRepository) FindAllActive(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&users).Error
	return users, err
}

// Count returns count of users - аналог repository.count()
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&count).Error
	return count, err
}

// Exists checks if user exists - аналог repository.exist({ where: { id } })
func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail checks if user with email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// LinkGitHub links GitHub account to user
func (r *UserRepository) LinkGitHub(ctx context.Context, userID uuid.UUID, input *entity.LinkGitHubInput) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"github_id":           input.GitHubID,
		"github_login":        input.GitHubLogin,
		"avatar_url":          input.AvatarURL,
		"github_access_token": input.GitHubAccessToken,
	}).Error
}

// UnlinkGitHub unlinks GitHub account from user
func (r *UserRepository) UnlinkGitHub(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"github_id":           nil,
		"github_login":        nil,
		"avatar_url":          nil,
		"github_access_token": nil,
	}).Error
}
