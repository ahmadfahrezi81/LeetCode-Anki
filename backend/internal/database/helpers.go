package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"leetcode-anki/backend/internal/models"
	"strings"

	"github.com/lib/pq"
)

// GetAllQuestionsWithProgress returns all questions with user's progress
func GetAllQuestionsWithProgress(userID string, filters models.QuestionFilters) ([]models.QuestionWithProgress, error) {
	// Build base query
	query := `
		SELECT 
			q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
			q.description_markdown, q.topics, q.created_at,
			r.card_state, r.interval_days, r.next_review_at,
			r.total_reviews, r.total_lapses, r.easiness_factor
		FROM questions q
		LEFT JOIN reviews r ON q.id = r.question_id AND r.user_id = $1
		WHERE 1=1
	`

	args := []interface{}{userID}
	argCount := 1

	// Apply filters
	if filters.Difficulty != "" {
		argCount++
		query += fmt.Sprintf(` AND q.difficulty = $%d`, argCount)
		args = append(args, filters.Difficulty)
	}

	if filters.State != "" {
		if filters.State == "unseen" {
			query += ` AND r.id IS NULL`
		} else {
			argCount++
			query += fmt.Sprintf(` AND r.card_state = $%d`, argCount)
			args = append(args, filters.State)
		}
	}

	if filters.Topic != "" {
		argCount++
		query += fmt.Sprintf(` AND $%d = ANY(q.topics)`, argCount)
		args = append(args, filters.Topic)
	}

	// Apply sorting
	switch filters.SortBy {
	case "difficulty":
		query += ` ORDER BY CASE q.difficulty WHEN 'Easy' THEN 1 WHEN 'Medium' THEN 2 WHEN 'Hard' THEN 3 END`
	case "title":
		query += ` ORDER BY q.title`
	case "progress":
		query += ` ORDER BY r.total_reviews DESC NULLS LAST`
	default:
		query += ` ORDER BY q.leetcode_id`
	}

	// Apply pagination
	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(` LIMIT $%d`, argCount)
		args = append(args, filters.Limit)
	}
	if filters.Offset > 0 {
		argCount++
		query += fmt.Sprintf(` OFFSET $%d`, argCount)
		args = append(args, filters.Offset)
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.QuestionWithProgress
	for rows.Next() {
		var qp models.QuestionWithProgress
		var topics pq.StringArray
		var cardState sql.NullString
		var intervalDays sql.NullInt32
		var nextReviewAt sql.NullTime
		var totalReviews sql.NullInt32
		var totalLapses sql.NullInt32
		var easinessFactor sql.NullFloat64

		err := rows.Scan(
			&qp.Question.ID, &qp.Question.LeetcodeID, &qp.Question.Title,
			&qp.Question.Slug, &qp.Question.Difficulty,
			&qp.Question.DescriptionMarkdown, &topics,
			&qp.Question.CreatedAt,
			&cardState, &intervalDays, &nextReviewAt,
			&totalReviews, &totalLapses, &easinessFactor,
		)
		if err != nil {
			return nil, err
		}

		qp.Question.Topics = topics

		// Set progress info
		if cardState.Valid {
			qp.CardState = cardState.String
			qp.IntervalDays = int(intervalDays.Int32)
			if nextReviewAt.Valid {
				qp.NextReviewAt = &nextReviewAt.Time
			}
			qp.TotalReviews = int(totalReviews.Int32)
			qp.TotalLapses = int(totalLapses.Int32)
			qp.EasinessFactor = easinessFactor.Float64
		} else {
			qp.CardState = "unseen"
		}

		questions = append(questions, qp)
	}

	return questions, nil
}

// GetTotalQuestionCount returns total count based on filters (for pagination)
func GetTotalQuestionCount(filters models.QuestionFilters) (int, error) {
	query := `SELECT COUNT(*) FROM questions WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if filters.Difficulty != "" {
		argCount++
		query += fmt.Sprintf(` AND difficulty = $%d`, argCount)
		args = append(args, filters.Difficulty)
	}

	if filters.Topic != "" {
		argCount++
		query += fmt.Sprintf(` AND $%d = ANY(topics)`, argCount)
		args = append(args, filters.Topic)
	}

	var count int
	err := DB.QueryRow(query, args...).Scan(&count)
	return count, err
}

// GetAllTopics returns all unique topics from questions
func GetAllTopics() ([]string, error) {
	query := `
		SELECT DISTINCT unnest(topics) as topic
		FROM questions
		ORDER BY topic
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}

	return topics, nil
}

