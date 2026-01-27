package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/secure-review/internal/domain"
)

var _ domain.CodeAnalyzer = (*OpenAICodeAnalyzer)(nil)

// OpenAICodeAnalyzer implements CodeAnalyzer using OpenAI API
type OpenAICodeAnalyzer struct {
	client *openai.Client
}

// NewOpenAICodeAnalyzer creates a new OpenAICodeAnalyzer
func NewOpenAICodeAnalyzer(apiKey string) *OpenAICodeAnalyzer {
	return &OpenAICodeAnalyzer{
		client: openai.NewClient(apiKey),
	}
}

// AnalyzeCode performs code review using OpenAI
func (a *OpenAICodeAnalyzer) AnalyzeCode(ctx context.Context, request *domain.AnalysisRequest) (*domain.AnalysisResult, error) {
	basePrompt := fmt.Sprintf(`You are an expert code reviewer. Analyze the following %s code and provide:
1. A brief summary of what the code does
2. Any security vulnerabilities found (with severity: critical, high, medium, low, info)
3. Code quality suggestions for improvement
4. An overall quality score from 0-100`, request.Language)

	if request.CustomPrompt != nil && *request.CustomPrompt != "" {
		basePrompt += fmt.Sprintf("\n\nUser specific instructions: %s", *request.CustomPrompt)
	}

	prompt := fmt.Sprintf(`%s

Code to review:
%s

Respond in JSON format with this structure:
{
  "summary": "string",
  "security_issues": [
    {
      "severity": "critical|high|medium|low|info",
      "title": "string",
      "description": "string",
      "line_start": number or null,
      "line_end": number or null,
      "suggestion": "string",
      "cwe": "string or null"
    }
  ],
  "suggestions": ["string"],
  "overall_score": number
}`, basePrompt, request.Code)

	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4TurboPreview,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert code reviewer specializing in security analysis and code quality. Always respond with valid JSON.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
	})

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, domain.ErrAnalysisFailed
	}

	var result domain.AnalysisResult
	content := resp.Choices[0].Message.Content

	// Clean up content if it contains markdown code blocks
	if len(content) > 7 && content[:8] == "```json\n" {
		content = content[8:]
		if len(content) > 3 && content[len(content)-3:] == "```" {
			content = content[:len(content)-3]
		}
	} else if len(content) > 3 && content[:3] == "```" {
		// Handle case where language might not be specified or just ```
		start := 3
		// finding newline
		for i := 3; i < len(content); i++ {
			if content[i] == '\n' {
				start = i + 1
				break
			}
		}
		content = content[start:]
		if len(content) > 3 && content[len(content)-3:] == "```" {
			content = content[:len(content)-3]
		}
	}

	// Try to extract JSON from the response
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// If direct parse fails, try to find JSON in the content
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return &result, nil
}

// AnalyzeSecurity performs security-focused analysis
func (a *OpenAICodeAnalyzer) AnalyzeSecurity(ctx context.Context, request *domain.AnalysisRequest) ([]domain.SecurityIssueInput, error) {
	prompt := fmt.Sprintf(`You are a security expert. Analyze the following %s code for security vulnerabilities.

Focus on:
- SQL injection
- XSS vulnerabilities
- Authentication/authorization issues
- Data exposure
- Input validation problems
- Cryptographic weaknesses
- Injection attacks
- Buffer overflows
- Path traversal
- Insecure configurations

Code to analyze:
%s

Respond in JSON format with an array of security issues:
[
  {
    "severity": "critical|high|medium|low|info",
    "title": "string",
    "description": "string",
    "line_start": number or null,
    "line_end": number or null,
    "suggestion": "string",
    "cwe": "CWE-XXX or null"
  }
]

If no security issues are found, return an empty array: []`, request.Language, request.Code)

	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4TurboPreview,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a security expert specializing in code vulnerability analysis. Always respond with valid JSON.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.2,
	})

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, domain.ErrAnalysisFailed
	}

	var issues []domain.SecurityIssueInput
	content := resp.Choices[0].Message.Content

	// Clean up content if it contains markdown code blocks
	if len(content) > 7 && content[:8] == "```json\n" {
		content = content[8:]
		if len(content) > 3 && content[len(content)-3:] == "```" {
			content = content[:len(content)-3]
		}
	} else if len(content) > 3 && content[:3] == "```" {
		start := 3
		for i := 3; i < len(content); i++ {
			if content[i] == '\n' {
				start = i + 1
				break
			}
		}
		content = content[start:]
		if len(content) > 3 && content[len(content)-3:] == "```" {
			content = content[:len(content)-3]
		}
	}

	if err := json.Unmarshal([]byte(content), &issues); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return issues, nil
}
