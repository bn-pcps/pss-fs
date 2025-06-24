package handlers

import "github.com/gofiber/fiber/v2"

// HealthHandler handles health check requests
func HealthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "File Service API is running",
		"status":  "healthy",
	})
}
