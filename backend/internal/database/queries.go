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
		       description_markdown, topics, correct_approach, created_at
		FROM questions
		WHERE id = $1
	`

	var q models.Question
	var topics pq.StringArray

	err := DB.QueryRow(query, questionID).Scan(
		&q.ID, &q.LeetcodeID, &q.Title, &q.Slug, &q.Difficulty,
		&q.DescriptionMarkdown, &topics, &q.CorrectApproach, &q.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	q.Topics = topics
	return &q, nil
}

// GetReview retrieves or creates a review record
func GetReview(userID, questionID string) (*models.Review, error) {
	query := `
		SELECT id, user_id, question_id, card_state, quality, 
		       easiness_factor, interval_days, repetitions,
		       next_review_at, last_reviewed_at, total_reviews, total_lapses, created_at
		FROM reviews
		WHERE user_id = $1 AND question_id = $2
	`

	var r models.Review
	err := DB.QueryRow(query, userID, questionID).Scan(
		&r.ID, &r.UserID, &r.QuestionID, &r.CardState, &r.Quality,
		&r.EasinessFactor, &r.IntervalDays, &r.Repetitions,
		&r.NextReviewAt, &r.LastReviewedAt, &r.TotalReviews, &r.TotalLapses, &r.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No review found (new card)
	}

	if err != nil {
		return nil, err
	}

	return &r, nil
}

// CreateReview inserts a new review record
func CreateReview(review *models.Review) error {
	query := `
		INSERT INTO reviews (user_id, question_id, card_state, quality, 
		                     easiness_factor, interval_days, repetitions,
		                     next_review_at, last_reviewed_at, total_reviews, total_lapses)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	return DB.QueryRow(
		query,
		review.UserID, review.QuestionID, review.CardState, review.Quality,
		review.EasinessFactor, review.IntervalDays, review.Repetitions,
		review.NextReviewAt, review.LastReviewedAt, review.TotalReviews, review.TotalLapses,
	).Scan(&review.ID, &review.CreatedAt)
}

// UpdateReview updates an existing review record
func UpdateReview(review *models.Review) error {
	query := `
		UPDATE reviews
		SET card_state = $1, quality = $2, easiness_factor = $3,
		    interval_days = $4, repetitions = $5, next_review_at = $6,
		    last_reviewed_at = $7, total_reviews = $8, total_lapses = $9
		WHERE id = $10
	`

	_, err := DB.Exec(
		query,
		review.CardState, review.Quality, review.EasinessFactor,
		review.IntervalDays, review.Repetitions, review.NextReviewAt,
		review.LastReviewedAt, review.TotalReviews, review.TotalLapses,
		review.ID,
	)

	return err
}

// GetNextCard retrieves the next card due for review
func GetNextCard(userID string) (*models.Card, error) {
	query := `
		SELECT 
			q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
			q.description_markdown, q.topics, q.correct_approach, q.created_at,
			r.id, r.user_id, r.question_id, r.card_state, r.quality,
			r.easiness_factor, r.interval_days, r.repetitions,
			r.next_review_at, r.last_reviewed_at, r.total_reviews, r.total_lapses, r.created_at
		FROM reviews r
		JOIN questions q ON r.question_id = q.id
		WHERE r.user_id = $1 AND r.next_review_at <= $2
		ORDER BY r.next_review_at ASC
		LIMIT 1
	`

	var card models.Card
	var topics pq.StringArray

	err := DB.QueryRow(query, userID, time.Now()).Scan(
		&card.Question.ID, &card.Question.LeetcodeID, &card.Question.Title,
		&card.Question.Slug, &card.Question.Difficulty, &card.Question.DescriptionMarkdown,
		&topics, &card.Question.CorrectApproach, &card.Question.CreatedAt,
		&card.Review.ID, &card.Review.UserID, &card.Review.QuestionID,
		&card.Review.CardState, &card.Review.Quality, &card.Review.EasinessFactor,
		&card.Review.IntervalDays, &card.Review.Repetitions, &card.Review.NextReviewAt,
		&card.Review.LastReviewedAt, &card.Review.TotalReviews, &card.Review.TotalLapses,
		&card.Review.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No cards due
	}

	if err != nil {
		return nil, err
	}

	card.Question.Topics = topics
	return &card, nil
}

// GetNewCard retrieves a random new question not yet reviewed by the user
func GetNewCard(userID string) (*models.Question, error) {
	query := `
		SELECT q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
		       q.description_markdown, q.topics, q.correct_approach, q.created_at
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
		&q.DescriptionMarkdown, &topics, &q.CorrectApproach, &q.CreatedAt,
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

// GetDueReviewCount counts reviews due today
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
