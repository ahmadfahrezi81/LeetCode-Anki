package database

import (
	"database/sql"
	"leetcode-anki/backend/internal/models"
	"time"

	"github.com/lib/pq"
)

// GetQuestionByID retrieves a question by ID
func GetQuestionByID(questionID string) (*models.Question, error) {
	query := `
		SELECT id, leetcode_id, title, slug, difficulty, 
		       description_markdown, topics, created_at
		FROM questions
		WHERE id = $1
	`

	var q models.Question
	var topics pq.StringArray

	err := DB.QueryRow(query, questionID).Scan(
		&q.ID, &q.LeetcodeID, &q.Title, &q.Slug, &q.Difficulty,
		&q.DescriptionMarkdown, &topics, &q.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	q.Topics = topics
	return &q, nil
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

	// New cards available (not yet started)
	err = DB.QueryRow(`
		SELECT COUNT(*)
		FROM questions q
		WHERE NOT EXISTS (
			SELECT 1 FROM reviews r
			WHERE r.question_id = q.id
			AND r.user_id = $1
		)
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

	// Reviews done today
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM reviews
		WHERE user_id = $1
		AND DATE(last_reviewed_at) = CURRENT_DATE
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

	// Time spent (approximate - based on review count * average time)
	// This is a placeholder - you'd want to track actual time in the future
	stats.TimeSpentMinutes = (stats.ReviewsDone + stats.NewCardsDone) * 3 // ~3 min per card

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
		       review_cards, mature_cards, updated_at
		FROM user_stats
		WHERE user_id = $1
	`

	var stats models.UserStats
	err := DB.QueryRow(query, userID).Scan(
		&stats.UserID, &stats.TotalCards, &stats.NewCards,
		&stats.LearningCards, &stats.ReviewCards, &stats.MatureCards,
		&stats.UpdatedAt,
	)

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
		INSERT INTO user_stats (user_id, total_cards, new_cards, learning_cards, review_cards, mature_cards)
		VALUES ($1, 0, 0, 0, 0, 0)
		RETURNING user_id, total_cards, new_cards, learning_cards, review_cards, mature_cards, updated_at
	`

	var stats models.UserStats
	err := DB.QueryRow(query, userID).Scan(
		&stats.UserID, &stats.TotalCards, &stats.NewCards,
		&stats.LearningCards, &stats.ReviewCards, &stats.MatureCards,
		&stats.UpdatedAt,
	)

	return &stats, err
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
