package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/secure-review/internal/domain"
)

// PDFService handles PDF generation
type PDFService struct{}

// NewPDFService creates a new PDFService
func NewPDFService() *PDFService {
	return &PDFService{}
}

// GenerateReviewPDF generates a PDF document from a code review
func (s *PDFService) GenerateReviewPDF(review *domain.ReviewResponse) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(33, 37, 41)
	pdf.CellFormat(0, 12, "Security Code Review Report", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Review Info Box
	pdf.SetFillColor(248, 249, 250)
	pdf.SetDrawColor(206, 212, 218)
	pdf.Rect(15, pdf.GetY(), 180, 35, "FD")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(73, 80, 87)
	y := pdf.GetY() + 5
	pdf.SetXY(20, y)
	pdf.CellFormat(30, 6, "Title:", "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, truncateString(review.Title, 60), "", 1, "", false, 0, "")

	pdf.SetXY(20, y+8)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(30, 6, "Language:", "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(40, 6, review.Language, "", 0, "", false, 0, "")

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(30, 6, "Status:", "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, string(review.Status), "", 1, "", false, 0, "")

	pdf.SetXY(20, y+16)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(30, 6, "Created:", "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(40, 6, review.CreatedAt.Format("2006-01-02 15:04"), "", 0, "", false, 0, "")

	if review.CompletedAt != nil {
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(30, 6, "Completed:", "", 0, "", false, 0, "")
		pdf.SetFont("Arial", "", 11)
		pdf.CellFormat(0, 6, review.CompletedAt.Format("2006-01-02 15:04"), "", 1, "", false, 0, "")
	}

	pdf.SetXY(20, y+24)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(30, 6, "Review ID:", "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 6, review.ID.String(), "", 1, "", false, 0, "")

	pdf.Ln(20)

	// Security Issues Summary
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(33, 37, 41)
	pdf.CellFormat(0, 8, "Security Issues Summary", "", 1, "", false, 0, "")
	pdf.Ln(2)

	if len(review.SecurityIssues) == 0 {
		pdf.SetFont("Arial", "I", 11)
		pdf.SetTextColor(40, 167, 69)
		pdf.CellFormat(0, 8, "No security issues found. Great job!", "", 1, "", false, 0, "")
	} else {
		// Count issues by severity
		severityCounts := map[domain.SecuritySeverity]int{
			domain.SeverityCritical: 0,
			domain.SeverityHigh:     0,
			domain.SeverityMedium:   0,
			domain.SeverityLow:      0,
			domain.SeverityInfo:     0,
		}
		for _, issue := range review.SecurityIssues {
			severityCounts[issue.Severity]++
		}

		// Summary table
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(52, 58, 64)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(45, 8, "Severity", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 8, "Count", "1", 1, "C", true, 0, "")

		pdf.SetTextColor(33, 37, 41)
		pdf.SetFont("Arial", "", 10)

		severities := []domain.SecuritySeverity{
			domain.SeverityCritical,
			domain.SeverityHigh,
			domain.SeverityMedium,
			domain.SeverityLow,
			domain.SeverityInfo,
		}

		for _, sev := range severities {
			count := severityCounts[sev]
			if count > 0 {
				r, g, b := getSeverityColor(sev)
				pdf.SetFillColor(r, g, b)
				pdf.SetTextColor(255, 255, 255)
				pdf.CellFormat(45, 7, strings.ToUpper(string(sev)), "1", 0, "C", true, 0, "")
				pdf.SetFillColor(255, 255, 255)
				pdf.SetTextColor(33, 37, 41)
				pdf.CellFormat(30, 7, fmt.Sprintf("%d", count), "1", 1, "C", false, 0, "")
			}
		}
		pdf.Ln(6)

		// Detailed Issues
		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(33, 37, 41)
		pdf.CellFormat(0, 8, "Detailed Issues", "", 1, "", false, 0, "")
		pdf.Ln(2)

		for i, issue := range review.SecurityIssues {
			// Check if we need a new page
			if pdf.GetY() > 240 {
				pdf.AddPage()
			}

			// Issue header
			r, g, b := getSeverityColor(issue.Severity)
			pdf.SetFillColor(r, g, b)
			pdf.SetTextColor(255, 255, 255)
			pdf.SetFont("Arial", "B", 10)
			severityLabel := fmt.Sprintf(" %s ", strings.ToUpper(string(issue.Severity)))
			pdf.CellFormat(25, 6, severityLabel, "", 0, "C", true, 0, "")

			pdf.SetFillColor(248, 249, 250)
			pdf.SetTextColor(33, 37, 41)
			pdf.SetFont("Arial", "B", 10)
			title := fmt.Sprintf(" #%d: %s", i+1, truncateString(issue.Title, 55))
			pdf.CellFormat(155, 6, title, "", 1, "L", true, 0, "")

			// Issue body
			pdf.SetFont("Arial", "", 9)
			pdf.SetTextColor(73, 80, 87)

			// Line numbers
			if issue.LineStart != nil {
				lineInfo := fmt.Sprintf("Lines: %d", *issue.LineStart)
				if issue.LineEnd != nil && *issue.LineEnd != *issue.LineStart {
					lineInfo = fmt.Sprintf("Lines: %d-%d", *issue.LineStart, *issue.LineEnd)
				}
				pdf.CellFormat(0, 5, lineInfo, "", 1, "", false, 0, "")
			}

			// CWE
			if issue.CWE != nil && *issue.CWE != "" {
				pdf.SetFont("Arial", "I", 9)
				pdf.CellFormat(0, 5, fmt.Sprintf("CWE: %s", *issue.CWE), "", 1, "", false, 0, "")
			}

			// Description
			pdf.SetFont("Arial", "", 9)
			pdf.SetTextColor(33, 37, 41)
			pdf.Ln(2)
			pdf.SetFont("Arial", "B", 9)
			pdf.CellFormat(0, 5, "Description:", "", 1, "", false, 0, "")
			pdf.SetFont("Arial", "", 9)
			pdf.MultiCell(0, 5, sanitizeText(issue.Description), "", "L", false)

			// Suggestion
			if issue.Suggestion != "" {
				pdf.Ln(2)
				pdf.SetFont("Arial", "B", 9)
				pdf.SetTextColor(40, 167, 69)
				pdf.CellFormat(0, 5, "Recommendation:", "", 1, "", false, 0, "")
				pdf.SetFont("Arial", "", 9)
				pdf.SetTextColor(33, 37, 41)
				pdf.MultiCell(0, 5, sanitizeText(issue.Suggestion), "", "L", false)
			}

			pdf.Ln(6)
		}
	}

	// Analysis Result
	if review.Result != nil && *review.Result != "" {
		if pdf.GetY() > 200 {
			pdf.AddPage()
		}

		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(33, 37, 41)
		pdf.CellFormat(0, 8, "Analysis Summary", "", 1, "", false, 0, "")
		pdf.Ln(2)

		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(73, 80, 87)
		pdf.MultiCell(0, 5, sanitizeText(*review.Result), "", "L", false)
	}

	// Footer
	pdf.SetY(-25)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(108, 117, 125)
	pdf.CellFormat(0, 10, fmt.Sprintf("Generated by Secure Review on %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "C", false, 0, "")

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// getSeverityColor returns RGB color for severity level
func getSeverityColor(severity domain.SecuritySeverity) (int, int, int) {
	switch severity {
	case domain.SeverityCritical:
		return 136, 14, 79 // Deep purple/magenta
	case domain.SeverityHigh:
		return 220, 53, 69 // Red
	case domain.SeverityMedium:
		return 255, 152, 0 // Orange
	case domain.SeverityLow:
		return 255, 193, 7 // Yellow
	case domain.SeverityInfo:
		return 23, 162, 184 // Cyan
	default:
		return 108, 117, 125 // Gray
	}
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// sanitizeText removes or replaces characters that fpdf cannot handle
func sanitizeText(s string) string {
	// Replace common problematic characters
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\t", "    ")

	// Remove null bytes and other control characters except newline
	var result strings.Builder
	for _, r := range s {
		if r == '\n' || (r >= 32 && r < 127) || r >= 160 {
			result.WriteRune(r)
		} else if r == '\t' {
			result.WriteString("    ")
		}
	}

	return result.String()
}
