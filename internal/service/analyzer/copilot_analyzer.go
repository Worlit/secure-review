package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/secure-review/internal/domain"
)

var _ domain.CodeAnalyzer = (*CopilotCodeAnalyzer)(nil)

const (
	// Default model for Copilot
	defaultCopilotModel = "gpt-4o"
	// Max characters per chunk (approx 10000 chars to enable system prompt and response within 8k tokens)
	maxChunkSize = 12000
)

// CopilotCodeAnalyzer implements CodeAnalyzer using GitHub Copilot API
type CopilotCodeAnalyzer struct {
	apiKey string
	model  string
	apiURL string
	client *http.Client
}

// NewCopilotCodeAnalyzer creates a new CopilotCodeAnalyzer
func NewCopilotCodeAnalyzer(apiKey, model, apiURL string) *CopilotCodeAnalyzer {
	if model == "" {
		model = defaultCopilotModel
	}
	// Use default endpoint if not provided, but we expect it to be passed from config now
	if apiURL == "" {
		apiURL = "https://models.inference.ai.azure.com/chat/completions"
	}
	return &CopilotCodeAnalyzer{
		apiKey: apiKey,
		model:  model,
		apiURL: apiURL,
		client: &http.Client{},
	}
}

// ChatMessage represents a message in the chat completion request
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request to Copilot API
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

// ChatCompletionChoice represents a choice in the response
type ChatCompletionChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
}

// ChatCompletionResponse represents the response from Copilot API
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Choices []ChatCompletionChoice `json:"choices"`
}

// AnalyzeCode performs code review using GitHub Copilot
func (a *CopilotCodeAnalyzer) AnalyzeCode(ctx context.Context, request *domain.AnalysisRequest) (*domain.AnalysisResult, error) {
	// If code is too large, split it and analyze only the first chunk for the summary/structure
	// Or simply process chunks and merge (complex for overall score/summary)
	// For "AnalyzeCode", the user expects a holistic review.
	// Simple strategy: If code > limit, truncate or analyze critical parts.
	// For now, let's just truncate for AnalyzeCode to avoid failure, and let AnalyzeSecurity (the main feature) handle full scan.

	codeToAnalyze := request.Code
	truncated := false
	if len(codeToAnalyze) > maxChunkSize {
		codeToAnalyze = codeToAnalyze[:maxChunkSize] + "\n... (code truncated for detailed analysis) ..."
		truncated = true
	}

	basePrompt := fmt.Sprintf(`You are an expert code reviewer. Analyze the following %s code and provide:
1. A brief summary of what the code does
2. Any security vulnerabilities found (with severity: critical, high, medium, low, info)
3. Code quality suggestions for improvement
4. An overall quality score from 0-100`, request.Language)

	if request.CustomPrompt != nil && *request.CustomPrompt != "" {
		basePrompt += fmt.Sprintf("\n\nUser specific instructions: %s", *request.CustomPrompt)
	}

	if truncated {
		basePrompt += "\n\nNote: The code was truncated due to size limits. Analyze the provided part."
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
}`, basePrompt, codeToAnalyze)

	content, err := a.sendRequest(ctx, prompt, "You are an expert code reviewer specializing in code quality. Always respond with valid JSON.", 0.3)
	if err != nil {
		return nil, err
	}

	var result domain.AnalysisResult
	content = cleanJSONContent(content)

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Copilot response: %w", err)
	}

	return &result, nil
}

// AnalyzeSecurity performs security-focused analysis
func (a *CopilotCodeAnalyzer) AnalyzeSecurity(ctx context.Context, request *domain.AnalysisRequest) ([]domain.SecurityIssueInput, error) {
	if len(request.Code) <= maxChunkSize {
		return a.analyzeSecurityChunk(ctx, request.Code, request.Language, 0)
	}

	// Split code into chunks
	chunks := splitCodeIntoChunks(request.Code, maxChunkSize)
	var allIssues []domain.SecurityIssueInput

	for i, chunk := range chunks {
		// Calculate rough line offset based on previous chunks (approximation)
		// To do this accurately we would need to count newlines in previous chunks.
		// For simplicity, we just pass the chunk. Improving line numbers would require better tracking.
		// Let's count lines in previous chunks.
		lineOffset := 0
		for j := 0; j < i; j++ {
			lineOffset += strings.Count(chunks[j], "\n")
		}

		issues, err := a.analyzeSecurityChunk(ctx, chunk, request.Language, lineOffset)
		if err != nil {
			// Log error but continue with valid chunks or fail partial?
			// Let's log and continue to get as much as possible
			fmt.Printf("Error analyzing chunk %d: %v\n", i, err)
			continue
		}
		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}

func (a *CopilotCodeAnalyzer) analyzeSecurityChunk(ctx context.Context, code, language string, lineOffset int) ([]domain.SecurityIssueInput, error) {
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

If no security issues are found, return an empty array: []`, language, code)

	content, err := a.sendRequest(ctx, prompt, "You are a security expert specializing in code vulnerability analysis. Always respond with valid JSON.", 0.2)
	if err != nil {
		return nil, err
	}

	var issues []domain.SecurityIssueInput
	content = cleanJSONContent(content)

	if err := json.Unmarshal([]byte(content), &issues); err != nil {
		return nil, fmt.Errorf("failed to parse Copilot response: %w", err)
	}

	// Adjust line numbers if offset is provided
	if lineOffset > 0 {
		for i := range issues {
			if issues[i].LineStart != nil {
				start := *issues[i].LineStart + lineOffset
				issues[i].LineStart = &start
			}
			if issues[i].LineEnd != nil {
				end := *issues[i].LineEnd + lineOffset
				issues[i].LineEnd = &end
			}
		}
	}

	return issues, nil
}

// splitCodeIntoChunks splits the code into chunks respecting max size and line boundaries
func splitCodeIntoChunks(code string, maxSize int) []string {
	if len(code) <= maxSize {
		return []string{code}
	}

	var chunks []string
	lines := strings.Split(code, "\n")
	currentChunk := ""

	for _, line := range lines {
		// +1 for newline character
		if len(currentChunk)+len(line)+1 > maxSize {
			chunks = append(chunks, currentChunk)
			currentChunk = line
		} else {
			if currentChunk != "" {
				currentChunk += "\n"
			}
			currentChunk += line
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// sendRequest sends a request to the Copilot API
func (a *CopilotCodeAnalyzer) sendRequest(ctx context.Context, userPrompt, systemPrompt string, temperature float64) (string, error) {
	reqBody := ChatCompletionRequest{
		Model: a.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	// GitHub Models might require explicit model name header or different handling, but usually Bearer is enough.
	// Some docs suggest `Authorization: Bearer <token>` is sufficient for models.inference.ai.azure.com

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Copilot API error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Copilot API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse Copilot response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", domain.ErrAnalysisFailed
	}

	return chatResp.Choices[0].Message.Content, nil
}

// cleanJSONContent removes markdown code blocks from the response
func cleanJSONContent(content string) string {
	content = strings.TrimSpace(content)

	// Remove ```json or ``` prefix
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "\n")
	} else if strings.HasPrefix(content, "```") {
		// Find the first newline after ```
		idx := strings.Index(content, "\n")
		if idx != -1 {
			content = content[idx+1:]
		}
	}

	// Remove trailing ```
	content = strings.TrimSuffix(content, "```")

	return strings.TrimSpace(content)
}
