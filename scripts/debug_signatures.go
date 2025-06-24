package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"planarcomputer/pss-fs/config"
	"planarcomputer/pss-fs/database"
	"planarcomputer/pss-fs/models"

	"github.com/google/uuid"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run scripts/debug_signatures.go list      - List all signatures")
		fmt.Println("  go run scripts/debug_signatures.go create    - Create a test signature")
		fmt.Println("  go run scripts/debug_signatures.go decode <base64_signature> - Decode a signature")
		fmt.Println("  go run scripts/debug_signatures.go check <signature> - Check signature status")
		return
	}

	// Load config and initialize database
	cfg := config.Load()
	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	command := os.Args[1]

	switch command {
	case "list":
		listSignatures()
	case "create":
		createTestSignature()
	case "decode":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a base64 signature to decode")
			return
		}
		decodeSignature(os.Args[2])
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a signature to check")
			return
		}
		checkSignature(os.Args[2])
	default:
		fmt.Println("Unknown command:", command)
	}
}

func listSignatures() {
	var signatures []models.PsUploadSignatures
	result := database.DB.Order("created_at DESC").Limit(10).Find(&signatures)
	if result.Error != nil {
		log.Fatal("Error fetching signatures:", result.Error)
	}

	fmt.Printf("Found %d signatures (showing last 10):\n\n", len(signatures))
	for _, sig := range signatures {
		status := "VALID"
		if sig.IsUsed {
			status = "USED"
		} else if sig.Expiry.Before(time.Now()) {
			status = "EXPIRED"
		}

		fmt.Printf("ID: %s\n", sig.ID)
		fmt.Printf("Share ID: %s\n", sig.ShareId)
		fmt.Printf("Signature: %s\n", sig.Signature)
		fmt.Printf("Status: %s\n", status)
		fmt.Printf("Expiry: %s\n", sig.Expiry.Format("2006-01-02 15:04:05"))
		fmt.Printf("Created: %s\n", sig.CreatedAt.Format("2006-01-02 15:04:05"))
		if sig.IsUsed && sig.UsedAt != nil {
			fmt.Printf("Used At: %s\n", sig.UsedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Println("---")
	}
}

func createTestSignature() {
	// First, create a test share if none exists
	var share models.PsShares
	result := database.DB.First(&share)
	if result.Error != nil {
		// Create a test user first
		testUser := models.PsUsers{
			GoogleId: "test_user_" + uuid.New().String(),
			Name:     "Test User",
			Email:    "test@example.com",
		}
		database.DB.Create(&testUser)

		// Create a test share
		share = models.PsShares{
			UserId:      testUser.ID,
			Title:       "Test Share",
			Description: getStringPtr("Test share for signature testing"),
		}
		database.DB.Create(&share)
		fmt.Printf("Created test share: %s\n", share.ID)
	}

	// Create upload signature
	signature := uuid.New().String() + ":" + uuid.New().String() + ":" + fmt.Sprintf("%d", time.Now().Unix())

	uploadSig := models.PsUploadSignatures{
		ShareId:   share.ID,
		Signature: signature,
		Expiry:    time.Now().Add(24 * time.Hour), // Valid for 24 hours
		IsUsed:    false,
	}

	result = database.DB.Create(&uploadSig)
	if result.Error != nil {
		log.Fatal("Error creating signature:", result.Error)
	}

	fmt.Printf("Created test signature:\n")
	fmt.Printf("Raw: %s\n", signature)
	fmt.Printf("Base64: %s\n", base64.StdEncoding.EncodeToString([]byte(signature)))
	fmt.Printf("Share ID: %s\n", share.ID)
	fmt.Printf("Expiry: %s\n", uploadSig.Expiry.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nTest upload URL: http://localhost:3000/up/%s\n", base64.StdEncoding.EncodeToString([]byte(signature)))
}

func decodeSignature(base64Sig string) {
	decoded, err := base64.StdEncoding.DecodeString(base64Sig)
	if err != nil {
		fmt.Printf("Error decoding base64: %v\n", err)
		return
	}
	fmt.Printf("Decoded signature: %s\n", string(decoded))
}

func checkSignature(signature string) {
	var uploadSig models.PsUploadSignatures
	result := database.DB.Where("signature = ?", signature).First(&uploadSig)
	if result.Error != nil {
		fmt.Printf("Signature not found in database: %v\n", result.Error)
		return
	}

	fmt.Printf("Signature found:\n")
	fmt.Printf("ID: %s\n", uploadSig.ID)
	fmt.Printf("Share ID: %s\n", uploadSig.ShareId)
	fmt.Printf("Is Used: %t\n", uploadSig.IsUsed)
	fmt.Printf("Expiry: %s\n", uploadSig.Expiry.Format("2006-01-02 15:04:05"))
	fmt.Printf("Created: %s\n", uploadSig.CreatedAt.Format("2006-01-02 15:04:05"))

	if uploadSig.IsUsed && uploadSig.UsedAt != nil {
		fmt.Printf("Used At: %s\n", uploadSig.UsedAt.Format("2006-01-02 15:04:05"))
	}

	// Check status
	now := time.Now()
	if uploadSig.IsUsed {
		fmt.Printf("Status: ALREADY USED\n")
	} else if uploadSig.Expiry.Before(now) {
		fmt.Printf("Status: EXPIRED (expired %v ago)\n", now.Sub(uploadSig.Expiry))
	} else {
		fmt.Printf("Status: VALID (expires in %v)\n", uploadSig.Expiry.Sub(now))
	}
}

func getStringPtr(s string) *string {
	return &s
}
