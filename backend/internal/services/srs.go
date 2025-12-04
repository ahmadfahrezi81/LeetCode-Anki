package services

import (
	"leetcode-anki/backend/internal/models"
	"math"
	"time"
)

// SM2Algorithm implements the SuperMemo-2 spaced repetition algorithm
type SM2Algorithm struct{}

func NewSM2Algorithm() *SM2Algorithm {
	return &SM2Algorithm{}
}

// LearningSteps defines the progression for learning cards (in minutes)
// AGGRESSIVE MODE: Single 10-minute step for rapid exploration
// This allows you to see all 150 cards quickly while still getting spaced repetition
var LearningSteps = []int{10}

// CalculateNextReview updates the review card based on the score
// Implements Anki-like spaced repetition with sub-day intervals for learning cards
func (s *SM2Algorithm) CalculateNextReview(review *models.Review, score int) {
	// Clamp score to 0-5
	if score < 0 {
		score = 0
	}
	if score > 5 {
		score = 5
	}

	review.Quality = &score
	review.TotalReviews++
	now := time.Now()

	// Handle based on current card state
	switch review.CardState {
	case "new", "learning", "relearning":
		s.handleLearningCard(review, score, now)
	case "review":
		s.handleReviewCard(review, score, now)
	default:
		// Unknown state, treat as new
		review.CardState = "learning"
		s.handleLearningCard(review, score, now)
	}

	// Update last reviewed time
	review.LastReviewedAt = &now
}

// handleLearningCard processes cards in learning/relearning state
func (s *SM2Algorithm) handleLearningCard(review *models.Review, score int, now time.Time) {
	if score < 3 {
		// Failed: Reset to first learning step (10 minutes)
		review.CurrentStep = 0
		review.IntervalMinutes = LearningSteps[0]
		review.Repetitions = 0
		review.TotalLapses++

		// Update state
		if review.CardState == "new" {
			review.CardState = "learning"
		} else if review.CardState == "review" {
			review.CardState = "relearning"
		}
	} else if score == 3 {
		// Hard: Stay at current step or move back one
		if review.CurrentStep > 0 {
			review.CurrentStep--
		}
		review.IntervalMinutes = LearningSteps[review.CurrentStep]
		review.Repetitions++

		if review.CardState == "new" {
			review.CardState = "learning"
		}
	} else if score == 4 {
		// Good: Advance to next learning step
		review.CurrentStep++
		review.Repetitions++

		if review.CurrentStep >= len(LearningSteps) {
			// Graduated! Move to review state
			review.CardState = "review"
			review.IntervalMinutes = 1440 // 1 day (changed from 6 days for rapid exploration)
			review.CurrentStep = 0
		} else {
			// Still in learning
			review.IntervalMinutes = LearningSteps[review.CurrentStep]
			if review.CardState == "new" {
				review.CardState = "learning"
			}
		}
	} else { // score == 5
		// Easy: Skip ahead or graduate immediately
		review.Repetitions++

		if review.CurrentStep >= len(LearningSteps)-2 {
			// Graduate to review with longer interval
			review.CardState = "review"
			review.IntervalMinutes = 2 * 1440 // 2 days (changed from 10 days for rapid exploration)
			review.CurrentStep = 0
		} else {
			// Skip to second-to-last learning step
			review.CurrentStep = len(LearningSteps) - 2
			review.IntervalMinutes = LearningSteps[review.CurrentStep]
			if review.CardState == "new" {
				review.CardState = "learning"
			}
		}
	}

	// Update easiness factor for learning cards too
	s.updateEasinessFactor(review, score)

	// Set next review time and interval days
	review.NextReviewAt = now.Add(time.Duration(review.IntervalMinutes) * time.Minute)
	review.IntervalDays = review.IntervalMinutes / 1440
	if review.IntervalDays == 0 && review.IntervalMinutes > 0 {
		review.IntervalDays = 1 // Display as "< 1 day" in UI
	}
}

// handleReviewCard processes cards in review state (graduated cards)
func (s *SM2Algorithm) handleReviewCard(review *models.Review, score int, now time.Time) {
	if score < 3 {
		// Failed: Move to relearning, reset to first learning step
		review.CardState = "relearning"
		review.CurrentStep = 0
		review.IntervalMinutes = LearningSteps[0]
		review.Repetitions = 0
		review.TotalLapses++
	} else {
		// Passed: Continue with SM-2 algorithm
		review.Repetitions++

		// Calculate new interval based on score
		var multiplier float64
		if score == 3 {
			// Hard: shorter interval
			multiplier = 1.2
		} else if score == 4 {
			// Good: normal interval
			multiplier = review.EasinessFactor
		} else { // score == 5
			// Easy: longer interval
			multiplier = review.EasinessFactor * 1.3
		}

		// Calculate new interval in days
		newIntervalDays := int(math.Round(float64(review.IntervalDays) * multiplier))
		if newIntervalDays < 1 {
			newIntervalDays = 1
		}

		review.IntervalDays = newIntervalDays
		review.IntervalMinutes = newIntervalDays * 1440
	}

	// Update easiness factor
	s.updateEasinessFactor(review, score)

	// Set next review time
	review.NextReviewAt = now.Add(time.Duration(review.IntervalMinutes) * time.Minute)
}

// updateEasinessFactor updates the easiness factor based on score
func (s *SM2Algorithm) updateEasinessFactor(review *models.Review, score int) {
	// SM-2 formula: EF' = EF + (0.1 - (5-q) * (0.08 + (5-q) * 0.02))
	newEF := review.EasinessFactor + (0.1 - float64(5-score)*(0.08+float64(5-score)*0.02))

	// Ensure EF doesn't drop below 1.3
	if newEF < 1.3 {
		newEF = 1.3
	}

	review.EasinessFactor = newEF
}

// GetCardStateDescription returns human-readable card state
func (s *SM2Algorithm) GetCardStateDescription(state string, intervalDays int) string {
	switch state {
	case "new":
		return "New card"
	case "learning":
		return "Learning"
	case "relearning":
		return "Relearning"
	case "review":
		if intervalDays > 21 {
			return "Mature (Review)"
		}
		return "Young (Review)"
	default:
		return "Unknown"
	}
}

// InitializeNewCard creates initial review record for a new question
func (s *SM2Algorithm) InitializeNewCard(userID, questionID string) *models.Review {
	now := time.Now()
	return &models.Review{
		UserID:          userID,
		QuestionID:      questionID,
		CardState:       "new",
		Quality:         nil,
		EasinessFactor:  2.5,
		IntervalDays:    0,
		IntervalMinutes: 0,
		CurrentStep:     0,
		Repetitions:     0,
		NextReviewAt:    now, // New cards are immediately available
		LastReviewedAt:  nil,
		TotalReviews:    0,
		TotalLapses:     0,
		CreatedAt:       now,
	}
}
