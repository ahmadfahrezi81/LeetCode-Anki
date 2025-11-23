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

// CalculateNextReview updates the review card based on the score
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

	// If score < 3, it's a lapse (failed card)
	if score < 3 {
		review.Repetitions = 0
		review.IntervalDays = 1
		review.TotalLapses++

		// Change state based on previous state
		if review.CardState == "new" {
			review.CardState = "learning"
		} else if review.CardState == "review" {
			review.CardState = "relearning"
		}
	} else {
		// Successful review (score >= 3)
		review.Repetitions++

		// Calculate new interval based on repetitions
		if review.Repetitions == 1 {
			review.IntervalDays = 1
			review.CardState = "learning"
		} else if review.Repetitions == 2 {
			review.IntervalDays = 6
			review.CardState = "review"
		} else {
			// For repetitions > 2, multiply by easiness factor
			review.IntervalDays = int(math.Round(float64(review.IntervalDays) * review.EasinessFactor))
			review.CardState = "review"

			// Mark as mature if interval > 21 days
			if review.IntervalDays > 21 {
				// Mature cards are still in "review" state, but with long intervals
				// This can be tracked in user_stats separately
			}
		}

		// Update easiness factor
		newEF := review.EasinessFactor + (0.1 - float64(5-score)*(0.08+float64(5-score)*0.02))

		// Ensure EF doesn't drop below 1.3
		if newEF < 1.3 {
			newEF = 1.3
		}

		review.EasinessFactor = newEF
	}

	// Set next review date
	now := time.Now()
	review.NextReviewAt = now.AddDate(0, 0, review.IntervalDays)
	review.LastReviewedAt = &now
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
		UserID:         userID,
		QuestionID:     questionID,
		CardState:      "new",
		Quality:        nil,
		EasinessFactor: 2.5,
		IntervalDays:   0,
		Repetitions:    0,
		NextReviewAt:   now, // New cards are immediately available
		LastReviewedAt: nil,
		TotalReviews:   0,
		TotalLapses:    0,
		CreatedAt:      now,
	}
}
