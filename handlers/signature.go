package handlers

import (
	"encoding/base64"
	"fmt"
	"time"

	"planarcomputer/pss-fs/database"
	"planarcomputer/pss-fs/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GenerateSignatureRequest represents the request body for generating signatures
type GenerateSignatureRequest struct {
	ShareId   string `json:"share_id"`
	ExpiryMin int    `json:"expiry_minutes,omitempty"` // Optional, defaults to 60 minutes
}

// GenerateSignatureResponse represents the response for signature generation
type GenerateSignatureResponse struct {
	Signature    string    `json:"signature"`
	Base64       string    `json:"base64"`
	UploadURL    string    `json:"upload_url"`
	ShareId      string    `json:"share_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	ExpiresInMin int       `json:"expires_in_minutes"`
}

// GenerateUploadSignatureHandler creates a new upload signature
func GenerateUploadSignatureHandler(c *fiber.Ctx) error {
	var req GenerateSignatureRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate share_id
	if req.ShareId == "" {
		return c.Status(400).JSON(fiber.Map{"error": "share_id is required"})
	}

	shareUUID, err := uuid.Parse(req.ShareId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid share_id format"})
	}

	// Check if share exists
	var share models.PsShares
	result := database.DB.Where("id = ? AND deleted_at IS NULL", shareUUID).First(&share)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Share not found"})
	}

	// Set default expiry if not provided
	expiryMinutes := req.ExpiryMin
	if expiryMinutes <= 0 {
		expiryMinutes = 60 // Default to 1 hour
	}

	// Generate signature
	signature := fmt.Sprintf("%s:%s:%d",
		uuid.New().String(),
		uuid.New().String(),
		time.Now().Unix())

	// Create upload signature record
	uploadSig := models.PsUploadSignatures{
		ShareId:   shareUUID,
		Signature: signature,
		Expiry:    time.Now().Add(time.Duration(expiryMinutes) * time.Minute),
		IsUsed:    false,
	}

	result = database.DB.Create(&uploadSig)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create signature"})
	}

	// Encode signature to base64
	base64Sig := base64.StdEncoding.EncodeToString([]byte(signature))

	// Create response
	response := GenerateSignatureResponse{
		Signature:    signature,
		Base64:       base64Sig,
		UploadURL:    fmt.Sprintf("%s://%s/up/%s", c.Protocol(), c.Get("Host"), base64Sig),
		ShareId:      req.ShareId,
		ExpiresAt:    uploadSig.Expiry,
		ExpiresInMin: expiryMinutes,
	}

	return c.JSON(response)
}

// CreateTestShareHandler creates a test share for development
func CreateTestShareHandler(c *fiber.Ctx) error {
	// Create a test user if none exists
	var testUser models.PsUsers
	result := database.DB.Where("email = ?", "test@example.com").First(&testUser)
	if result.Error != nil {
		testUser = models.PsUsers{
			GoogleId: "test_user_" + uuid.New().String(),
			Name:     "Test User",
			Email:    "test@example.com",
		}
		database.DB.Create(&testUser)
	}

	// Create a test share
	share := models.PsShares{
		UserId:      testUser.ID,
		Title:       "Test Share " + time.Now().Format("15:04:05"),
		Description: getStringPtr("Test share created via API"),
	}

	result = database.DB.Create(&share)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create test share"})
	}

	return c.JSON(fiber.Map{
		"message":  "Test share created successfully",
		"share_id": share.ID,
		"title":    share.Title,
	})
}

func getStringPtr(s string) *string {
	return &s
}
