package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadConfig loads environment variables from a .env file based on APP_ENV
func LoadConfig() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
		log.Println("APP_ENV not set, defaulting to development")
	}

	// Try to find the root directory where .env files are located
	// Starting from the current working directory
	envFile := ".env." + env

	// If we are running from cmd/api, we might need to go up two levels
	// Check if the file exists in the current directory
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// Try going up two levels (common if running from cmd/api)
		envFile = filepath.Join("..", "..", envFile)
	}

	log.Printf("Loading configuration for environment: %s from %s", env, envFile)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("Warning: Could not load %s file, using system environment variables", envFile)
		// Fallback to default .env if specific one not found and it's development
		if env == "development" {
			godotenv.Load() // default .env
		}
	}
}

// GetEnv returns the current APP_ENV
func GetEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return "development"
	}
	return env
}
