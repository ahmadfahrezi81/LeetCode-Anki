package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort        string
	DatabaseURL       string
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseJWTSecret string
	OpenAIKey         string
	Environment       string
}

var AppConfig *Config

func Load() error {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	AppConfig = &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:   getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseJWTSecret: getEnv("SUPABASE_JWT_SECRET", ""),
		OpenAIKey:         getEnv("OPENAI_API_KEY", ""),
		Environment:       getEnv("ENVIRONMENT", "development"),
	}

	// Validate required fields
	if AppConfig.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if AppConfig.SupabaseURL == "" {
		return fmt.Errorf("SUPABASE_URL is required")
	}
	if AppConfig.SupabaseJWTSecret == "" {
		return fmt.Errorf("SUPABASE_JWT_SECRET is required")
	}
	if AppConfig.OpenAIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
