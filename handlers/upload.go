package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"planarcomputer/pss-fs/database"
	"planarcomputer/pss-fs/models"
	"planarcomputer/pss-fs/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UploadHandler handles file uploads with signature validation
func UploadHandler(filesDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		signatureParam := c.Params("signature")
		if signatureParam == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Signature is required"})
		}

		log.Printf("Upload attempt with base64 signature: %s", signatureParam)

		// Validate signature (search database with base64 signature directly)
		var uploadSig models.PsUploadSignatures
		result := database.DB.Where("signature = ? AND is_used = false AND expiry > ?", signatureParam, time.Now()).First(&uploadSig)
		if result.Error != nil {
			log.Printf("Signature validation failed: %v", result.Error)

			// Check if signature exists at all
			var existingSig models.PsUploadSignatures
			existsResult := database.DB.Where("signature = ?", signatureParam).First(&existingSig)
			if existsResult.Error != nil {
				log.Printf("Base64 signature does not exist in database")
				return c.Status(401).JSON(fiber.Map{"error": "Invalid signature"})
			}

			// Check if already used
			if existingSig.IsUsed {
				log.Printf("Signature has already been used at: %v", existingSig.UsedAt)
				return c.Status(401).JSON(fiber.Map{"error": "Signature has already been used"})
			}

			// Check if expired
			if existingSig.Expiry.Before(time.Now()) {
				log.Printf("Signature expired at: %v", existingSig.Expiry)
				return c.Status(401).JSON(fiber.Map{"error": "Signature has expired"})
			}

			return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired signature"})
		}

		log.Printf("Signature validated successfully for share_id: %s", uploadSig.ShareId)

		// Get current file count and size for this share
		var share models.PsShares
		if err := database.DB.Where("id = ?", uploadSig.ShareId).First(&share).Error; err != nil {
			log.Printf("Failed to get share info: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to validate share"})
		}

		// Check if adding this file would exceed expected file count
		if share.FileCount+1 > uploadSig.ExpectedFileCount {
			log.Printf("File count limit exceeded: %d/%d", share.FileCount+1, uploadSig.ExpectedFileCount)
			return c.Status(400).JSON(fiber.Map{"error": "File count limit exceeded"})
		}

		// Handle single file upload (matching SvelteKit service)
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Failed to parse multipart form"})
		}

		// Look for 'file' field (singular) instead of 'files'
		files := form.File["file"]
		if len(files) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "No file provided"})
		}

		// Process the single file
		file := files[0] // Take the first (and should be only) file

		// Check if adding this file would exceed expected size (convert to MB)
		// fileSizeMB := (file.Size + 1024*1024 - 1) / (1024 * 1024) // Round up to nearest MB
		// if share.Size+fileSizeMB > uploadSig.ExpectedFileSize {
		// 	log.Printf("File size limit exceeded: %d/%d MB", share.Size+fileSizeMB, uploadSig.ExpectedFileSize)
		// 	return c.Status(400).JSON(fiber.Map{"error": "File size limit exceeded"})
		// }

		// Generate file ID
		fileID := uuid.New()

		// Save file to disk using file ID as filename
		filePath := filepath.Join(filesDir, fileID.String())

		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
		}

		// Calculate file hash using SHA-256
		fileHandle, err := os.Open(filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to open file for hashing"})
		}
		defer fileHandle.Close()

		hasher := sha256.New()
		if _, err := io.Copy(hasher, fileHandle); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to calculate file hash"})
		}
		hash := hex.EncodeToString(hasher.Sum(nil))

		// Create file record
		fileRecord := models.PsFiles{
			ID:       fileID,
			ShareId:  uploadSig.ShareId,
			FileName: file.Filename,
			Mimetype: file.Header.Get("Content-Type"),
			Hash:     hash,
			Size:     file.Size,
		}

		database.DB.Create(&fileRecord)

		// Update share file count and size (increment by 1 file)
		database.DB.Model(&models.PsShares{}).Where("id = ?", uploadSig.ShareId).Updates(map[string]interface{}{
			"file_count": gorm.Expr("file_count + 1"),
			"size":       gorm.Expr("size + ?", file.Size),
		})

		// Update user quota after successful upload
		if err := utils.UpdateUserQuotaByShareID(uploadSig.ShareId); err != nil {
			log.Printf("Warning: Failed to update user quota after upload: %v", err)
			// Don't fail the upload if quota update fails, just log the warning
		}

		return c.JSON(fiber.Map{
			"message": "File uploaded successfully",
			"file":    fileRecord,
		})
	}
}
