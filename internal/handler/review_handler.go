package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
	"github.com/secure-review/internal/logger"
	"github.com/secure-review/internal/middleware"
)

// ReviewHandler handles code review endpoints
type ReviewHandler struct {
	reviewService domain.ReviewService
}

// NewReviewHandler creates a new ReviewHandler
func NewReviewHandler(reviewService domain.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// CreateReview creates a new code review
// POST /api/reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req domain.CreateReviewInput
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("[CreateReview] Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	review, err := h.reviewService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create review",
		})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// GetReview returns a specific review
// GET /api/reviews/:id
func (h *ReviewHandler) GetReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid review ID",
		})
		return
	}

	review, err := h.reviewService.GetByID(c.Request.Context(), userID, reviewID)
	if err != nil {
		if err == domain.ErrReviewNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Review not found",
			})
			return
		}
		if err == domain.ErrReviewAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get review",
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

// ListReviews returns all reviews for the current user
// GET /api/reviews
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	reviews, err := h.reviewService.GetUserReviews(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list reviews",
		})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// DeleteReview deletes a review
// DELETE /api/reviews/:id
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid review ID",
		})
		return
	}

	err = h.reviewService.Delete(c.Request.Context(), userID, reviewID)
	if err != nil {
		if err == domain.ErrReviewNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Review not found",
			})
			return
		}
		if err == domain.ErrReviewAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete review",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Review deleted successfully",
	})
}

// ReanalyzeReview re-runs analysis on an existing review
// POST /api/reviews/:id/reanalyze
func (h *ReviewHandler) ReanalyzeReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid review ID",
		})
		return
	}

	review, err := h.reviewService.ReanalyzeReview(c.Request.Context(), userID, reviewID)
	if err != nil {
		if err == domain.ErrReviewNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Review not found",
			})
			return
		}
		if err == domain.ErrReviewAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reanalyze review",
		})
		return
	}

	c.JSON(http.StatusOK, review)
}
