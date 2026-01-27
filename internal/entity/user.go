package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user entity with GORM tags - аналог @Entity() в TypeORM
type User struct {
	// Primary key с автогенерацией UUID - аналог @PrimaryGeneratedColumn("uuid")
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Unique email - аналог @Column({ unique: true })
	Email string `gorm:"size:255;uniqueIndex;not null" json:"email"`

	// Username - аналог @Column()
	Username string `gorm:"size:100;not null" json:"username"`

	// Password hash - не отдаём в JSON - аналог @Column({ select: false })
	PasswordHash string `gorm:"size:255" json:"-"`

	// GitHub integration - nullable columns - аналог @Column({ nullable: true })
	GitHubID          *int64  `gorm:"column:github_id;index" json:"github_id,omitempty"`
	GitHubLogin       *string `gorm:"column:github_login;size:100" json:"github_login,omitempty"`
	AvatarURL         *string `gorm:"column:avatar_url;size:500" json:"avatar_url,omitempty"`
	GitHubAccessToken *string `gorm:"column:github_access_token;size:255" json:"-"`

	// Active flag - аналог @Column({ default: true })
	IsActive bool `gorm:"default:true" json:"is_active"`

	// Timestamps - аналог @CreateDateColumn() и @UpdateDateColumn()
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete - аналог @DeleteDateColumn()

	// Relations - аналог @OneToMany(() => CodeReview, review => review.user)
	Reviews []CodeReview `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"reviews,omitempty"`
}

// TableName returns the table name for GORM - аналог @Entity('users')
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook - аналог @BeforeInsert() в TypeORM
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserResponse represents the API response for user data
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	GitHubLogin *string   `json:"github_login,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToResponse converts User entity to UserResponse DTO
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		GitHubLogin: u.GitHubLogin,
		AvatarURL:   u.AvatarURL,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
	}
}

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Username  *string `json:"username,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// LinkGitHubInput represents input for linking GitHub account
type LinkGitHubInput struct {
	GitHubID          int64  `json:"github_id"`
	GitHubLogin       string `json:"github_login"`
	AvatarURL         string `json:"avatar_url"`
	GitHubAccessToken string `json:"github_access_token"`
}
