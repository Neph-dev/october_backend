package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	AI       AIConfig
	Market   MarketConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URI string
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level string
}

// AIConfig holds AI/OpenAI configuration
type AIConfig struct {
	OpenAIAPIKey           string
	CustomSearchAPIKey     string
	CustomSearchEngineID   string
}

// MarketConfig holds market data configuration
type MarketConfig struct {
	FinnhubAPIKey string
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, proceeding with existing environment variables")
		// Don't return error - continue with environment variables
	}

	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URI: getEnv("DATABASE_URI", "mongodb://localhost:27017/october"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		AI: AIConfig{
			OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
			CustomSearchAPIKey:   getEnv("CUSTOM_SEARCH_API_KEY", ""),
			CustomSearchEngineID: getEnv("CUSTOM_SEARCH_ENGINE_ID", ""),
		},
		Market: MarketConfig{
			FinnhubAPIKey: getEnv("FINNHUB_API_KEY", ""),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validate ensures all required configuration is present and valid
func (c *Config) validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	// Validate port is a valid number
	if _, err := strconv.Atoi(c.Server.Port); err != nil {
		return fmt.Errorf("invalid server port: %s", c.Server.Port)
	}

	if c.Database.URI == "" {
		return fmt.Errorf("database URI cannot be empty")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[c.Logger.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logger.Level)
	}

	// API keys are optional for development/testing but should be logged as warnings
	if c.AI.OpenAIAPIKey == "" || c.AI.OpenAIAPIKey == "dummy_key" {
		log.Println("Warning: OpenAI API key not set - AI features will be limited")
	}

	if c.AI.CustomSearchAPIKey == "" || c.AI.CustomSearchAPIKey == "dummy_key" {
		log.Println("Warning: Custom Search API key not set - search features will be limited")
	}

	if c.AI.CustomSearchEngineID == "" || c.AI.CustomSearchEngineID == "dummy_id" {
		log.Println("Warning: Custom Search Engine ID not set - search features will be limited")
	}

	if c.Market.FinnhubAPIKey == "" || c.Market.FinnhubAPIKey == "dummy_key" {
		log.Println("Warning: Finnhub API key not set - market data features will be limited")
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv gets a duration from environment variable or returns default
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}