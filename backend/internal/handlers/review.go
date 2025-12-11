package handlers

import (
	"context"
	"fmt"
	"leetcode-anki/backend/config"
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/models"
	"leetcode-anki/backend/internal/services"
	"log"
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

// GetNextCard retrieves the next card following Anki's priority order
// Priority: 1. New cards (within daily limit, USER REQUESTED PRIORITY) -> 2. Learning cards -> 3. Review cards
func (h *ReviewHandler) GetNextCard(c *gin.Context) {
	userID := c.GetString("user_id")

	// Get due counts for response
	dueCounts, err := database.GetDueCountsByType(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch due counts"})
		return
	}

	// PRIORITY 1: New cards (check daily limit)
	// User specifically requested to prioritize 5 new cards a day over reviews

	// 1.1 Check if we have an existing 'new' card that is in progress (already created but not studied)
	// This fixes the bug where refreshing fetches a NEW card instead of the pending one
	existingNewCard, err := database.GetNextNewCardReview(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing new card"})
		return
	}

	if existingNewCard != nil {
		c.JSON(http.StatusOK, models.NextCardResponse{
			Card:      existingNewCard,
			Type:      "new",
			Message:   "New card to learn",
			DueCounts: *dueCounts,
		})
		return
	}

	// 1.2 If no existing new card, check if we can fetch a FRESH new card
	newCardsLimit := config.AppConfig.NewCardsPerDay // Default: 5
	if dueCounts.NewStudiedToday < newCardsLimit {
		// Try to get a new card
		question, err := database.GetNewCard(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch new card"})
			return
		}

		if question != nil {
			// Found a new card! Process it.

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

			// Update due counts after creating new card
			dueCounts.NewStudiedToday++
			dueCounts.NewAvailable--

			c.JSON(http.StatusOK, models.NextCardResponse{
				Card: &models.Card{
					Question: *question,
					Review:   *review,
				},
				Type:      "new",
				Message:   "New card to learn",
				DueCounts: *dueCounts,
			})
			return
		}
		// If question == nil, it means no new cards are available in the DB.
		// Fall through to check Learning/Review cards.
	}

	// PRIORITY 2: Learning/Relearning cards (due now)
	card, err := database.GetNextLearningCard(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch learning card"})
		return
	}

	if card != nil {
		c.JSON(http.StatusOK, models.NextCardResponse{
			Card:      card,
			Type:      "learning",
			Message:   "Continue learning this card",
			DueCounts: *dueCounts,
		})
		return
	}

	// PRIORITY 3: Review cards (due today)
	card, err = database.GetNextReviewCard(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review card"})
		return
	}

	if card != nil {
		c.JSON(http.StatusOK, models.NextCardResponse{
			Card:      card,
			Type:      "review",
			Message:   "Review this card",
			DueCounts: *dueCounts,
		})
		return
	}

	// If we're here, no cards are available to show immediately
	nextCardTime, _ := database.GetNextDueCardTime(userID)

	if nextCardTime == nil {
		// All cards studied!
		message := "🎉 Congratulations! You've studied all available cards."
		if dueCounts.NewStudiedToday >= newCardsLimit {
			message += fmt.Sprintf(" You also reached your daily limit of %d new cards!", newCardsLimit)
		}

		c.JSON(http.StatusOK, models.NextCardResponse{
			Card:          nil,
			Type:          "",
			Message:       message,
			NextCardDueAt: nil,
			DueCounts:     *dueCounts,
		})
	} else {
		// Cards coming up later
		timeUntil := time.Until(*nextCardTime)
		message := fmt.Sprintf("No cards due right now. Next card in %s", formatDuration(timeUntil))

		if dueCounts.NewStudiedToday >= newCardsLimit {
			message = fmt.Sprintf("Daily new card limit reached (%d/%d). %s", dueCounts.NewStudiedToday, newCardsLimit, message)
		}

		c.JSON(http.StatusOK, models.NextCardResponse{
			Card:          nil,
			Type:          "",
			Message:       message,
			NextCardDueAt: nextCardTime,
			DueCounts:     *dueCounts,
		})
	}
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var score int
	var feedback string
	var correctApproach string
	var subScores *models.SubScores
	var solutionBreakdown *models.SolutionBreakdown

	// 🚀 OPTIMIZATION: Check if we have cached solution breakdown
	if question.SolutionBreakdown != nil {
		// ⚡ FAST PATH: Use cached solution, only score and provide feedback (~3-5s)
		log.Printf("⚡ Using cached solution for question %s - FAST scoring", question.ID)

		score, feedback, subScores, err = h.llmService.ScoreAnswerOnly(
			ctx,
			question.Title,
			question.DescriptionMarkdown,
			req.Answer,
			question.SolutionBreakdown,
		)

		if err != nil {
			log.Printf("❌ Fast LLM ERROR: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to score answer: %v", err)})
			return
		}

		// Use cached solution breakdown
		solutionBreakdown = question.SolutionBreakdown
		correctApproach = solutionBreakdown.Pattern + ": " + solutionBreakdown.WhyThisPattern

		log.Printf("⚡ Fast scoring complete in ~3-5s")
	} else {
		// 🐢 SLOW PATH: First time seeing this question, generate full solution (~16s)
		log.Printf("🔄 No cached solution for question %s - FULL scoring and caching", question.ID)

		score, feedback, correctApproach, subScores, solutionBreakdown, err = h.llmService.ScoreAnswer(
			ctx,
			question.Title,
			question.DescriptionMarkdown,
			req.Answer,
		)

		if err != nil {
			log.Printf("❌ LLM ERROR: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to score answer: %v", err)})
			return
		}

		// 💾 Cache the solution breakdown for future use
		if solutionBreakdown != nil {
			if err := database.UpdateQuestionSolution(question.ID, solutionBreakdown); err != nil {
				log.Printf("⚠️ Failed to cache solution breakdown: %v", err)
				// Don't fail the request, just log the error
			} else {
				log.Printf("✅ Solution breakdown cached for question %s", question.ID)
			}
		}

		log.Printf("🐢 Full scoring complete in ~16s")
	}

	log.Printf("📊 Score: %d", score)
	log.Printf("📝 Feedback: %s", feedback)
	log.Printf("📈 SubScores: %+v", subScores)

	// Update review using SM-2 algorithm
	h.srsService.CalculateNextReview(review, score)

	// Save updated review
	err = database.UpdateReview(review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	// Save to history
	history := &models.History{
		UserID:            userID,
		QuestionID:        req.QuestionID,
		UserAnswer:        req.Answer,
		SubmittedAt:       time.Now(),
		Score:             score,
		Feedback:          feedback,
		CorrectApproach:   correctApproach,
		SubScores:         subScores,
		SolutionBreakdown: solutionBreakdown,
		NextReviewAt:      review.NextReviewAt,
		CardState:         review.CardState,
		IntervalMinutes:   review.IntervalMinutes,
		IntervalDays:      review.IntervalDays,
		TimeSpentSeconds:  req.TimeSpentSeconds,
	}

	err = database.CreateHistory(history)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("⚠️ Failed to save history: %v", err)
	} else {
		log.Printf("✅ History saved successfully: ID=%s", history.ID)
	}

	// Refresh user stats
	_ = database.RefreshUserStats(userID)

	// Return response with enhanced info
	c.JSON(http.StatusOK, models.SubmitAnswerResponse{
		Score:             score,
		Feedback:          feedback,
		CorrectApproach:   correctApproach,
		SubScores:         subScores,
		SolutionBreakdown: solutionBreakdown,
		NextReviewAt:      review.NextReviewAt,
		CardState:         review.CardState,
		IntervalMinutes:   review.IntervalMinutes,
		IntervalDays:      review.IntervalDays,
	})
}

