package handlers

import (
	"leetcode-anki/backend/internal/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

type UpdateLimitRequest struct {
	NewCardsLimit int `json:"new_cards_limit" binding:"required,min=0,max=100"`
}

// UpdateDailyLimit updates the user's daily new card limit
func (h *SettingsHandler) UpdateDailyLimit(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit. Must be between 0 and 100."})
		return
	}

	err := database.UpdateUserLimit(userID, req.NewCardsLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update limit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Daily limit updated successfully",
		"new_cards_limit": req.NewCardsLimit,
	})
}
