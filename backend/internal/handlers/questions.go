package handlers

import (
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QuestionsHandler struct{}

func NewQuestionsHandler() *QuestionsHandler {
	return &QuestionsHandler{}
}

// GetAllQuestions returns all questions with user's progress
// Supports filtering by difficulty, state, topic, and sorting
func (h *QuestionsHandler) GetAllQuestions(c *gin.Context) {
	userID := c.GetString("user_id")

	// Parse filters from query params
	var filters models.QuestionFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	// Set default pagination if not provided
	if filters.Limit == 0 {
		filters.Limit = 100 // Default: return 100 questions
	}

	// Get questions with progress
	questions, err := database.GetAllQuestionsWithProgress(userID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
		return
	}

	// Get total count (for pagination)
	totalCount, err := database.GetTotalQuestionCount(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count questions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"questions":   questions,
		"total_count": totalCount,
		"filters":     filters,
	})
}

// GetQuestionDetail returns detailed info about a specific question
func (h *QuestionsHandler) GetQuestionDetail(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("id")

	// Get question
	question, err := database.GetQuestionByID(questionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	// Get user's review if exists
	review, err := database.GetReview(userID, questionID)
	if err != nil {
		// No review yet, that's okay
		review = nil
	}

	response := gin.H{
		"question": question,
	}

	if review != nil {
		response["review"] = review
		response["has_started"] = true
	} else {
		response["has_started"] = false
	}

	c.JSON(http.StatusOK, response)
}

// GetTopics returns all unique topics from questions
func (h *QuestionsHandler) GetTopics(c *gin.Context) {
	topics, err := database.GetAllTopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": topics,
	})
}

// SuspendCard allows user to suspend a card (hide it from reviews)
func (h *QuestionsHandler) SuspendCard(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("id")

	// Get review
	review, err := database.GetReview(userID, questionID)
	if err != nil || review == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	// Update to suspended state
	review.CardState = "suspended"
	err = database.UpdateReview(review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to suspend card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Card suspended successfully",
		"review":  review,
	})
}

// UnsuspendCard allows user to unsuspend a card
func (h *QuestionsHandler) UnsuspendCard(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("id")

	// Get review
	review, err := database.GetReview(userID, questionID)
	if err != nil || review == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	// Restore previous state (default to learning)
	if review.CardState == "suspended" {
		if review.IntervalDays >= 1 {
			review.CardState = "review"
		} else {
			review.CardState = "learning"
		}
	}

	err = database.UpdateReview(review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsuspend card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Card unsuspended successfully",
		"review":  review,
	})
}