// SkipCard handles skipping a card (treats as "Again" - failed)
func (h *ReviewHandler) SkipCard(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.SkipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
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

	// Treat skip as "Again" (score 0 = failed)
	h.srsService.CalculateNextReview(review, 0)

	// Save updated review
	err = database.UpdateReview(review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	// Refresh user stats
	_ = database.RefreshUserStats(userID)

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"message":          "Card skipped and marked as 'Again'",
		"next_review_at":   review.NextReviewAt,
		"card_state":       review.CardState,
		"interval_minutes": review.IntervalMinutes,
		"interval_days":    review.IntervalDays,
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
	topics := make([]string, len(problem.TopicTags))
	for i, tag := range problem.TopicTags {
		topics[i] = tag.Name
	}

	leetcodeID := 0
	fmt.Sscanf(problem.QuestionID, "%d", &leetcodeID)

	descriptionMarkdown := services.StripHTMLTags(problem.Content)

	query := `
		INSERT INTO questions 
		(leetcode_id, title, slug, difficulty, description_markdown, topics)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (leetcode_id) 
		DO UPDATE SET 
			description_markdown = EXCLUDED.description_markdown,
			topics = EXCLUDED.topics
	`

	_, err := database.DB.Exec(
		query,
		leetcodeID,
		problem.Title,
		problem.TitleSlug,
		problem.Difficulty,
		descriptionMarkdown,
		topics,
	)

	return err
}

// Helper function to format duration nicely
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "less than a minute"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

// GetSolutionBreakdown returns the solution breakdown for a question
// If not cached, it generates it in the background
func (h *ReviewHandler) GetSolutionBreakdown(c *gin.Context) {
	questionID := c.Param("questionId")

	// Get question
	question, err := database.GetQuestionByID(questionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	// If we have cached solution, return it immediately
	if question.SolutionBreakdown != nil {
		c.JSON(http.StatusOK, gin.H{
			"solution_breakdown": question.SolutionBreakdown,
			"cached":             true,
		})
		return
	}

	// No cached solution - generate it now
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	log.Printf("🔄 Generating solution breakdown for question %s", question.ID)

	_, _, _, _, solutionBreakdown, err := h.llmService.ScoreAnswer(
		ctx,
		question.Title,
		question.DescriptionMarkdown,
		"", // Empty answer - we just want the solution
	)

	if err != nil {
		log.Printf("❌ Failed to generate solution breakdown: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate solution breakdown"})
		return
	}

	// Cache the solution for future use
	if solutionBreakdown != nil {
		if err := database.UpdateQuestionSolution(question.ID, solutionBreakdown); err != nil {
			log.Printf("⚠️ Failed to cache solution breakdown: %v", err)
		} else {
			log.Printf("✅ Solution breakdown generated and cached for question %s", question.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"solution_breakdown": solutionBreakdown,
		"cached":             false,
	})
}
