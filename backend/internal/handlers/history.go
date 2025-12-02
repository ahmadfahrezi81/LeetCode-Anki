package handlers

import (
	"leetcode-anki/backend/internal/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HistoryHandler struct{}

func NewHistoryHandler() *HistoryHandler {
	return &HistoryHandler{}
}

// GetHistory retrieves the user's submission history with pagination
func (h *HistoryHandler) GetHistory(c *gin.Context) {
	userID := c.GetString("user_id")

	// Parse pagination params
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	history, err := database.GetHistoryByUser(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   history,
		"limit":  limit,
		"offset": offset,
	})
}

// GetQuestionHistory retrieves all attempts for a specific question
func (h *HistoryHandler) GetQuestionHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("question_id")

	if questionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question ID is required"})
		return
	}

	history, err := database.GetHistoryByQuestion(userID, questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch question history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": history,
	})
}
