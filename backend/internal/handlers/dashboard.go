package handlers

import (
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// GetDashboard returns user stats and due review count
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID := c.GetString("user_id")

	// Get user stats
	stats, err := database.GetUserStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	// Get due review count
	dueCount, err := database.GetDueReviewCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count due reviews"})
		return
	}

	// Check if new cards are available
	newCard, _ := database.GetNewCard(userID)
	hasNewCard := newCard != nil

	dashboard := models.DashboardData{
		Stats:            *stats,
		DueReviews:       dueCount,
		AvailableNewCard: hasNewCard,
	}

	c.JSON(http.StatusOK, dashboard)
}