// SearchQuestions searches questions by title or description
func SearchQuestions(userID, searchTerm string, limit int) ([]models.QuestionWithProgress, error) {
	query := `
		SELECT 
			q.id, q.leetcode_id, q.title, q.slug, q.difficulty,
			q.description_markdown, q.topics, q.created_at,
			r.card_state, r.interval_days, r.next_review_at,
			r.total_reviews, r.total_lapses, r.easiness_factor
		FROM questions q
		LEFT JOIN reviews r ON q.id = r.question_id AND r.user_id = $1
		WHERE q.title ILIKE $2 OR q.description_markdown ILIKE $2
		ORDER BY 
			CASE 
				WHEN q.title ILIKE $2 THEN 1
				ELSE 2
			END,
			q.leetcode_id
		LIMIT $3
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := DB.Query(query, userID, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.QuestionWithProgress
	for rows.Next() {
		var qp models.QuestionWithProgress
		var topics pq.StringArray
		var cardState sql.NullString
		var intervalDays sql.NullInt32
		var nextReviewAt sql.NullTime
		var totalReviews sql.NullInt32
		var totalLapses sql.NullInt32
		var easinessFactor sql.NullFloat64

		err := rows.Scan(
			&qp.Question.ID, &qp.Question.LeetcodeID, &qp.Question.Title,
			&qp.Question.Slug, &qp.Question.Difficulty,
			&qp.Question.DescriptionMarkdown, &topics,
			&qp.Question.CreatedAt,
			&cardState, &intervalDays, &nextReviewAt,
			&totalReviews, &totalLapses, &easinessFactor,
		)
		if err != nil {
			return nil, err
		}

		qp.Question.Topics = topics

		if cardState.Valid {
			qp.CardState = cardState.String
			qp.IntervalDays = int(intervalDays.Int32)
			if nextReviewAt.Valid {
				qp.NextReviewAt = &nextReviewAt.Time
			}
			qp.TotalReviews = int(totalReviews.Int32)
			qp.TotalLapses = int(totalLapses.Int32)
			qp.EasinessFactor = easinessFactor.Float64
		} else {
			qp.CardState = "unseen"
		}

		questions = append(questions, qp)
	}

	return questions, nil
}

// GetQuestionsByTopic returns questions filtered by topic with user progress
func GetQuestionsByTopic(userID, topic string) ([]models.QuestionWithProgress, error) {
	filters := models.QuestionFilters{
		Topic:  topic,
		SortBy: "leetcode_id",
		Limit:  1000,
	}
	return GetAllQuestionsWithProgress(userID, filters)
}

// GetQuestionsByDifficulty returns questions filtered by difficulty with user progress
func GetQuestionsByDifficulty(userID, difficulty string) ([]models.QuestionWithProgress, error) {
	filters := models.QuestionFilters{
		Difficulty: difficulty,
		SortBy:     "leetcode_id",
		Limit:      1000,
	}
	return GetAllQuestionsWithProgress(userID, filters)
}

// GetQuestionsByState returns questions filtered by review state
func GetQuestionsByState(userID, state string) ([]models.QuestionWithProgress, error) {
	filters := models.QuestionFilters{
		State:  state,
		SortBy: "progress",
		Limit:  1000,
	}
	return GetAllQuestionsWithProgress(userID, filters)
}

// BulkUpdateCardState updates card state for multiple questions (for bulk operations)
func BulkUpdateCardState(userID string, questionIDs []string, newState string) error {
	if len(questionIDs) == 0 {
		return nil
	}

	query := `
		UPDATE reviews
		SET card_state = $1
		WHERE user_id = $2
		AND question_id = ANY($3)
	`

	_, err := DB.Exec(query, newState, userID, pq.Array(questionIDs))
	return err
}

// GetQuestionStats returns statistics about a specific question across all users
func GetQuestionStats(questionID string) (*models.QuestionStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_attempts,
			AVG(easiness_factor) as avg_easiness,
			SUM(total_lapses) as total_failures,
			AVG(total_reviews) as avg_reviews,
			COUNT(*) FILTER (WHERE card_state = 'review' AND interval_days > 21) as mastered_count
		FROM reviews
		WHERE question_id = $1
	`

	var stats models.QuestionStats
	var avgEasiness sql.NullFloat64
	var avgReviews sql.NullFloat64

	err := DB.QueryRow(query, questionID).Scan(
		&stats.TotalAttempts,
		&avgEasiness,
		&stats.TotalFailures,
		&avgReviews,
		&stats.MasteredCount,
	)

	if err != nil {
		return nil, err
	}

	if avgEasiness.Valid {
		stats.AvgEasiness = avgEasiness.Float64
	}
	if avgReviews.Valid {
		stats.AvgReviews = avgReviews.Float64
	}

	// Calculate difficulty score (higher = harder)
	if stats.TotalAttempts > 0 {
		stats.DifficultyScore = float64(stats.TotalFailures) / float64(stats.TotalAttempts)
	}

	return &stats, nil
}

