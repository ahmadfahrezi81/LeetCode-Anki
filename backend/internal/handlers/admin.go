package handlers

import (
	"fmt"
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type AdminHandler struct {
	leetcodeService *services.LeetCodeService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		leetcodeService: services.NewLeetCodeService(),
	}
}

// RefreshProblems manually fetches new problems from LeetCode
func (h *AdminHandler) RefreshProblems(c *gin.Context) {
	// Get counts from query params (default: 5 easy, 10 medium, 5 hard)
	easyCount := getIntParam(c, "easy", 5)
	mediumCount := getIntParam(c, "medium", 10)
	hardCount := getIntParam(c, "hard", 5)

	// Fetch problems
	problems, err := h.leetcodeService.FetchRandomProblems(easyCount, mediumCount, hardCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch problems from LeetCode",
		})
		return
	}

	// Insert problems
	inserted := 0
	skipped := 0

	for _, problem := range problems {
		err := h.insertProblem(problem)
		if err != nil {
			if isDuplicate(err) {
				skipped++
			}
		} else {
			inserted++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Problems refreshed successfully",
		"inserted": inserted,
		"skipped":  skipped,
		"total":    len(problems),
	})
}

// GetProblemStats returns statistics about the problem pool
func (h *AdminHandler) GetProblemStats(c *gin.Context) {
	var stats struct {
		Total  int `json:"total"`
		Easy   int `json:"easy"`
		Medium int `json:"medium"`
		Hard   int `json:"hard"`
	}

	// Count total
	database.DB.QueryRow("SELECT COUNT(*) FROM questions").Scan(&stats.Total)

	// Count by difficulty
	database.DB.QueryRow("SELECT COUNT(*) FROM questions WHERE difficulty = 'Easy'").Scan(&stats.Easy)
	database.DB.QueryRow("SELECT COUNT(*) FROM questions WHERE difficulty = 'Medium'").Scan(&stats.Medium)
	database.DB.QueryRow("SELECT COUNT(*) FROM questions WHERE difficulty = 'Hard'").Scan(&stats.Hard)

	c.JSON(http.StatusOK, stats)
}

// Helper functions

func getIntParam(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func (h *AdminHandler) insertProblem(problem *services.LeetCodeProblem) error {
	topics := make([]string, len(problem.TopicTags))
	for i, tag := range problem.TopicTags {
		topics[i] = tag.Name
	}

	leetcodeID := 0
	fmt.Sscanf(problem.QuestionID, "%d", &leetcodeID)

	descriptionMarkdown := services.StripHTMLTags(problem.Content)
	correctApproach := services.GenerateApproachHint(problem)

	query := `
		INSERT INTO questions 
		(leetcode_id, title, slug, difficulty, description_markdown, topics, correct_approach)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (leetcode_id) DO NOTHING
		RETURNING id
	`

	var id string
	err := database.DB.QueryRow(
		query,
		leetcodeID,
		problem.Title,
		problem.TitleSlug,
		problem.Difficulty,
		descriptionMarkdown,
		pq.Array(topics),
		correctApproach,
	).Scan(&id)

	return err
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
