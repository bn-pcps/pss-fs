package main

import (
	"log"
	"math"
	"os"

	"planarcomputer/pss-fs/config"
	"planarcomputer/pss-fs/database"
	"planarcomputer/pss-fs/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Create files directory if it doesn't exist
	if err := os.MkdirAll(cfg.Storage.FilesDirectory, 0755); err != nil {
		log.Fatal("Failed to create files directory:", err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		// BodyLimit: 200 * 1024 * 1024, // 200MB limit (increased from 100MB)
		// BodyLimit: 0, // Unlimited file size
		// BodyLimit: 1024 ^ 100,
		// BodyLimit: 1024 * 1024 * 1024 * 1024 * 1024,
		BodyLimit: int(math.Pow(1024, 10)),
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://localhost:3000, http://127.0.0.1:5173, http://127.0.0.1:3000, https://planarshare.com", // TODO: change to the actual domain
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Origin, Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With",
		AllowCredentials: true,
	}))

	// Main API routes
	app.Post("/up/:signature", handlers.UploadHandler(cfg.Storage.FilesDirectory))
	app.Get("/d/f/:fileID", handlers.DownloadFileHandler(cfg.Storage.FilesDirectory))
	app.Get("/d/s/:shareID", handlers.DownloadShareHandler(cfg.Storage.FilesDirectory))

	// Development and testing routes
	app.Post("/api/generate-signature", handlers.GenerateUploadSignatureHandler)
	app.Post("/api/create-test-share", handlers.CreateTestShareHandler)

	// Health check
	app.Get("/", handlers.HealthHandler)

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Available endpoints:")
	log.Printf("  POST /up/:signature             - Upload files")
	log.Printf("  GET  /d/f/:fileID               - Download file")
	log.Printf("  GET  /d/s/:shareID              - Download share")
	log.Printf("  POST /api/generate-signature    - Generate upload signature")
	log.Printf("  POST /api/create-test-share     - Create test share")
	log.Printf("  GET  /                          - Health check")
	log.Fatal(app.Listen(":" + cfg.Server.Port))
}
