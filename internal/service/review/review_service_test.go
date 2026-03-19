package review

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/secure-review/internal/domain"
)

// Mock implementations for testing

type mockReviewRepository struct {
	mu      sync.Mutex
	reviews map[uuid.UUID]*domain.CodeReview
	issues  map[uuid.UUID][]domain.SecurityIssue
}

func newMockReviewRepository() *mockReviewRepository {
	return &mockReviewRepository{
		reviews: make(map[uuid.UUID]*domain.CodeReview),
		issues:  make(map[uuid.UUID][]domain.SecurityIssue),
	}
}

func (r *mockReviewRepository) Create(ctx context.Context, review *domain.CodeReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	r.reviews[review.ID] = review
	return nil
}

func (r *mockReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CodeReview, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	review, ok := r.reviews[id]
	if !ok {
		return nil, domain.ErrReviewNotFound
	}
	return review, nil
}

func (r *mockReviewRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.CodeReview, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []domain.CodeReview
	for _, review := range r.reviews {
		if review.UserID == userID {
			result = append(result, *review)
		}
	}
	return result, len(result), nil
}

func (r *mockReviewRepository) Update(ctx context.Context, review *domain.CodeReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reviews[review.ID]; !ok {
		return domain.ErrReviewNotFound
	}
	r.reviews[review.ID] = review
	return nil
}

func (r *mockReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.reviews, id)
	return nil
}

func (r *mockReviewRepository) CreateSecurityIssue(ctx context.Context, issue *domain.SecurityIssue) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issues[issue.ReviewID] = append(r.issues[issue.ReviewID], *issue)
	return nil
}

func (r *mockReviewRepository) GetSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) ([]domain.SecurityIssue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.issues[reviewID], nil
}

func (r *mockReviewRepository) DeleteSecurityIssuesByReviewID(ctx context.Context, reviewID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.issues, reviewID)
	return nil
}

type mockCodeAnalyzer struct {
	result *domain.AnalysisResult
	err    error
}

func newMockCodeAnalyzer() *mockCodeAnalyzer {
	return &mockCodeAnalyzer{
		result: &domain.AnalysisResult{
			Summary:      "Test analysis completed",
			OverallScore: 85,
			SecurityIssues: []domain.SecurityIssueInput{
				{
					Severity:    domain.SeverityMedium,
					Title:       "Test Issue",
					Description: "Test description",
					Suggestion:  "Fix this",
				},
			},
			Suggestions: []string{"Consider using parameterized queries"},
		},
	}
}

func (a *mockCodeAnalyzer) AnalyzeCode(ctx context.Context, request *domain.AnalysisRequest) (*domain.AnalysisResult, error) {
	if a.err != nil {
		return nil, a.err
	}
	return a.result, nil
}

func (a *mockCodeAnalyzer) AnalyzeSecurity(ctx context.Context, request *domain.AnalysisRequest) ([]domain.SecurityIssueInput, error) {
	if a.err != nil {
		return nil, a.err
	}
	return a.result.SecurityIssues, nil
}

type mockGitHubAuthService struct {
	repoContent string
	err         error
}

func newMockGitHubAuthService() *mockGitHubAuthService {
	return &mockGitHubAuthService{
		repoContent: "package main\n\nfunc main() {}",
	}
}

func (s *mockGitHubAuthService) GetAuthURL(state string) string {
	return "https://github.com/login/oauth/authorize?state=" + state
}

func (s *mockGitHubAuthService) ExchangeCode(ctx context.Context, code string) (string, error) {
	return "test-access-token", nil
}

func (s *mockGitHubAuthService) GetUser(ctx context.Context, accessToken string) (*domain.GitHubUser, error) {
	return &domain.GitHubUser{
		ID:    12345,
		Login: "testuser",
		Email: "test@example.com",
	}, nil
}

func (s *mockGitHubAuthService) AuthenticateOrCreate(ctx context.Context, code string) (*domain.AuthResponse, error) {
	return nil, nil
}

