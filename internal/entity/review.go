package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReviewStatus represents the status of a code review
type ReviewStatus string

const (
	ReviewStatusPending    ReviewStatus = "pending"
	ReviewStatusProcessing ReviewStatus = "processing"
	ReviewStatusCompleted  ReviewStatus = "completed"
	ReviewStatusFailed     ReviewStatus = "failed"
)

// SecuritySeverity represents the severity level of a security issue
type SecuritySeverity string

const (
	SeverityCritical SecuritySeverity = "critical"
	SeverityHigh     SecuritySeverity = "high"
	SeverityMedium   SecuritySeverity = "medium"
	SeverityLow      SecuritySeverity = "low"
	SeverityInfo     SecuritySeverity = "info"
)

// CodeReview represents a code review entity - аналог @Entity() в TypeORM
type CodeReview struct {
	// Primary key - аналог @PrimaryGeneratedColumn("uuid")
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Foreign key to User - аналог @ManyToOne(() => User, user => user.reviews)
	UserID uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Review data - аналог @Column()
	Title    string       `gorm:"size:255;not null" json:"title"`
	Code     string       `gorm:"type:text;not null" json:"code"`
	Language string       `gorm:"size:50;not null" json:"language"`
	Status   ReviewStatus `gorm:"size:20;default:'pending'" json:"status"`
	Result   *string      `gorm:"type:text" json:"result,omitempty"`

	// CustomPrompt - allows user to guide the review
	CustomPrompt *string `gorm:"type:text" json:"custom_prompt,omitempty"`

	// Timestamps - аналог @CreateDateColumn(), @UpdateDateColumn()
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations - аналог @OneToMany(() => SecurityIssue, issue => issue.review)
	SecurityIssues []SecurityIssue `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE" json:"security_issues,omitempty"`
}

// TableName returns the table name for GORM
func (CodeReview) TableName() string {
	return "code_reviews"
}

// BeforeCreate hook - generates UUID if not set
func (r *CodeReview) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// SecurityIssue represents a security vulnerability found in code - аналог @Entity()
type SecurityIssue struct {
	// Primary key - аналог @PrimaryGeneratedColumn("uuid")
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Foreign key to CodeReview - аналог @ManyToOne(() => CodeReview, review => review.securityIssues)
	ReviewID uuid.UUID   `gorm:"type:uuid;index;not null" json:"review_id"`
	Review   *CodeReview `gorm:"foreignKey:ReviewID" json:"review,omitempty"`

	// Issue details - аналог @Column()
	Severity    SecuritySeverity `gorm:"size:20;not null" json:"severity"`
	Title       string           `gorm:"size:255;not null" json:"title"`
	Description string           `gorm:"type:text;not null" json:"description"`
	LineStart   *int             `json:"line_start,omitempty"`
	LineEnd     *int             `json:"line_end,omitempty"`
	Suggestion  string           `gorm:"type:text" json:"suggestion"`
	CWE         *string          `gorm:"size:20" json:"cwe,omitempty"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for GORM
func (SecurityIssue) TableName() string {
	return "security_issues"
}

// BeforeCreate hook - generates UUID if not set
func (s *SecurityIssue) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ReviewResponse represents the API response for a code review
type ReviewResponse struct {
	ID             uuid.UUID       `json:"id"`
	UserID         uuid.UUID       `json:"user_id"`
	Title          string          `json:"title"`
	Code           string          `json:"code"`
	Language       string          `json:"language"`
	Status         ReviewStatus    `json:"status"`
	Result         *string         `json:"result,omitempty"`
	SecurityIssues []SecurityIssue `json:"security_issues,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}

// ToResponse converts CodeReview entity to ReviewResponse DTO
func (r *CodeReview) ToResponse() *ReviewResponse {
	return &ReviewResponse{
		ID:             r.ID,
		UserID:         r.UserID,
		Title:          r.Title,
		Code:           r.Code,
		Language:       r.Language,
		Status:         r.Status,
		Result:         r.Result,
		SecurityIssues: r.SecurityIssues,
		CreatedAt:      r.CreatedAt,
		CompletedAt:    r.CompletedAt,
	}
}

// CreateReviewInput represents input for creating a new code review
type CreateReviewInput struct {
	Title    string `json:"title" binding:"required,min=1,max=255"`
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required"`
}

// AnalysisRequest represents the request to OpenAI for code analysis
type AnalysisRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

// AnalysisResult represents the result from OpenAI code analysis
type AnalysisResult struct {
	Summary        string               `json:"summary"`
	SecurityIssues []SecurityIssueInput `json:"security_issues"`
	Suggestions    []string             `json:"suggestions"`
	OverallScore   int                  `json:"overall_score"`
}

// SecurityIssueInput represents input for creating a security issue
type SecurityIssueInput struct {
	Severity    SecuritySeverity `json:"severity"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	LineStart   *int             `json:"line_start,omitempty"`
	LineEnd     *int             `json:"line_end,omitempty"`
	Suggestion  string           `json:"suggestion"`
	CWE         *string          `json:"cwe,omitempty"`
}
