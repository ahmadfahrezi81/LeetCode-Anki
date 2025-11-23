package main

import (
	"fmt"
	"leetcode-anki/backend/config"
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/services"
	"log/slog"
	"os"
	"strconv"

	"github.com/lib/pq"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("🌱 Starting LeetCode Anki Database Seeder (Test 5 problems)...")

	// Load configuration
	if err := config.Load(); err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize LeetCode service
	leetcodeService := services.NewLeetCodeService()

	logger.Info("📥 Fetching 5 random problems from LeetCode API...")

	problems, err := leetcodeService.FetchRandomProblems(2, 2, 1) // 2 Easy, 2 Medium, 1 Hard
	if err != nil {
		logger.Error("Failed to fetch problems", "error", err)
		os.Exit(1)
	}

	logger.Info("✅ Fetched problems", "count", len(problems))

	inserted := 0
	skipped := 0

	for i, problem := range problems {
		status := ""
		err := insertProblem(problem)
		if err != nil {
			if isDuplicateError(err) {
				skipped++
				status = "⚠️ Skipped (duplicate)"
			} else {
				status = fmt.Sprintf("❌ Failed: %v", err)
			}
		} else {
			inserted++
			status = "✅ Inserted"
		}

		logger.Info("Processing problem",
			"index", i+1,
			"total", len(problems),
			"difficulty", problem.Difficulty,
			"title", problem.Title,
			"status", status,
		)
	}

	logger.Info("🎉 Seeding Complete!",
		"inserted", inserted,
		"skipped", skipped,
	)
}

func insertProblem(problem *services.LeetCodeProblem) error {
	topics := make([]string, len(problem.TopicTags))
	for i, tag := range problem.TopicTags {
		topics[i] = tag.Name
	}

	leetcodeID, err := strconv.Atoi(problem.QuestionID)
	if err != nil {
		return fmt.Errorf("invalid question ID: %s", problem.QuestionID)
	}

	descriptionMarkdown := services.StripHTMLTags(problem.Content)
	correctApproach := services.GenerateApproachHint(problem)

	query := `
		INSERT INTO questions 
		(leetcode_id, title, slug, difficulty, description_markdown, topics, correct_approach)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id string
	err = database.DB.QueryRow(
		query,
		leetcodeID,
		problem.Title,
		problem.TitleSlug,
		problem.Difficulty,
		descriptionMarkdown,
		pq.Array(topics),
		correctApproach,
	).Scan(&id)

	return err
}

func isDuplicateError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
