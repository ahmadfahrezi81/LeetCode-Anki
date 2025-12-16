package main

import (
	"leetcode-anki/backend/config"
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/handlers"
	"leetcode-anki/backend/internal/middleware"
	"log/slog"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration (reads .env locally or env vars in prod)
	if err := config.Load(); err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	logger.Info("✅ Database connected successfully")
	defer database.Close()

	// Initialize Gin router
	router := gin.Default()

	// CORS using env value
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.AppConfig.AllowOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	reviewHandler := handlers.NewReviewHandler()
	dashboardHandler := handlers.NewDashboardHandler()
	adminHandler := handlers.NewAdminHandler()
	historyHandler := handlers.NewHistoryHandler()
	transcribeHandler := handlers.NewTranscribeHandler()
	questionsHandler := handlers.NewQuestionsHandler()
	settingsHandler := handlers.NewSettingsHandler()

	// Public routes
	router.GET("/health", healthHandler.HealthCheck)

	// Protected routes (require authentication)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Dashboard
		api.GET("/dashboard", dashboardHandler.GetDashboard)

		// Study session
		api.GET("/card/next", reviewHandler.GetNextCard)
		api.POST("/review/submit", reviewHandler.SubmitAnswer)
		api.POST("/review/skip", reviewHandler.SkipCard)
		api.GET("/review/solution/:questionId", reviewHandler.GetSolutionBreakdown)

		// Questions
		api.GET("/questions/:id", questionsHandler.GetQuestionDetail)

		// History
		api.GET("/history", historyHandler.GetHistory)
		api.GET("/history/:question_id", historyHandler.GetQuestionHistory)

		// Voice transcription
		api.POST("/transcribe", transcribeHandler.TranscribeAudio)

		// Admin endpoints
		api.POST("/admin/refresh-problems", adminHandler.RefreshProblems)
		api.GET("/admin/problem-stats", adminHandler.GetProblemStats)

		// Settings
		api.POST("/settings/limit", settingsHandler.UpdateDailyLimit)
	}

	port := config.AppConfig.ServerPort
	logger.Info("🚀 Server starting", "port", port)

	if err := router.Run(":" + port); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
