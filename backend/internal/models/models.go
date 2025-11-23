package models

import (
	"time"
)

// Question represents a LeetCode problem
type Question struct {
	ID                  string    `json:"id"`
	LeetcodeID          int       `json:"leetcode_id"`
	Title               string    `json:"title"`
	Slug                string    `json:"slug"`
	Difficulty          string    `json:"difficulty"`
	DescriptionMarkdown string    `json:"description_markdown"`
	Topics              []string  `json:"topics"`
	CorrectApproach     string    `json:"correct_approach"`
	CreatedAt           time.Time `json:"created_at"`
}

// Review represents a user's review card
type Review struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	QuestionID     string     `json:"question_id"`
	CardState      string     `json:"card_state"` // 'new', 'learning', 'review', 'relearning'
	Quality        *int       `json:"quality"`
	EasinessFactor float64    `json:"easiness_factor"`
	IntervalDays   int        `json:"interval_days"`
	Repetitions    int        `json:"repetitions"`
	NextReviewAt   time.Time  `json:"next_review_at"`
	LastReviewedAt *time.Time `json:"last_reviewed_at"`
	TotalReviews   int        `json:"total_reviews"`
	TotalLapses    int        `json:"total_lapses"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Card combines Question and Review for study sessions
type Card struct {
	Question Question `json:"question"`
	Review   Review   `json:"review"`
}

// SubmitAnswerRequest is the payload for answer submission
type SubmitAnswerRequest struct {
	QuestionID string `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

// SubmitAnswerResponse is the response after scoring
type SubmitAnswerResponse struct {
	Score           int       `json:"score"`
	Feedback        string    `json:"feedback"`
	CorrectApproach string    `json:"correct_approach"`
	NextReviewAt    time.Time `json:"next_review_at"`
	CardState       string    `json:"card_state"`
}

// UserStats represents user's overall statistics
type UserStats struct {
	UserID        string    `json:"user_id"`
	TotalCards    int       `json:"total_cards"`
	NewCards      int       `json:"new_cards"`
	LearningCards int       `json:"learning_cards"`
	ReviewCards   int       `json:"review_cards"`
	MatureCards   int       `json:"mature_cards"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DashboardData combines stats and due counts
type DashboardData struct {
	Stats            UserStats `json:"stats"`
	DueReviews       int       `json:"due_reviews"`
	AvailableNewCard bool      `json:"available_new_card"`
}
