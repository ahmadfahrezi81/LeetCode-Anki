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
	CreatedAt           time.Time `json:"created_at"`
}

// Review represents a user's review card
type Review struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	QuestionID      string     `json:"question_id"`
	CardState       string     `json:"card_state"` // 'new', 'learning', 'review', 'relearning'
	Quality         *int       `json:"quality"`
	EasinessFactor  float64    `json:"easiness_factor"`
	IntervalDays    int        `json:"interval_days"`    // For display purposes (backward compatibility)
	IntervalMinutes int        `json:"interval_minutes"` // Precise interval in minutes
	CurrentStep     int        `json:"current_step"`     // Learning step position (0-3)
	Repetitions     int        `json:"repetitions"`
	NextReviewAt    time.Time  `json:"next_review_at"`
	LastReviewedAt  *time.Time `json:"last_reviewed_at"`
	TotalReviews    int        `json:"total_reviews"`
	TotalLapses     int        `json:"total_lapses"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Card combines Question and Review for study sessions
type Card struct {
	Question Question `json:"question"`
	Review   Review   `json:"review"`
}

// QuestionWithProgress combines Question with user's progress info
type QuestionWithProgress struct {
	Question       Question   `json:"question"`
	CardState      string     `json:"card_state"` // "unseen", "learning", "review", etc.
	IntervalDays   int        `json:"interval_days"`
	NextReviewAt   *time.Time `json:"next_review_at"`
	TotalReviews   int        `json:"total_reviews"`
	TotalLapses    int        `json:"total_lapses"`
	EasinessFactor float64    `json:"easiness_factor"`
}

// QuestionFilters for filtering question list
type QuestionFilters struct {
	Difficulty string `form:"difficulty"` // "Easy", "Medium", "Hard"
	State      string `form:"state"`      // "unseen", "new", "learning", "review", "relearning"
	Topic      string `form:"topic"`      // Any topic tag
	SortBy     string `form:"sort_by"`    // "difficulty", "title", "progress", "leetcode_id"
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
}

// SubmitAnswerRequest is the payload for answer submission
type SubmitAnswerRequest struct {
	QuestionID       string `json:"question_id" binding:"required"`
	Answer           string `json:"answer" binding:"required"`
	TimeSpentSeconds int    `json:"time_spent_seconds"` // Time spent on this card in seconds
}

// SkipRequest is the payload for skipping a card
type SkipRequest struct {
	QuestionID string `json:"question_id" binding:"required"`
}

// SubmitAnswerResponse is the response after scoring
type SubmitAnswerResponse struct {
	Score             int                `json:"score"`
	Feedback          string             `json:"feedback"`
	CorrectApproach   string             `json:"correct_approach"`
	SubScores         *SubScores         `json:"sub_scores"`         // Remove ,omitempty
	SolutionBreakdown *SolutionBreakdown `json:"solution_breakdown"` // Remove ,omitempty
	NextReviewAt      time.Time          `json:"next_review_at"`
	CardState         string             `json:"card_state"`
	IntervalMinutes   int                `json:"interval_minutes"`
	IntervalDays      int                `json:"interval_days"`
}

// SubScores provides granular feedback on different aspects
type SubScores struct {
	PatternRecognition      int `json:"pattern_recognition"`      // 0-5: Did they identify the right pattern?
	AlgorithmicCorrectness  int `json:"algorithmic_correctness"`  // 0-5: Is their approach correct?
	ComplexityUnderstanding int `json:"complexity_understanding"` // 0-5: Do they understand time/space complexity?
	EdgeCaseAwareness       int `json:"edge_case_awareness"`      // 0-5: Did they consider edge cases?
}

// SolutionBreakdown provides comprehensive solution explanation
type SolutionBreakdown struct {
	Pattern               string   `json:"pattern"`                // e.g., "Two Pointers", "Dynamic Programming"
	WhyThisPattern        string   `json:"why_this_pattern"`       // Explanation of why this pattern fits
	ApproachSteps         []string `json:"approach_steps"`         // Step-by-step algorithm explanation
	Pseudocode            string   `json:"pseudocode"`             // Reference pseudocode
	TimeComplexity        string   `json:"time_complexity"`        // e.g., "O(n)"
	SpaceComplexity       string   `json:"space_complexity"`       // e.g., "O(1)"
	ComplexityExplanation string   `json:"complexity_explanation"` // Why this complexity
	KeyInsights           []string `json:"key_insights"`           // What makes this solution optimal
	CommonPitfalls        []string `json:"common_pitfalls"`        // What to watch out for
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

// DueCounts represents cards due by type (Anki-style)
type DueCounts struct {
	LearningDue     int `json:"learning_due"`      // Learning/relearning cards due now
	ReviewsDue      int `json:"reviews_due"`       // Review cards due today
	NewAvailable    int `json:"new_available"`     // Total new cards available
	NewStudiedToday int `json:"new_studied_today"` // New cards studied today
}

// TodayStats represents today's study session
type TodayStats struct {
	ReviewsDone      int `json:"reviews_done"`
	NewCardsDone     int `json:"new_cards_done"`
	TimeSpentMinutes int `json:"time_spent_minutes"`
}

// DashboardData combines stats and due counts (Anki-style)
type DashboardData struct {
	Stats           UserStats  `json:"stats"`
	DueCounts       DueCounts  `json:"due_counts"`
	TodayStats      TodayStats `json:"today_stats"`
	NextCardDueAt   *time.Time `json:"next_card_due_at"`
	AllCardsStudied bool       `json:"all_cards_studied"` // Congrats message
}

// NextCardResponse provides info about the next card or when it's due
type NextCardResponse struct {
	Card          *Card      `json:"card"`
	Type          string     `json:"type"` // "learning", "review", "new"
	Message       string     `json:"message"`
	NextCardDueAt *time.Time `json:"next_card_due_at"`
	DueCounts     DueCounts  `json:"due_counts"`
}

// QuestionStats provides aggregated statistics for a question across all users
type QuestionStats struct {
	TotalAttempts   int     `json:"total_attempts"`
	AvgEasiness     float64 `json:"avg_easiness"`
	TotalFailures   int     `json:"total_failures"`
	AvgReviews      float64 `json:"avg_reviews"`
	MasteredCount   int     `json:"mastered_count"`
	DifficultyScore float64 `json:"difficulty_score,omitempty"`
}

// DifficultyProgress aggregates progress metrics for a specific difficulty level
type DifficultyProgress struct {
	Total    int `json:"total"`
	Mastered int `json:"mastered"`
	Learning int `json:"learning"`
	Unseen   int `json:"unseen"`
}

// TopicProgress aggregates progress metrics for a specific topic tag
type TopicProgress struct {
	Total    int `json:"total"`
	Mastered int `json:"mastered"`
	Learning int `json:"learning"`
	Unseen   int `json:"unseen"`
}

// ProgressSummary provides a summary of a user's progress broken down by difficulty and topic
type ProgressSummary struct {
	ByDifficulty map[string]DifficultyProgress `json:"by_difficulty"`
	ByTopic      map[string]TopicProgress      `json:"by_topic"`
}

// History represents a saved submission attempt with LLM evaluation
type History struct {
	ID                 string             `json:"id"`
	UserID             string             `json:"user_id"`
	QuestionID         string             `json:"question_id"`
	UserAnswer         string             `json:"user_answer"`
	SubmittedAt        time.Time          `json:"submitted_at"`
	Score              int                `json:"score"`
	Feedback           string             `json:"feedback"`
	CorrectApproach    string             `json:"correct_approach"`
	SubScores          *SubScores         `json:"sub_scores"`
	SolutionBreakdown  *SolutionBreakdown `json:"solution_breakdown"`
	NextReviewAt       time.Time          `json:"next_review_at"`
	CardState          string             `json:"card_state"`
	IntervalMinutes    int                `json:"interval_minutes"`
	IntervalDays       int                `json:"interval_days"`
	TimeSpentSeconds   int                `json:"time_spent_seconds"` // Actual time spent on this card
	CreatedAt          time.Time          `json:"created_at"`
	QuestionTitle      string             `json:"question_title"`
	QuestionLeetcodeID int                `json:"question_leetcode_id"`
	QuestionDifficulty string             `json:"question_difficulty"`
}
