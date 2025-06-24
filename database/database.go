package database

import (
	"fmt"
	"log"

	"planarcomputer/pss-fs/config"
	"planarcomputer/pss-fs/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Initialize sets up the database connection and runs migrations
func Initialize(cfg *config.Config) error {
	var dsn string

	// Use DATABASE_URL if provided, otherwise construct from individual components
	if cfg.Database.DatabaseURL != "" {
		dsn = cfg.Database.DatabaseURL
	} else {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			cfg.Database.Host,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.Port,
			cfg.Database.SSLMode,
			cfg.Database.TimeZone)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")

	// Check if tables already exist (from Drizzle/SvelteKit app)
	var tablesExist bool
	DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'ps_upload_signatures')").Scan(&tablesExist)

	if tablesExist {
		log.Println("Database tables already exist (created by Drizzle/SvelteKit), skipping GORM migrations")
		return nil
	}

	// Only run migrations if tables don't exist
	log.Println("Running GORM migrations...")
	if err := DB.AutoMigrate(
		&models.PsUsers{},
		&models.PsUsedQuota{},
		&models.PsShares{},
		&models.PsUploadSignatures{},
		&models.PsFiles{},
	); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
