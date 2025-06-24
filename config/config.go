package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Storage  StorageConfig
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host        string
	User        string
	Password    string
	Name        string
	Port        string
	SSLMode     string
	TimeZone    string
	DatabaseURL string
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	FilesDirectory string
	MaxFileSize    string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Try to load from multiple possible env files
	envFiles := []string{".env", "config.env", ".env.local"}

	var loadedFile string
	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			loadedFile = file
			log.Printf("Loaded environment variables from: %s", file)
			break
		}
	}

	if loadedFile == "" {
		log.Println("No .env file found, using system environment variables")
	}

	// Debug: Print what DB_NAME is actually loaded
	dbName := getEnv("DB_NAME", "")
	log.Printf("DB_NAME from environment: '%s'", dbName)

	config := &Config{
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", ""),
			Name:        dbName,
			Port:        getEnv("DB_PORT", "5432"),
			SSLMode:     getEnv("DB_SSLMODE", "disable"),
			TimeZone:    getEnv("DB_TIMEZONE", "UTC"),
			DatabaseURL: getEnv("DATABASE_URL", ""),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "3000"),
		},
		Storage: StorageConfig{
			FilesDirectory: getEnv("FILES_DIRECTORY", "./files"),
			MaxFileSize:    getEnv("MAX_FILE_SIZE", "104857600"), // 100MB
		},
	}

	// Validate required fields
	if config.Database.Name == "" {
		log.Fatal("DB_NAME is required but not set in environment variables")
	}

	return config
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
