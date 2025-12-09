package handlers

import (
	"leetcode-anki/backend/internal/database"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type HistoryHandler struct{}

func NewHistoryHandler() *HistoryHandler {
	return &HistoryHandler{}
}

// GetHistory retrieves the user's submission history with pagination and filters
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

	// Parse filter params
	var difficulties []string
	if diffStr := c.Query("difficulty"); diffStr != "" {
		difficulties = strings.Split(diffStr, ",")
	}

	var minScore, maxScore *int
	if minScoreStr := c.Query("minScore"); minScoreStr != "" {
		if val, err := strconv.Atoi(minScoreStr); err == nil && val >= 1 && val <= 5 {
			minScore = &val
		}
	}
	if maxScoreStr := c.Query("maxScore"); maxScoreStr != "" {
		if val, err := strconv.Atoi(maxScoreStr); err == nil && val >= 1 && val <= 5 {
			maxScore = &val
		}
	}

	var states []string
	if stateStr := c.Query("state"); stateStr != "" {
		states = strings.Split(stateStr, ",")
	}

	history, err := database.GetHistoryByUser(userID, limit, offset, difficulties, minScore, maxScore, states)
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
