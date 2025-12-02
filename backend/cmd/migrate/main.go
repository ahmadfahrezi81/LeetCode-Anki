package main

import (
	"fmt"
	"leetcode-anki/backend/config"
	"leetcode-anki/backend/internal/database"
	"log/slog"
	"os"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("🔄 Starting database migration for SRS sub-day intervals...")

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

	// Run migration
	if err := runMigration(); err != nil {
		logger.Error("Migration failed", "error", err)
		os.Exit(1)
	}

	logger.Info("✅ Migration completed successfully!")
}

func runMigration() error {
	// Add new columns
	addColumnsSQL := `
		-- Add interval_minutes for sub-day precision
		ALTER TABLE reviews ADD COLUMN IF NOT EXISTS interval_minutes INTEGER DEFAULT 0;

		-- Add current_step to track learning progression
		ALTER TABLE reviews ADD COLUMN IF NOT EXISTS current_step INTEGER DEFAULT 0;
	`

	if _, err := database.DB.Exec(addColumnsSQL); err != nil {
		return fmt.Errorf("failed to add columns: %w", err)
	}

	slog.Info("✓ Added new columns: interval_minutes, current_step")

	// Migrate existing data: convert interval_days to interval_minutes
	migrateDataSQL := `
		UPDATE reviews 
		SET interval_minutes = interval_days * 1440
		WHERE interval_minutes = 0;
	`

	result, err := database.DB.Exec(migrateDataSQL)
	if err != nil {
		return fmt.Errorf("failed to migrate data: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	slog.Info("✓ Migrated existing data", "rows_affected", rowsAffected)

	// Add index for performance
	addIndexSQL := `
		CREATE INDEX IF NOT EXISTS idx_reviews_next_review 
		ON reviews(user_id, next_review_at);
	`

	if _, err := database.DB.Exec(addIndexSQL); err != nil {
		return fmt.Errorf("failed to add index: %w", err)
	}

	slog.Info("✓ Added index on (user_id, next_review_at)")

	return nil
}
