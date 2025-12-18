package database

import (
	"database/sql"
	"leetcode-anki/backend/internal/models"
	"strconv"
	"time"

	"github.com/lib/pq"
)

// GetQuestionByID retrieves a question by ID
func GetQuestionByID(questionID string) (*models.Question, error) {
	query := `
		SELECT id, leetcode_id, title, slug, difficulty, 
		       description_markdown, topics, solution_breakdown, created_at
		FROM questions
		WHERE id = $1
	`

	var q models.Question
	var topics pq.StringArray
	var solutionBreakdownJSON []byte

	err := DB.QueryRow(query, questionID).Scan(
		&q.ID, &q.LeetcodeID, &q.Title, &q.Slug, &q.Difficulty,
		&q.DescriptionMarkdown, &topics, &solutionBreakdownJSON, &q.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	q.Topics = topics

	// Unmarshal solution breakdown if it exists
	if len(solutionBreakdownJSON) > 0 {
		if err := jsonUnmarshal(solutionBreakdownJSON, &q.SolutionBreakdown); err != nil {
			// Don't fail the query if solution breakdown is corrupted, just log it
			// log.Printf("Warning: Failed to unmarshal solution breakdown for question %s: %v", questionID, err)
		}
	}

	return &q, nil
}

// UpdateQuestionSolution updates the cached solution breakdown for a question
func UpdateQuestionSolution(questionID string, solution *models.SolutionBreakdown) error {
	query := `
		UPDATE questions
		SET solution_breakdown = $1
		WHERE id = $2
	`

	solutionJSON, err := jsonMarshal(solution)
	if err != nil {
		return err
	}

	_, err = DB.Exec(query, solutionJSON, questionID)
	return err
}

// GetReview retrieves a review record (handles nullable fields properly)
func GetReview(userID, questionID string) (*models.Review, error) {
	query := `
		SELECT id, user_id, question_id, card_state, quality, 
		       easiness_factor, interval_days, interval_minutes, current_step, repetitions,
		       next_review_at, last_reviewed_at, total_reviews, total_lapses, created_at
		FROM reviews
		WHERE user_id = $1 AND question_id = $2
	`

	var r models.Review
	var quality sql.NullInt32
	var lastReviewedAt sql.NullTime

	err := DB.QueryRow(query, userID, questionID).Scan(
		&r.ID, &r.UserID, &r.QuestionID, &r.CardState, &quality,
		&r.EasinessFactor, &r.IntervalDays, &r.IntervalMinutes, &r.CurrentStep, &r.Repetitions,
		&r.NextReviewAt, &lastReviewedAt, &r.TotalReviews, &r.TotalLapses, &r.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No review found (new card)
	}

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if quality.Valid {
		q := int(quality.Int32)
		r.Quality = &q
	}
	if lastReviewedAt.Valid {
		r.LastReviewedAt = &lastReviewedAt.Time
	}

	return &r, nil
}

// CreateReview inserts a new review record
func CreateReview(review *models.Review) error {
	query := `
		INSERT INTO reviews (user_id, question_id, card_state, quality, 
		                     easiness_factor, interval_days, interval_minutes, current_step, repetitions,
		                     next_review_at, last_reviewed_at, total_reviews, total_lapses)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	return DB.QueryRow(
		query,
		review.UserID, review.QuestionID, review.CardState, review.Quality,
		review.EasinessFactor, review.IntervalDays, review.IntervalMinutes, review.CurrentStep, review.Repetitions,
		review.NextReviewAt, review.LastReviewedAt, review.TotalReviews, review.TotalLapses,
	).Scan(&review.ID, &review.CreatedAt)
}

// UpdateReview updates an existing review record
func UpdateReview(review *models.Review) error {
	query := `
		UPDATE reviews
		SET card_state = $1, quality = $2, easiness_factor = $3,
		    interval_days = $4, interval_minutes = $5, current_step = $6,
		    repetitions = $7, next_review_at = $8,
		    last_reviewed_at = $9, total_reviews = $10, total_lapses = $11
		WHERE id = $12
	`

	_, err := DB.Exec(
		query,
		review.CardState, review.Quality, review.EasinessFactor,
		review.IntervalDays, review.IntervalMinutes, review.CurrentStep,
		review.Repetitions, review.NextReviewAt,
		review.LastReviewedAt, review.TotalReviews, review.TotalLapses,
		review.ID,
	)

	return err
}

// ============================================
// ANKI-STYLE PRIORITY QUERIES
// ============================================

// GetNextLearningCard retrieves the next learning/relearning card that's due
// PRIORITY 1: These are cards with intervals < 1 day (short-term memory window)
func GetNextLearningCard(userID string) (*models.Card, error) {
	query := `
		SELECT 
			r.id, r.user_id, r.question_id, r.card_state, r.quality,
			r.easiness_factor, r.interval_days, r.interval_minutes,
			r.current_step, r.repetitions, r.next_review_at,
			r.last_reviewed_at, r.total_reviews, r.total_lapses, r.created_at,
			q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
			q.description_markdown, q.topics, q.created_at
		FROM reviews r
		JOIN questions q ON r.question_id = q.id
		WHERE r.user_id = $1
		AND r.card_state IN ('learning', 'relearning', 'new')
		AND r.next_review_at <= NOW()
		ORDER BY r.next_review_at ASC
		LIMIT 1
	`

	var card models.Card
	var topics pq.StringArray
	var quality sql.NullInt32
	var lastReviewedAt sql.NullTime

	err := DB.QueryRow(query, userID).Scan(
		&card.Review.ID, &card.Review.UserID, &card.Review.QuestionID,
		&card.Review.CardState, &quality, &card.Review.EasinessFactor,
		&card.Review.IntervalDays, &card.Review.IntervalMinutes,
		&card.Review.CurrentStep, &card.Review.Repetitions,
		&card.Review.NextReviewAt, &lastReviewedAt,
		&card.Review.TotalReviews, &card.Review.TotalLapses, &card.Review.CreatedAt,
		&card.Question.ID, &card.Question.LeetcodeID, &card.Question.Title,
		&card.Question.Slug, &card.Question.Difficulty,
		&card.Question.DescriptionMarkdown, &topics,
		&card.Question.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if quality.Valid {
		q := int(quality.Int32)
		card.Review.Quality = &q
	}
	if lastReviewedAt.Valid {
		card.Review.LastReviewedAt = &lastReviewedAt.Time
	}
	card.Question.Topics = topics

	return &card, nil
}

// GetNextReviewCard retrieves the next review card that's due today
// PRIORITY 2: These are graduated cards (intervals >= 1 day)
func GetNextReviewCard(userID string) (*models.Card, error) {
	query := `
		SELECT 
			r.id, r.user_id, r.question_id, r.card_state, r.quality,
			r.easiness_factor, r.interval_days, r.interval_minutes,
			r.current_step, r.repetitions, r.next_review_at,
			r.last_reviewed_at, r.total_reviews, r.total_lapses, r.created_at,
			q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
			q.description_markdown, q.topics, q.created_at
		FROM reviews r
		JOIN questions q ON r.question_id = q.id
		WHERE r.user_id = $1
		AND r.card_state = 'review'
		AND r.next_review_at <= NOW()
		ORDER BY r.next_review_at ASC
		LIMIT 1
	`

	var card models.Card
	var topics pq.StringArray
	var quality sql.NullInt32
	var lastReviewedAt sql.NullTime

	err := DB.QueryRow(query, userID).Scan(
		&card.Review.ID, &card.Review.UserID, &card.Review.QuestionID,
		&card.Review.CardState, &quality, &card.Review.EasinessFactor,
		&card.Review.IntervalDays, &card.Review.IntervalMinutes,
		&card.Review.CurrentStep, &card.Review.Repetitions,
		&card.Review.NextReviewAt, &lastReviewedAt,
		&card.Review.TotalReviews, &card.Review.TotalLapses, &card.Review.CreatedAt,
		&card.Question.ID, &card.Question.LeetcodeID, &card.Question.Title,
		&card.Question.Slug, &card.Question.Difficulty,
		&card.Question.DescriptionMarkdown, &topics,
		&card.Question.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if quality.Valid {
		q := int(quality.Int32)
		card.Review.Quality = &q
	}
	if lastReviewedAt.Valid {
		card.Review.LastReviewedAt = &lastReviewedAt.Time
	}
	card.Question.Topics = topics

	return &card, nil
}

// GetNextNewCardReview retrieves an existing card in 'new' state
// Used to prioritize new cards that have been created but not yet studied
func GetNextNewCardReview(userID string) (*models.Card, error) {
	query := `
		SELECT 
			r.id, r.user_id, r.question_id, r.card_state, r.quality,
			r.easiness_factor, r.interval_days, r.interval_minutes,
			r.current_step, r.repetitions, r.next_review_at,
			r.last_reviewed_at, r.total_reviews, r.total_lapses, r.created_at,
			q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
			q.description_markdown, q.topics, q.created_at
		FROM reviews r
		JOIN questions q ON r.question_id = q.id
		WHERE r.user_id = $1
		AND r.card_state = 'new'
		ORDER BY r.created_at ASC
		LIMIT 1
	`

	var card models.Card
	var topics pq.StringArray
	var quality sql.NullInt32
	var lastReviewedAt sql.NullTime

	err := DB.QueryRow(query, userID).Scan(
		&card.Review.ID, &card.Review.UserID, &card.Review.QuestionID,
		&card.Review.CardState, &quality, &card.Review.EasinessFactor,
		&card.Review.IntervalDays, &card.Review.IntervalMinutes,
		&card.Review.CurrentStep, &card.Review.Repetitions,
		&card.Review.NextReviewAt, &lastReviewedAt,
		&card.Review.TotalReviews, &card.Review.TotalLapses, &card.Review.CreatedAt,
		&card.Question.ID, &card.Question.LeetcodeID, &card.Question.Title,
		&card.Question.Slug, &card.Question.Difficulty,
		&card.Question.DescriptionMarkdown, &topics,
		&card.Question.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if quality.Valid {
		q := int(quality.Int32)
		card.Review.Quality = &q
	}
	if lastReviewedAt.Valid {
		card.Review.LastReviewedAt = &lastReviewedAt.Time
	}
	card.Question.Topics = topics

	return &card, nil
}

// CountNewStateCards counts cards in "new" state
func CountNewStateCards(userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM reviews
		WHERE user_id = $1 AND card_state = 'new'
	`
	var count int
	err := DB.QueryRow(query, userID).Scan(&count)
	return count, err
}

// CountReviewsCreatedToday counts how many reviews were created today
// Used to enforce daily new card limits regardless of queue state
func CountReviewsCreatedToday(userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM reviews
		WHERE user_id = $1
		AND DATE(created_at) = CURRENT_DATE
	`
	var count int
	err := DB.QueryRow(query, userID).Scan(&count)
	return count, err
}

// GetNewCardsStudiedToday counts how many new cards the user has studied today
func GetNewCardsStudiedToday(userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM reviews
		WHERE user_id = $1
		AND card_state != 'new'
		AND DATE(created_at) = CURRENT_DATE
	`

	var count int
	err := DB.QueryRow(query, userID).Scan(&count)
	return count, err
}

// GetNextDueCardTime returns when the next card will be due
func GetNextDueCardTime(userID string) (*time.Time, error) {
	query := `
		SELECT next_review_at
		FROM reviews
		WHERE user_id = $1
		AND next_review_at > NOW()
		ORDER BY next_review_at ASC
		LIMIT 1
	`

	var nextTime time.Time
	err := DB.QueryRow(query, userID).Scan(&nextTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &nextTime, nil
}

// GetDueCountsByType returns counts for learning, review, and new cards
func GetDueCountsByType(userID string) (*models.DueCounts, error) {
	counts := &models.DueCounts{}

	// Learning cards due now
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM reviews
		WHERE user_id = $1
		AND card_state IN ('learning', 'relearning', 'new')
		AND next_review_at <= NOW()
	`, userID).Scan(&counts.LearningDue)
	if err != nil {
		return nil, err
	}

	// Review cards due today
	err = DB.QueryRow(`
		SELECT COUNT(*)
		FROM reviews
		WHERE user_id = $1
		AND card_state = 'review'
		AND next_review_at <= NOW()
	`, userID).Scan(&counts.ReviewsDue)
	if err != nil {
		return nil, err
	}

	// New cards available (cards in "new" state)
	err = DB.QueryRow(`
		SELECT COUNT(*)
		FROM reviews
		WHERE user_id = $1 AND card_state = 'new'
	`, userID).Scan(&counts.NewAvailable)
	if err != nil {
		return nil, err
	}

	// New cards studied today
	newToday, err := GetNewCardsStudiedToday(userID)
	if err != nil {
		return nil, err
	}
	counts.NewStudiedToday = newToday

	return counts, nil
}

// GetTodayStats returns today's study statistics
func GetTodayStats(userID string) (*models.TodayStats, error) {
	stats := &models.TodayStats{}

	// Reviews done today - count all history entries (attempts) instead of unique cards
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM history
		WHERE user_id = $1
		AND DATE(submitted_at) = CURRENT_DATE
	`, userID).Scan(&stats.ReviewsDone)
	if err != nil {
		return nil, err
	}

	// New cards studied today
	newToday, err := GetNewCardsStudiedToday(userID)
	if err != nil {
		return nil, err
	}
	stats.NewCardsDone = newToday

	// Time spent - calculate from actual history records
	var totalSeconds sql.NullInt64
	err = DB.QueryRow(`
		SELECT COALESCE(SUM(time_spent_seconds), 0)
		FROM history
		WHERE user_id = $1
		AND DATE(submitted_at) = CURRENT_DATE
	`, userID).Scan(&totalSeconds)
	if err != nil {
		return nil, err
	}

	// Convert seconds to minutes (round up) and multiply by 2
	if totalSeconds.Valid {
		stats.TimeSpentMinutes = int((totalSeconds.Int64+59)/60) * 2 // Round up and multiply by 2
	} else {
		stats.TimeSpentMinutes = 0
	}

	return stats, nil
}

// ============================================
// LEGACY FUNCTIONS (Keep for compatibility)
// ============================================

// GetNextCard retrieves the next card due for review (OLD - for backward compatibility)
// Use GetNextLearningCard() and GetNextReviewCard() instead for Anki priority
func GetNextCard(userID string) (*models.Card, error) {
	query := `
		SELECT 
			q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
			q.description_markdown, q.topics, q.created_at,
			r.id, r.user_id, r.question_id, r.card_state, r.quality,
			r.easiness_factor, r.interval_days, r.interval_minutes, r.current_step, r.repetitions,
			r.next_review_at, r.last_reviewed_at, r.total_reviews, r.total_lapses, r.created_at
		FROM reviews r
		JOIN questions q ON r.question_id = q.id
		WHERE r.user_id = $1 AND r.next_review_at <= $2
		ORDER BY r.next_review_at ASC
		LIMIT 1
	`

	var card models.Card
	var topics pq.StringArray
	var quality sql.NullInt32
	var lastReviewedAt sql.NullTime

	err := DB.QueryRow(query, userID, time.Now()).Scan(
		&card.Question.ID, &card.Question.LeetcodeID, &card.Question.Title,
		&card.Question.Slug, &card.Question.Difficulty, &card.Question.DescriptionMarkdown,
		&topics, &card.Question.CreatedAt,
		&card.Review.ID, &card.Review.UserID, &card.Review.QuestionID,
		&card.Review.CardState, &quality, &card.Review.EasinessFactor,
		&card.Review.IntervalDays, &card.Review.IntervalMinutes, &card.Review.CurrentStep,
		&card.Review.Repetitions, &card.Review.NextReviewAt,
		&lastReviewedAt, &card.Review.TotalReviews, &card.Review.TotalLapses,
		&card.Review.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No cards due
	}

	if err != nil {
		return nil, err
	}

	if quality.Valid {
		q := int(quality.Int32)
		card.Review.Quality = &q
	}
	if lastReviewedAt.Valid {
		card.Review.LastReviewedAt = &lastReviewedAt.Time
	}
	card.Question.Topics = topics
	return &card, nil
}

// GetNewCard retrieves a random new question not yet reviewed by the user
func GetNewCard(userID string) (*models.Question, error) {
	query := `
		SELECT q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
		       q.description_markdown, q.topics, q.created_at
		FROM questions q
		WHERE NOT EXISTS (
			SELECT 1 FROM reviews r 
			WHERE r.user_id = $1 AND r.question_id = q.id
		)
		ORDER BY RANDOM()
		LIMIT 1
	`

	var q models.Question
	var topics pq.StringArray

	err := DB.QueryRow(query, userID).Scan(
		&q.ID, &q.LeetcodeID, &q.Title, &q.Slug, &q.Difficulty,
		&q.DescriptionMarkdown, &topics, &q.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No new cards available
	}

	if err != nil {
		return nil, err
	}

	q.Topics = topics
	return &q, nil
}

// GetDueReviewCount counts reviews due today (OLD - for backward compatibility)
func GetDueReviewCount(userID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM reviews 
		WHERE user_id = $1 AND next_review_at <= $2
	`

	var count int
	err := DB.QueryRow(query, userID, time.Now()).Scan(&count)
	return count, err
}

// GetUserStats retrieves or creates user statistics
func GetUserStats(userID string) (*models.UserStats, error) {
	query := `
		SELECT user_id, total_cards, new_cards, learning_cards, 
		       review_cards, mature_cards, new_cards_limit, coins,
		       current_streak, max_streak, last_streak_date, updated_at
		FROM user_stats
		WHERE user_id = $1
	`

	var stats models.UserStats
	var lastStreakDate sql.NullTime
	err := DB.QueryRow(query, userID).Scan(
		&stats.UserID, &stats.TotalCards, &stats.NewCards,
		&stats.LearningCards, &stats.ReviewCards, &stats.MatureCards,
		&stats.NewCardsLimit, &stats.Coins,
		&stats.CurrentStreak, &stats.MaxStreak, &lastStreakDate, &stats.UpdatedAt,
	)

	if lastStreakDate.Valid {
		stats.LastStreakDate = &lastStreakDate.Time
	}

	if err == sql.ErrNoRows {
		// Create initial stats
		return CreateUserStats(userID)
	}

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// CreateUserStats initializes stats for a new user
func CreateUserStats(userID string) (*models.UserStats, error) {
	query := `
		INSERT INTO user_stats (user_id, total_cards, new_cards, learning_cards, review_cards, mature_cards, new_cards_limit, coins, current_streak, max_streak)
		VALUES ($1, 0, 0, 0, 0, 0, 5, 0, 0, 0)
		RETURNING user_id, total_cards, new_cards, learning_cards, review_cards, mature_cards, new_cards_limit, coins, current_streak, max_streak, last_streak_date, updated_at
	`

	var stats models.UserStats
	var lastStreakDate sql.NullTime
	err := DB.QueryRow(query, userID).Scan(
		&stats.UserID, &stats.TotalCards, &stats.NewCards,
		&stats.LearningCards, &stats.ReviewCards, &stats.MatureCards,
		&stats.NewCardsLimit, &stats.Coins,
		&stats.CurrentStreak, &stats.MaxStreak, &lastStreakDate, &stats.UpdatedAt,
	)

	if lastStreakDate.Valid {
		stats.LastStreakDate = &lastStreakDate.Time
	}

	return &stats, err
}

// UpdateUserStreak handles incrementing or resetting a user's daily streak
func UpdateUserStreak(userID string) (int, error) {
	// 1. Get current stats
	stats, err := GetUserStats(userID)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	// Use explicit DATE comparison (truncating time)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// If already updated today, do nothing
	if stats.LastStreakDate != nil {
		lastDate := time.Date(stats.LastStreakDate.Year(), stats.LastStreakDate.Month(), stats.LastStreakDate.Day(), 0, 0, 0, 0, stats.LastStreakDate.Location())
		if lastDate.Equal(today) {
			return stats.CurrentStreak, nil
		}
	}

	newStreak := 1
	if stats.LastStreakDate != nil {
		yesterday := today.AddDate(0, 0, -1)
		lastDate := time.Date(stats.LastStreakDate.Year(), stats.LastStreakDate.Month(), stats.LastStreakDate.Day(), 0, 0, 0, 0, stats.LastStreakDate.Location())
		if lastDate.Equal(yesterday) {
			newStreak = stats.CurrentStreak + 1
		}
	}

	newMaxStreak := stats.MaxStreak
	if newStreak > stats.MaxStreak {
		newMaxStreak = newStreak
	}

	query := `
		UPDATE user_stats
		SET current_streak = $1,
			max_streak = $2,
			last_streak_date = $3,
			updated_at = NOW()
		WHERE user_id = $4
	`
	_, err = DB.Exec(query, newStreak, newMaxStreak, today, userID)
	return newStreak, err
}

// UpdateUserLimit updates the daily new card limit for a user
func UpdateUserLimit(userID string, limit int) error {
	query := `
		UPDATE user_stats
		SET new_cards_limit = $2, updated_at = NOW()
		WHERE user_id = $1
	`
	// Ensure stats exist first
	if _, err := GetUserStats(userID); err != nil {
		return err
	}

	_, err := DB.Exec(query, userID, limit)
	return err
}

// GetUnusedProblemCount counts problems not yet reviewed by user
func GetUnusedProblemCount(userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM questions q
		WHERE NOT EXISTS (
			SELECT 1 FROM reviews r
			WHERE r.user_id = $1 AND r.question_id = q.id
		)
	`

	var count int
	err := DB.QueryRow(query, userID).Scan(&count)
	return count, err
}

// RefreshUserStats recalculates user statistics from reviews
func RefreshUserStats(userID string) error {
	query := `
		INSERT INTO user_stats (user_id, total_cards, new_cards, learning_cards, review_cards, mature_cards, updated_at)
		SELECT 
			$1,
			COUNT(*),
			SUM(CASE WHEN card_state = 'new' THEN 1 ELSE 0 END),
			SUM(CASE WHEN card_state IN ('learning', 'relearning') THEN 1 ELSE 0 END),
			SUM(CASE WHEN card_state = 'review' AND interval_days <= 21 THEN 1 ELSE 0 END),
			SUM(CASE WHEN card_state = 'review' AND interval_days > 21 THEN 1 ELSE 0 END),
			NOW()
		FROM reviews
		WHERE user_id = $1
		ON CONFLICT (user_id) 
		DO UPDATE SET
			total_cards = EXCLUDED.total_cards,
			new_cards = EXCLUDED.new_cards,
			learning_cards = EXCLUDED.learning_cards,
			review_cards = EXCLUDED.review_cards,
			mature_cards = EXCLUDED.mature_cards,
			updated_at = EXCLUDED.updated_at
	`

	_, err := DB.Exec(query, userID)
	return err
}

// ============================================
// HISTORY FUNCTIONS
// ============================================

// CreateHistory saves a submission attempt to history
func CreateHistory(history *models.History) error {
	query := `
		INSERT INTO history (
			user_id, question_id, user_answer, submitted_at,
			score, feedback, correct_approach,
			sub_scores, solution_breakdown,
			next_review_at, card_state, interval_minutes, interval_days, time_spent_seconds
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at
	`

	// Convert SubScores and SolutionBreakdown to JSON
	subScoresJSON, err := jsonMarshal(history.SubScores)
	if err != nil {
		return err
	}

	solutionBreakdownJSON, err := jsonMarshal(history.SolutionBreakdown)
	if err != nil {
		return err
	}

	return DB.QueryRow(
		query,
		history.UserID,
		history.QuestionID,
		history.UserAnswer,
		history.SubmittedAt,
		history.Score,
		history.Feedback,
		history.CorrectApproach,
		subScoresJSON,
		solutionBreakdownJSON,
		history.NextReviewAt,
		history.CardState,
		history.IntervalMinutes,
		history.IntervalDays,
		history.TimeSpentSeconds,
	).Scan(&history.ID, &history.CreatedAt)
}

// GetHistoryByUser retrieves all history for a user (paginated with optional filters)
func GetHistoryByUser(userID string, limit, offset int, difficulties []string, minScore, maxScore *int, states []string) ([]models.History, error) {
	// Build dynamic query with filters
	query := `
		SELECT 
			h.id, h.user_id, h.question_id, h.user_answer, h.submitted_at,
			h.score, h.feedback, h.correct_approach,
			h.sub_scores, h.solution_breakdown,
			h.next_review_at, h.card_state, h.interval_minutes, h.interval_days,
			h.time_spent_seconds, h.created_at,
			q.title, q.leetcode_id, q.difficulty
		FROM history h
		JOIN questions q ON h.question_id = q.id
		WHERE h.user_id = $1
	`

	args := []interface{}{userID}
	paramCount := 1

	// Add difficulty filter
	if len(difficulties) > 0 {
		paramCount++
		query += " AND q.difficulty = ANY($" + strconv.Itoa(paramCount) + ")"
		args = append(args, pq.Array(difficulties))
	}

	// Add score filters
	if minScore != nil {
		paramCount++
		query += " AND h.score >= $" + strconv.Itoa(paramCount)
		args = append(args, *minScore)
	}
	if maxScore != nil {
		paramCount++
		query += " AND h.score <= $" + strconv.Itoa(paramCount)
		args = append(args, *maxScore)
	}

	// Add state filter
	if len(states) > 0 {
		paramCount++
		query += " AND h.card_state = ANY($" + strconv.Itoa(paramCount) + ")"
		args = append(args, pq.Array(states))
	}

	// Add ordering and pagination
	query += " ORDER BY h.submitted_at DESC"
	paramCount++
	query += " LIMIT $" + strconv.Itoa(paramCount)
	args = append(args, limit)
	paramCount++
	query += " OFFSET $" + strconv.Itoa(paramCount)
	args = append(args, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []models.History
	for rows.Next() {
		var h models.History
		var subScoresJSON, solutionBreakdownJSON []byte

		err := rows.Scan(
			&h.ID, &h.UserID, &h.QuestionID, &h.UserAnswer, &h.SubmittedAt,
			&h.Score, &h.Feedback, &h.CorrectApproach,
			&subScoresJSON, &solutionBreakdownJSON,
			&h.NextReviewAt, &h.CardState, &h.IntervalMinutes, &h.IntervalDays,
			&h.TimeSpentSeconds, &h.CreatedAt,
			&h.QuestionTitle, &h.QuestionLeetcodeID, &h.QuestionDifficulty,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if err := jsonUnmarshal(subScoresJSON, &h.SubScores); err != nil {
			return nil, err
		}
		if err := jsonUnmarshal(solutionBreakdownJSON, &h.SolutionBreakdown); err != nil {
			return nil, err
		}

		histories = append(histories, h)
	}

	return histories, rows.Err()
}

// GetHistoryByQuestion retrieves all attempts for a specific question by a user
func GetHistoryByQuestion(userID, questionID string) ([]models.History, error) {
	query := `
		SELECT 
			h.id, h.user_id, h.question_id, h.user_answer, h.submitted_at,
			h.score, h.feedback, h.correct_approach,
			h.sub_scores, h.solution_breakdown,
			h.next_review_at, h.card_state, h.interval_minutes, h.interval_days,
			h.time_spent_seconds, h.created_at,
			q.title, q.leetcode_id, q.difficulty
		FROM history h
		JOIN questions q ON h.question_id = q.id
		WHERE h.user_id = $1 AND h.question_id = $2
		ORDER BY h.submitted_at DESC
	`

	rows, err := DB.Query(query, userID, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []models.History
	for rows.Next() {
		var h models.History
		var subScoresJSON, solutionBreakdownJSON []byte

		err := rows.Scan(
			&h.ID, &h.UserID, &h.QuestionID, &h.UserAnswer, &h.SubmittedAt,
			&h.Score, &h.Feedback, &h.CorrectApproach,
			&subScoresJSON, &solutionBreakdownJSON,
			&h.NextReviewAt, &h.CardState, &h.IntervalMinutes, &h.IntervalDays,
			&h.TimeSpentSeconds, &h.CreatedAt,
			&h.QuestionTitle, &h.QuestionLeetcodeID, &h.QuestionDifficulty,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if err := jsonUnmarshal(subScoresJSON, &h.SubScores); err != nil {
			return nil, err
		}
		if err := jsonUnmarshal(solutionBreakdownJSON, &h.SolutionBreakdown); err != nil {
			return nil, err
		}

		histories = append(histories, h)
	}

	return histories, rows.Err()
}

// GetLatestAttempt retrieves the most recent attempt for a question
func GetLatestAttempt(userID, questionID string) (*models.History, error) {
	query := `
		SELECT 
			h.id, h.user_id, h.question_id, h.user_answer, h.submitted_at,
			h.score, h.feedback, h.correct_approach,
			h.sub_scores, h.solution_breakdown,
			h.next_review_at, h.card_state, h.interval_minutes, h.interval_days,
			h.time_spent_seconds, h.created_at,
			q.title, q.leetcode_id, q.difficulty
		FROM history h
		JOIN questions q ON h.question_id = q.id
		WHERE h.user_id = $1 AND h.question_id = $2
		ORDER BY h.submitted_at DESC
		LIMIT 1
	`

	var h models.History
	var subScoresJSON, solutionBreakdownJSON []byte

	err := DB.QueryRow(query, userID, questionID).Scan(
		&h.ID, &h.UserID, &h.QuestionID, &h.UserAnswer, &h.SubmittedAt,
		&h.Score, &h.Feedback, &h.CorrectApproach,
		&subScoresJSON, &solutionBreakdownJSON,
		&h.NextReviewAt, &h.CardState, &h.IntervalMinutes, &h.IntervalDays,
		&h.TimeSpentSeconds, &h.CreatedAt,
		&h.QuestionTitle, &h.QuestionLeetcodeID, &h.QuestionDifficulty,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if err := jsonUnmarshal(subScoresJSON, &h.SubScores); err != nil {
		return nil, err
	}
	if err := jsonUnmarshal(solutionBreakdownJSON, &h.SolutionBreakdown); err != nil {
		return nil, err
	}

	return &h, nil
}
