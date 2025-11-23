package handlers

import (
	"context"
	"fmt"
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/models"
	"leetcode-anki/backend/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	srsService      *services.SM2Algorithm
	llmService      *services.LLMService
	leetcodeService *services.LeetCodeService
}

func NewReviewHandler() *ReviewHandler {
	return &ReviewHandler{
		srsService:      services.NewSM2Algorithm(),
		llmService:      services.NewLLMService(),
		leetcodeService: services.NewLeetCodeService(),
	}
}

// GetNextCard retrieves the next card for study (review or new)
func (h *ReviewHandler) GetNextCard(c *gin.Context) {
	userID := c.GetString("user_id")

	// First, check for due reviews
	card, err := database.GetNextCard(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review card"})
		return
	}

	if card != nil {
		c.JSON(http.StatusOK, gin.H{
			"card": card,
			"type": "review",
		})
		return
	}

	// No reviews due, get a new card
	question, err := database.GetNewCard(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch new card"})
		return
	}

	if question == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "No cards available. Check back later!",
			"card":    nil,
		})
		return
	}

	// Check if we need to refresh problem pool (background task)
	go h.checkAndRefreshProblems(userID)

	// Initialize new review record
	review := h.srsService.InitializeNewCard(userID, question.ID)
	err = database.CreateReview(review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	// Refresh user stats
	_ = database.RefreshUserStats(userID)

	c.JSON(http.StatusOK, gin.H{
		"card": models.Card{
			Question: *question,
			Review:   *review,
		},
		"type": "new",
	})
}

// SubmitAnswer handles answer submission and scoring
func (h *ReviewHandler) SubmitAnswer(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get question
	question, err := database.GetQuestionByID(req.QuestionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	// Get review record
	review, err := database.GetReview(userID, req.QuestionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review"})
		return
	}

	if review == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Review not found. Get the card first."})
		return
	}

	// Score the answer using LLM
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	score, feedback, err := h.llmService.ScoreAnswer(
		ctx,
		question.Title,
		question.DescriptionMarkdown,
		question.CorrectApproach,
		req.Answer,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to score answer"})
		return
	}

	// Update review using SM-2 algorithm
	h.srsService.CalculateNextReview(review, score)

	// Save updated review
	err = database.UpdateReview(review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	// Refresh user stats
	_ = database.RefreshUserStats(userID)

	// Return response
	c.JSON(http.StatusOK, models.SubmitAnswerResponse{
		Score:           score,
		Feedback:        feedback,
		CorrectApproach: question.CorrectApproach,
		NextReviewAt:    review.NextReviewAt,
		CardState:       review.CardState,
	})
}

// checkAndRefreshProblems triggers background refresh if problem pool is low
func (h *ReviewHandler) checkAndRefreshProblems(userID string) {
	count, err := database.GetUnusedProblemCount(userID)
	if err != nil || count > 20 {
		return // Enough problems available
	}

	// Problem pool is low, fetch more
	// This runs in background, doesn't block user
	problems, err := h.leetcodeService.FetchRandomProblems(5, 10, 5)
	if err != nil {
		return // Silently fail, not critical
	}

	// Insert new problems
	for _, problem := range problems {
		_ = h.insertProblem(problem) // Ignore errors (duplicates are fine)
	}
}

// insertProblem helper to insert a LeetCode problem
func (h *ReviewHandler) insertProblem(problem *services.LeetCodeProblem) error {
	// This is duplicated from seed/main.go - could be refactored to a shared function
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
	`

	_, err := database.DB.Exec(
		query,
		leetcodeID,
		problem.Title,
		problem.TitleSlug,
		problem.Difficulty,
		descriptionMarkdown,
		topics,
		correctApproach,
	)

	return err
}
