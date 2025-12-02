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

// GetDashboard returns Anki-style dashboard with due counts and today's stats
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID := c.GetString("user_id")

	// Get user stats
	stats, err := database.GetUserStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	// Get due counts by type (learning, review, new)
	dueCounts, err := database.GetDueCountsByType(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch due counts"})
		return
	}

	// Get today's stats
	todayStats, err := database.GetTodayStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch today's stats"})
		return
	}

	// Get next card due time
	nextCardTime, err := database.GetNextDueCardTime(userID)
	if err != nil {
		// Not critical, continue
		nextCardTime = nil
	}

	// Check if all cards are studied
	allStudied := dueCounts.LearningDue == 0 &&
		dueCounts.ReviewsDue == 0 &&
		(dueCounts.NewAvailable == 0 || dueCounts.NewStudiedToday >= 20) // Assuming 20 daily limit

	dashboard := models.DashboardData{
		Stats:           *stats,
		DueCounts:       *dueCounts,
		TodayStats:      *todayStats,
		NextCardDueAt:   nextCardTime,
		AllCardsStudied: allStudied,
	}

	c.JSON(http.StatusOK, dashboard)
}
