package main

import (
	"fmt"
	"log"

	"planarcomputer/pss-fs/config"
	"planarcomputer/pss-fs/database"
	"planarcomputer/pss-fs/models"
	"planarcomputer/pss-fs/utils"

	"github.com/google/uuid"
)

func main() {
	// Load config and initialize database
	cfg := config.Load()
	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	fmt.Println("=== Quota Testing Script ===")

	// Create a test user if none exists
	var testUser models.PsUsers
	result := database.DB.Where("email = ?", "quota_test@example.com").First(&testUser)
	if result.Error != nil {
		testUser = models.PsUsers{
			GoogleId: "quota_test_user_" + uuid.New().String(),
			Name:     "Quota Test User",
			Email:    "quota_test@example.com",
		}
		database.DB.Create(&testUser)
		fmt.Printf("Created test user: %s (%s)\n", testUser.Name, testUser.ID)
	} else {
		fmt.Printf("Using existing test user: %s (%s)\n", testUser.Name, testUser.ID)
	}

	// Create a test share for this user
	testShare := models.PsShares{
		UserId:      testUser.ID,
		Title:       "Quota Test Share",
		Description: getStringPtr("Test share for quota calculation"),
		FileCount:   0,
		Size:        0,
	}
	database.DB.Create(&testShare)
	fmt.Printf("Created test share: %s (%s)\n", testShare.Title, testShare.ID)

	// Create some test files for this share
	testFiles := []models.PsFiles{
		{
			ShareId:  testShare.ID,
			FileName: "test1.txt",
			Mimetype: "text/plain",
			Hash:     "hash1",
			Size:     1024 * 1024, // 1MB
		},
		{
			ShareId:  testShare.ID,
			FileName: "test2.jpg",
			Mimetype: "image/jpeg",
			Hash:     "hash2",
			Size:     2 * 1024 * 1024, // 2MB
		},
		{
			ShareId:  testShare.ID,
			FileName: "test3.pdf",
			Mimetype: "application/pdf",
			Hash:     "hash3",
			Size:     5 * 1024 * 1024, // 5MB
		},
	}

	for i, file := range testFiles {
		database.DB.Create(&file)
		fmt.Printf("Created test file %d: %s (%d bytes)\n", i+1, file.FileName, file.Size)
	}

	// Update share totals
	totalSize := int64(8 * 1024 * 1024) // 8MB total
	database.DB.Model(&testShare).Updates(map[string]interface{}{
		"file_count": 3,
		"size":       totalSize,
	})

	fmt.Printf("\nUpdated share totals: %d files, %d bytes\n", 3, totalSize)

	// Test quota calculation
	fmt.Println("\n=== Testing Quota Calculation ===")

	// Calculate quota for this user
	err := utils.UpdateUserQuota(testUser.ID)
	if err != nil {
		log.Printf("Error updating quota: %v", err)
	} else {
		fmt.Println("✅ Successfully updated user quota")
	}

	// Retrieve and display the quota
	quota, err := utils.GetUserQuota(testUser.ID)
	if err != nil {
		log.Printf("Error retrieving quota: %v", err)
	} else {
		fmt.Printf("User quota: %d MB (%d bytes calculated)\n", quota.UsedQuota, quota.UsedQuota*1024*1024)
		fmt.Printf("Last updated: %s\n", quota.LastUpdated.Format("2006-01-02 15:04:05"))
	}

	// Test quota calculation by share ID
	fmt.Println("\n=== Testing Quota Update by Share ID ===")
	err = utils.UpdateUserQuotaByShareID(testShare.ID)
	if err != nil {
		log.Printf("Error updating quota by share ID: %v", err)
	} else {
		fmt.Println("✅ Successfully updated user quota by share ID")
	}

	// Show all quota records
	fmt.Println("\n=== All User Quota Records ===")
	var allQuotas []models.PsUsedQuota
	database.DB.Find(&allQuotas)

	if len(allQuotas) == 0 {
		fmt.Println("No quota records found")
	} else {
		for i, q := range allQuotas {
			fmt.Printf("%d. User ID: %s, Used: %d MB, Updated: %s\n",
				i+1, q.UserId, q.UsedQuota, q.LastUpdated.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println("\n=== Cleanup ===")
	// Clean up test data
	database.DB.Where("share_id = ?", testShare.ID).Delete(&models.PsFiles{})
	database.DB.Delete(&testShare)
	database.DB.Where("user_id = ?", testUser.ID).Delete(&models.PsUsedQuota{})
	database.DB.Delete(&testUser)
	fmt.Println("✅ Cleaned up test data")
}

func getStringPtr(s string) *string {
	return &s
}
