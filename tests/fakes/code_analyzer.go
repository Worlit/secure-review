package fakes

import (
	"context"

	"github.com/secure-review/internal/domain"
)

type FakeCodeAnalyzer struct{}

func NewFakeCodeAnalyzer() *FakeCodeAnalyzer {
	return &FakeCodeAnalyzer{}
}

func (a *FakeCodeAnalyzer) AnalyzeCode(ctx context.Context, request *domain.AnalysisRequest) (*domain.AnalysisResult, error) {
	return &domain.AnalysisResult{
		OverallScore: 85,
		Summary:      "Good code, minor issues.",
		Suggestions: []string{
			"Use meaningful variable names",
			"Add comments",
		},
		SecurityIssues: []domain.SecurityIssueInput{
			{
				Title:       "Potential SQL Injection",
				Description: "Use parameterized queries.",
				Severity:    domain.SeverityHigh,
				Suggestion:  "Use sql arguments",
			},
		},
	}, nil
}

func (a *FakeCodeAnalyzer) AnalyzeSecurity(ctx context.Context, request *domain.AnalysisRequest) ([]domain.SecurityIssueInput, error) {
	line := 10
	return []domain.SecurityIssueInput{
		{
			Title:       "Potential SQL Injection",
			Description: "Use parameterized queries.",
			Severity:    domain.SeverityHigh,
			LineStart:   &line,
			Suggestion:  "Use sql arguments",
		},
	}, nil
}