// GetUserProgressSummary returns a summary of user's progress by difficulty and topic
func GetUserProgressSummary(userID string) (*models.ProgressSummary, error) {
	summary := &models.ProgressSummary{
		ByDifficulty: make(map[string]models.DifficultyProgress),
		ByTopic:      make(map[string]models.TopicProgress),
	}

	// Get progress by difficulty
	diffQuery := `
		SELECT 
			q.difficulty,
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE r.card_state = 'review' AND r.interval_days > 21) as mastered,
			COUNT(*) FILTER (WHERE r.card_state IN ('learning', 'relearning')) as learning,
			COUNT(*) FILTER (WHERE r.id IS NULL) as unseen
		FROM questions q
		LEFT JOIN reviews r ON q.id = r.question_id AND r.user_id = $1
		GROUP BY q.difficulty
	`

	rows, err := DB.Query(diffQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var difficulty string
		var progress models.DifficultyProgress
		err := rows.Scan(
			&difficulty,
			&progress.Total,
			&progress.Mastered,
			&progress.Learning,
			&progress.Unseen,
		)
		if err != nil {
			return nil, err
		}
		summary.ByDifficulty[difficulty] = progress
	}

	// Get progress by topic
	topicQuery := `
		SELECT 
			topic,
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE r.card_state = 'review' AND r.interval_days > 21) as mastered,
			COUNT(*) FILTER (WHERE r.card_state IN ('learning', 'relearning')) as learning,
			COUNT(*) FILTER (WHERE r.id IS NULL) as unseen
		FROM (
			SELECT q.id, unnest(q.topics) as topic
			FROM questions q
		) topics
		LEFT JOIN reviews r ON topics.id = r.question_id AND r.user_id = $1
		GROUP BY topic
		ORDER BY total DESC
		LIMIT 20
	`

	rows, err = DB.Query(topicQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic string
		var progress models.TopicProgress
		err := rows.Scan(
			&topic,
			&progress.Total,
			&progress.Mastered,
			&progress.Learning,
			&progress.Unseen,
		)
		if err != nil {
			return nil, err
		}
		summary.ByTopic[topic] = progress
	}

	return summary, nil
}

// BuildFilterQuery is a helper to build dynamic SQL queries with filters
func BuildFilterQuery(baseQuery string, filters map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argCount := 1

	for key, value := range filters {
		if value != nil && value != "" {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", key, argCount))
			args = append(args, value)
			argCount++
		}
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	return baseQuery, args
}

// ============================================
// JSON HELPERS FOR JSONB COLUMNS
// ============================================

// jsonMarshal converts a Go value to JSON bytes for JSONB storage
func jsonMarshal(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return json.Marshal(v)
}

// jsonUnmarshal converts JSON bytes from JSONB to a Go value
func jsonUnmarshal(data []byte, v interface{}) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	return json.Unmarshal(data, v)
}