func (s *mockGitHubAuthService) LinkAccount(ctx context.Context, userID uuid.UUID, code string) error {
	return nil
}

func (s *mockGitHubAuthService) UnlinkAccount(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *mockGitHubAuthService) ListRepositories(ctx context.Context, userID uuid.UUID) ([]domain.Repository, error) {
	return nil, nil
}

func (s *mockGitHubAuthService) ListBranches(ctx context.Context, userID uuid.UUID, owner, repo string) ([]string, error) {
	return []string{"main", "develop"}, nil
}

func (s *mockGitHubAuthService) GetRepositoryContent(ctx context.Context, userID uuid.UUID, owner, repo, ref string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.repoContent, nil
}

// Tests

func TestReviewService_Create_Success_WithCode(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	userID := uuid.New()
	code := "function test() { return 'hello'; }"
	input := &domain.CreateReviewInput{
		Title:    "Test Review",
		Code:     &code,
		Language: "javascript",
	}

	result, err := svc.Create(context.Background(), userID, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Review", result.Title)
	assert.Equal(t, "javascript", result.Language)
	assert.Equal(t, domain.ReviewStatusPending, result.Status)
}

func TestReviewService_Create_WithGitHubRepo(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	userID := uuid.New()
	owner := "testuser"
	repoName := "testrepo"
	branch := "main"
	input := &domain.CreateReviewInput{
		Title:      "Test GitHub Review",
		Language:   "go",
		RepoOwner:  &owner,
		RepoName:   &repoName,
		RepoBranch: &branch,
	}

	result, err := svc.Create(context.Background(), userID, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test GitHub Review", result.Title)
	assert.Equal(t, domain.ReviewStatusPending, result.Status)
}

func TestReviewService_Create_NoCodeOrRepo(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	userID := uuid.New()
	input := &domain.CreateReviewInput{
		Title:    "Test Review",
		Language: "python",
	}

	result, err := svc.Create(context.Background(), userID, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either code or repository details must be provided")
	assert.Nil(t, result)
}

func TestReviewService_Create_EmptyCode(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	userID := uuid.New()
	code := ""
	input := &domain.CreateReviewInput{
		Title:    "Test Review",
		Code:     &code,
		Language: "python",
	}

	result, err := svc.Create(context.Background(), userID, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code content is empty")
	assert.Nil(t, result)
}

func TestReviewService_GetByID_Success(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	userID := uuid.New()
	reviewID := uuid.New()
	resultJSON := `{"overall_score":85,"summary":"Good code","suggestions":["Add tests"]}`

	testReview := &domain.CodeReview{
		ID:        reviewID,
		UserID:    userID,
		Title:     "Test Review",
		Code:      "test code",
		Language:  "python",
		Status:    domain.ReviewStatusCompleted,
		Result:    &resultJSON,
		CreatedAt: time.Now(),
	}
	repo.reviews[reviewID] = testReview

	svc := NewReviewService(repo, analyzer, githubSvc)

	result, err := svc.GetByID(context.Background(), userID, reviewID)

	require.NoError(t, err)
	assert.Equal(t, reviewID, result.ID)
	assert.Equal(t, "Test Review", result.Title)
	assert.Equal(t, 85, result.OverallScore)
	assert.Equal(t, "Good code", result.Summary)
}

func TestReviewService_GetByID_NotFound(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	userID := uuid.New()

	result, err := svc.GetByID(context.Background(), userID, uuid.New())

	assert.Error(t, err)
	assert.Equal(t, domain.ErrReviewNotFound, err)
	assert.Nil(t, result)
}

func TestReviewService_GetByID_AccessDenied(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	ownerID := uuid.New()
	reviewID := uuid.New()

	testReview := &domain.CodeReview{
		ID:        reviewID,
		UserID:    ownerID,
		Title:     "Test Review",
		Code:      "test code",
		Language:  "python",
		Status:    domain.ReviewStatusCompleted,
		CreatedAt: time.Now(),
	}
	repo.reviews[reviewID] = testReview

	svc := NewReviewService(repo, analyzer, githubSvc)

	// Try to access with different user
	differentUser := uuid.New()
	result, err := svc.GetByID(context.Background(), differentUser, reviewID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrReviewAccessDenied, err)
	assert.Nil(t, result)
}

func TestReviewService_GetUserReviews_Success(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	userID := uuid.New()

	// Create multiple reviews
	for i := 0; i < 3; i++ {
		review := &domain.CodeReview{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "Test Review",
			Code:      "test code",
			Language:  "go",
			Status:    domain.ReviewStatusCompleted,
			CreatedAt: time.Now(),
		}
		repo.reviews[review.ID] = review
	}

	svc := NewReviewService(repo, analyzer, githubSvc)

	result, err := svc.GetUserReviews(context.Background(), userID, 1, 10)

	require.NoError(t, err)
	assert.Equal(t, 3, result.Total)
	assert.Len(t, result.Reviews, 3)
}

func TestReviewService_GetUserReviews_Pagination(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	userID := uuid.New()

	// Test invalid page
	result, err := svc.GetUserReviews(context.Background(), userID, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Page) // Should default to 1

	// Test invalid page size
	result, err = svc.GetUserReviews(context.Background(), userID, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, 20, result.PageSize) // Should default to 20

	// Test max page size
	result, err = svc.GetUserReviews(context.Background(), userID, 1, 150)
	require.NoError(t, err)
	assert.Equal(t, 20, result.PageSize) // Should cap at 20
}

func TestReviewService_Delete_Success(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	userID := uuid.New()
	reviewID := uuid.New()

	testReview := &domain.CodeReview{
		ID:        reviewID,
		UserID:    userID,
		Title:     "Test Review",
		Code:      "test code",
		Language:  "go",
		Status:    domain.ReviewStatusCompleted,
		CreatedAt: time.Now(),
	}
	repo.reviews[reviewID] = testReview

	svc := NewReviewService(repo, analyzer, githubSvc)

	err := svc.Delete(context.Background(), userID, reviewID)

	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(context.Background(), reviewID)
	assert.Error(t, err)
}

func TestReviewService_Delete_NotFound(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())

	assert.Error(t, err)
	assert.Equal(t, domain.ErrReviewNotFound, err)
}

func TestReviewService_Delete_AccessDenied(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	ownerID := uuid.New()
	reviewID := uuid.New()

	testReview := &domain.CodeReview{
		ID:        reviewID,
		UserID:    ownerID,
		Title:     "Test Review",
		Code:      "test code",
		Language:  "go",
		Status:    domain.ReviewStatusCompleted,
		CreatedAt: time.Now(),
	}
	repo.reviews[reviewID] = testReview

	svc := NewReviewService(repo, analyzer, githubSvc)

	// Try to delete with different user
	differentUser := uuid.New()
	err := svc.Delete(context.Background(), differentUser, reviewID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrReviewAccessDenied, err)

	// Verify review still exists
	_, err = repo.GetByID(context.Background(), reviewID)
	assert.NoError(t, err)
}

func TestReviewService_ReanalyzeReview_Success(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	userID := uuid.New()
	reviewID := uuid.New()

	testReview := &domain.CodeReview{
		ID:        reviewID,
		UserID:    userID,
		Title:     "Test Review",
		Code:      "test code",
		Language:  "go",
		Status:    domain.ReviewStatusCompleted,
		CreatedAt: time.Now(),
	}
	repo.reviews[reviewID] = testReview

	svc := NewReviewService(repo, analyzer, githubSvc)

	result, err := svc.ReanalyzeReview(context.Background(), userID, reviewID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.ReviewStatusPending, result.Status)
}

func TestReviewService_ReanalyzeReview_NotFound(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	result, err := svc.ReanalyzeReview(context.Background(), uuid.New(), uuid.New())

	assert.Error(t, err)
	assert.Equal(t, domain.ErrReviewNotFound, err)
	assert.Nil(t, result)
}

func TestReviewService_ReanalyzeReview_AccessDenied(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	ownerID := uuid.New()
	reviewID := uuid.New()

	testReview := &domain.CodeReview{
		ID:        reviewID,
		UserID:    ownerID,
		Title:     "Test Review",
		Code:      "test code",
		Language:  "go",
		Status:    domain.ReviewStatusCompleted,
		CreatedAt: time.Now(),
	}
	repo.reviews[reviewID] = testReview

	svc := NewReviewService(repo, analyzer, githubSvc)

	differentUser := uuid.New()
	result, err := svc.ReanalyzeReview(context.Background(), differentUser, reviewID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrReviewAccessDenied, err)
	assert.Nil(t, result)
}

func TestReviewService_parseResultMetadata(t *testing.T) {
	svc := &ReviewServiceImpl{}

	// Test with valid JSON
	validJSON := `{"overall_score":85,"summary":"Good code quality","suggestions":["Add tests","Improve docs"]}`
	score, summary, suggestions := svc.parseResultMetadata(&validJSON)

	assert.Equal(t, 85, score)
	assert.Equal(t, "Good code quality", summary)
	assert.Equal(t, []string{"Add tests", "Improve docs"}, suggestions)
}

func TestReviewService_parseResultMetadata_NilResult(t *testing.T) {
	svc := &ReviewServiceImpl{}

	score, summary, suggestions := svc.parseResultMetadata(nil)

	assert.Equal(t, 0, score)
	assert.Empty(t, summary)
	assert.Nil(t, suggestions)
}

func TestReviewService_parseResultMetadata_EmptyResult(t *testing.T) {
	svc := &ReviewServiceImpl{}

	emptyStr := ""
	score, summary, suggestions := svc.parseResultMetadata(&emptyStr)

	assert.Equal(t, 0, score)
	assert.Empty(t, summary)
	assert.Nil(t, suggestions)
}

func TestReviewService_parseResultMetadata_InvalidJSON(t *testing.T) {
	svc := &ReviewServiceImpl{}

	invalidJSON := "not valid json"
	score, summary, suggestions := svc.parseResultMetadata(&invalidJSON)

	assert.Equal(t, 0, score)
	assert.Empty(t, summary)
	assert.Nil(t, suggestions)
}

func TestReviewService_Create_WithCustomPrompt(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	userID := uuid.New()
	code := "def hello(): pass"
	customPrompt := "Focus on security vulnerabilities"
	input := &domain.CreateReviewInput{
		Title:        "Test Review",
		Code:         &code,
		Language:     "python",
		CustomPrompt: &customPrompt,
	}

	result, err := svc.Create(context.Background(), userID, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.CustomPrompt)
	assert.Equal(t, customPrompt, *result.CustomPrompt)
}

func TestNewReviewService(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	assert.NotNil(t, svc)
	assert.Equal(t, repo, svc.reviewRepo)
	assert.Equal(t, analyzer, svc.codeAnalyzer)
	assert.Equal(t, githubSvc, svc.githubAuthService)
}

func TestReviewService_AnalyzeCode_Failure(t *testing.T) {
	repo := newMockReviewRepository()
	analyzer := newMockCodeAnalyzer()
	analyzer.err = errors.New("analysis failed")
	githubSvc := newMockGitHubAuthService()

	svc := NewReviewService(repo, analyzer, githubSvc)

	userID := uuid.New()
	code := "test code"
	input := &domain.CreateReviewInput{
		Title:    "Test Review",
		Code:     &code,
		Language: "go",
	}

	result, err := svc.Create(context.Background(), userID, input)

	// Create should still succeed, analysis happens async
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.ReviewStatusPending, result.Status)
}
