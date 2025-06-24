package handlers

import (
	"archive/zip"
	"fmt"
	"io"
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

// DownloadFileHandler handles individual file downloads
func DownloadFileHandler(filesDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fileID := c.Params("fileID")
		if fileID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "File ID is required"})
		}

		// Parse UUID
		fileUUID, err := uuid.Parse(fileID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid file ID format"})
		}

		// Get file from database
		var file models.PsFiles
		result := database.DB.Where("id = ? AND deleted_at IS NULL", fileUUID).First(&file)
		if result.Error != nil {
			return c.Status(404).JSON(fiber.Map{"error": "File not found"})
		}

		// Log download analytics
		analytics := models.PsDownloadAnalytics{
			ShareId:   file.ShareId,
			FileId:    &file.ID,
			IpAddress: utils.GetStringPtr(c.IP()),
			UserAgent: utils.GetStringPtr(c.Get("User-Agent")),
		}
		database.DB.Create(&analytics)

		// Update download count for the share
		database.DB.Model(&models.PsShares{}).Where("id = ?", file.ShareId).Update("download_count", gorm.Expr("download_count + 1"))

		// Serve file using file ID as filename
		filePath := filepath.Join(filesDir, file.ID.String())

		// Set original filename in Content-Disposition
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.FileName))
		c.Set("Content-Type", file.Mimetype)

		return c.SendFile(filePath)
	}
}

// DownloadShareHandler handles share downloads (single file or zip)
func DownloadShareHandler(filesDir string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		shareID := c.Params("shareID")
		if shareID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Share ID is required"})
		}

		// Parse UUID
		shareUUID, err := uuid.Parse(shareID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid share ID format"})
		}

		// Get share from database
		var share models.PsShares
		result := database.DB.Where("id = ? AND deleted_at IS NULL", shareUUID).First(&share)
		if result.Error != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Share not found"})
		}

		// Get files in the share
		var files []models.PsFiles
		database.DB.Where("share_id = ? AND deleted_at IS NULL", shareUUID).Find(&files)

		if len(files) == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "No files found in share"})
		}

		// Log download analytics
		analytics := models.PsDownloadAnalytics{
			ShareId:   shareUUID,
			IpAddress: utils.GetStringPtr(c.IP()),
			UserAgent: utils.GetStringPtr(c.Get("User-Agent")),
		}
		database.DB.Create(&analytics)

		// Update download count
		database.DB.Model(&models.PsShares{}).Where("id = ?", shareUUID).Update("download_count", gorm.Expr("download_count + 1"))

		if len(files) == 1 {
			// Single file - serve directly
			file := files[0]
			filePath := filepath.Join(filesDir, file.ID.String())

			// Set original filename in Content-Disposition
			c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.FileName))
			c.Set("Content-Type", file.Mimetype)

			return c.SendFile(filePath)
		}

		// Multiple files - create zip
		return createAndServeZip(c, files, share.Title, filesDir)
	}
}

// createAndServeZip creates a ZIP file from multiple files and serves it
func createAndServeZip(c *fiber.Ctx, files []models.PsFiles, shareTitle string, filesDir string) error {
	tempDir := os.TempDir()
	zipFileName := fmt.Sprintf("share_%s_%d.zip", uuid.New().String(), time.Now().Unix())
	zipPath := filepath.Join(tempDir, zipFileName)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create zip file"})
	}
	defer zipFile.Close()
	defer os.Remove(zipPath) // Clean up temp file

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		// Add file to zip using file ID as filename
		sourcePath := filepath.Join(filesDir, file.ID.String())
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			continue
		}

		zipEntry, err := zipWriter.Create(file.FileName)
		if err != nil {
			sourceFile.Close()
			continue
		}

		_, err = io.Copy(zipEntry, sourceFile)
		sourceFile.Close()
		if err != nil {
			continue
		}
	}

	zipWriter.Close()
	zipFile.Close()

	// Set headers for zip download
	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", shareTitle))

	return c.SendFile(zipPath)
}
