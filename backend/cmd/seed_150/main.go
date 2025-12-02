package main

import (
	"fmt"
	"leetcode-anki/backend/config"
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/services"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/lib/pq"
)

// LeetCode 150 problem IDs (Top Interview Questions)
var leetcode150IDs = []int{
	// Array / String
	88, 27, 26, 80, 169, 189, 121, 122, 55, 45, 274, 380, 238, 134, 135, 42,
	13, 12, 58, 14, 151, 6, 28, 68,

	// Two Pointers
	125, 392, 167, 11, 15,

	// Sliding Window
	209, 3, 30, 76,

	// Matrix
	36, 54, 48, 73, 289,

	// Hashmap
	383, 205, 290, 242, 49, 1, 202, 219, 128,

	// Intervals
	228, 56, 57, 452,

	// Stack
	20, 71, 155, 150, 224,

	// Linked List
	141, 2, 21, 138, 92, 25, 19, 82, 61, 86, 146,

	// Binary Tree General
	104, 100, 226, 101, 105, 106, 117, 114, 112, 129, 124, 173, 222, 236,

	// Binary Tree BFS
	199, 637, 102, 103,

	// Binary Search Tree
	530, 230, 98,

	// Graph General
	200, 130, 133, 399, 207, 210,

	// Graph BFS
	909, 433, 127,

	// Trie
	208, 211, 212,

	// Backtracking
	17, 77, 46, 39, 52, 22, 79,

	// Divide & Conquer
	108, 148, 427, 23,

	// Kadane's Algorithm
	53, 918,

	// Binary Search
	35, 74, 162, 33, 34, 153, 4,

	// Heap
	215, 502, 373, 295,

	// Bit Manipulation
	67, 190, 191, 136, 137, 201,

	// Math
	9, 66, 172, 69, 50, 149,

	// 1D DP
	70, 198, 139, 322, 300,

	// Multidimensional DP
	120, 64, 63, 5, 97, 72, 123, 188, 221,
}

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("🌱 Starting LeetCode 150 Database Seeder...")

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

	logger.Info("📥 Building problem ID to slug mapping...")

	// Build ID -> slug mapping first (more efficient than individual lookups)
	mapping, err := leetcodeService.FetchAllProblems()
	if err != nil {
		logger.Error("Failed to build problem mapping", "error", err)
		os.Exit(1)
	}

	logger.Info("✅ Mapping built successfully", "total_problems", len(mapping))
	logger.Info("📥 Fetching LeetCode 150 problems...", "total", len(leetcode150IDs))

	inserted := 0
	skipped := 0
	failed := 0

	for i, problemID := range leetcode150IDs {
		// Get slug from mapping
		slug, ok := mapping[problemID]
		if !ok {
			logger.Error("Problem ID not found in mapping",
				"index", i+1,
				"problem_id", problemID,
			)
			failed++
			continue
		}

		// Fetch problem details using slug
		problem, err := leetcodeService.FetchProblemDetail(slug)
		if err != nil {
			logger.Error("Failed to fetch problem",
				"index", i+1,
				"problem_id", problemID,
				"slug", slug,
				"error", err,
			)
			failed++
			continue
		}

		err = insertLeetCode150Problem(problem)
		status := ""
		if err != nil {
			if isDuplicateLeetCode150Error(err) {
				skipped++
				status = "⚠️ Skipped (duplicate)"
			} else {
				failed++
				status = fmt.Sprintf("❌ Failed: %v", err)
				logger.Error("Failed to insert problem",
					"problem_id", problemID,
					"title", problem.Title,
					"error", err,
				)
			}
		} else {
			inserted++
			status = "✅ Inserted"
		}

		logger.Info("Processing problem",
			"index", i+1,
			"total", len(leetcode150IDs),
			"problem_id", problemID,
			"slug", slug,
			"difficulty", problem.Difficulty,
			"title", problem.Title,
			"status", status,
		)

		// Rate limiting between problem fetches
		if i < len(leetcode150IDs)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	logger.Info("🎉 LeetCode 150 Seeding Complete!",
		"inserted", inserted,
		"skipped", skipped,
		"failed", failed,
		"total", len(leetcode150IDs),
	)

	if failed > 0 {
		logger.Warn("Some problems failed to seed. Check logs above for details.")
		os.Exit(1)
	}
}

func insertLeetCode150Problem(problem *services.LeetCodeProblem) error {
	topics := make([]string, len(problem.TopicTags))
	for i, tag := range problem.TopicTags {
		topics[i] = tag.Name
	}

	leetcodeID, err := strconv.Atoi(problem.QuestionID)
	if err != nil {
		return fmt.Errorf("invalid question ID: %s", problem.QuestionID)
	}

	descriptionMarkdown := services.StripHTMLTags(problem.Content)

	// Note: We are not inserting 'correct_approach' as it is not fetched from the public LeetCode API
	query := `
        INSERT INTO questions 
        (leetcode_id, title, slug, difficulty, description_markdown, topics)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (leetcode_id) 
        DO UPDATE SET 
            description_markdown = EXCLUDED.description_markdown,
            topics = EXCLUDED.topics
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
	).Scan(&id)

	return err
}

func isDuplicateLeetCode150Error(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		// PostgreSQL unique constraint violation error code
		return pqErr.Code == "23505"
	}
	return false
}
