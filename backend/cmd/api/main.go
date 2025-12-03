package main

import (
	"leetcode-anki/backend/config"
	"leetcode-anki/backend/internal/database"
	"leetcode-anki/backend/internal/handlers"
	"leetcode-anki/backend/internal/middleware"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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
	logger.Info("✅ Database connected successfully")
	defer database.Close()

	// Initialize Gin router
	router := gin.Default()

	// CORS middleware (adjust origins as needed)
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	reviewHandler := handlers.NewReviewHandler()
	dashboardHandler := handlers.NewDashboardHandler()
	adminHandler := handlers.NewAdminHandler()
	historyHandler := handlers.NewHistoryHandler()
	transcribeHandler := handlers.NewTranscribeHandler()

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

		// History
		api.GET("/history", historyHandler.GetHistory)
		api.GET("/history/:question_id", historyHandler.GetQuestionHistory)

		// Voice transcription
		api.POST("/transcribe", transcribeHandler.TranscribeAudio)

		// Admin endpoints (optional: add auth check)
		api.POST("/admin/refresh-problems", adminHandler.RefreshProblems)
		api.GET("/admin/problem-stats", adminHandler.GetProblemStats)
	}

	port := config.AppConfig.ServerPort
	logger.Info("🚀 Server starting", "port", port)
	if err := router.Run(":" + port); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
