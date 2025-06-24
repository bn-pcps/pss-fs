package utils

import (
	"log"
	"time"

	"planarcomputer/pss-fs/database"
	"planarcomputer/pss-fs/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UpdateUserQuota calculates and updates the total quota used by a user
func UpdateUserQuota(userID uuid.UUID) error {
	// Calculate total size of all files for this user across all shares
	var totalUsedBytes int64

	result := database.DB.Raw(`
		SELECT COALESCE(SUM(f.size), 0) as total_size
		FROM ps_files f
		JOIN ps_shares s ON f.share_id = s.id
		WHERE s.user_id = ? AND f.deleted_at IS NULL AND s.deleted_at IS NULL
	`, userID).Scan(&totalUsedBytes)

	if result.Error != nil {
		log.Printf("Error calculating quota for user %s: %v", userID, result.Error)
		return result.Error
	}

	// Convert bytes to MB for storage (as per schema)
	totalUsedMB := (totalUsedBytes + 1024*1024 - 1) / (1024 * 1024)

	// Use ON CONFLICT to update if exists, insert if not
	result = database.DB.Exec(`
		INSERT INTO ps_used_quota (user_id, used_quota, last_updated)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id)
		DO UPDATE SET
			used_quota = EXCLUDED.used_quota,
			last_updated = EXCLUDED.last_updated
	`, userID, totalUsedMB, time.Now())

	if result.Error != nil {
		log.Printf("Error updating quota for user %s: %v", userID, result.Error)
		return result.Error
	}

	log.Printf("Updated quota for user %s: %d MB (%d bytes)", userID, totalUsedMB, totalUsedBytes)
	return nil
}

// UpdateUserQuotaByShareID calculates and updates quota for the user who owns the given share
func UpdateUserQuotaByShareID(shareID uuid.UUID) error {
	// Get the user ID from the share
	var share models.PsShares
	result := database.DB.Select("user_id").Where("id = ?", shareID).First(&share)
	if result.Error != nil {
		log.Printf("Error finding share %s: %v", shareID, result.Error)
		return result.Error
	}

	return UpdateUserQuota(share.UserId)
}

// GetUserQuota retrieves the current quota usage for a user
func GetUserQuota(userID uuid.UUID) (*models.PsUsedQuota, error) {
	var quota models.PsUsedQuota
	result := database.DB.Where("user_id = ?", userID).First(&quota)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Return zero quota if no record exists
			return &models.PsUsedQuota{
				UserId:      userID,
				UsedQuota:   0,
				LastUpdated: time.Now(),
			}, nil
		}
		return nil, result.Error
	}
	return &quota, nil
}
