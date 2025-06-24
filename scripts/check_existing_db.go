package main

import (
	"fmt"
	"log"

	"planarcomputer/pss-fs/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg := config.Load()

	// Connect to database WITHOUT running migrations
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
		cfg.Database.TimeZone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Successfully connected to database!")
	fmt.Println()

	// Get the current database name
	var currentDB string
	db.Raw("SELECT current_database()").Scan(&currentDB)
	fmt.Printf("=== Currently Connected Database ===\n")
	fmt.Printf("Database Name: %s\n", currentDB)
	fmt.Println()

	// List all tables in the current database
	fmt.Println("=== Tables in Current Database ===")
	var tables []string
	db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename").Scan(&tables)

	if len(tables) == 0 {
		fmt.Println("No tables found in public schema!")
	} else {
		for i, table := range tables {
			fmt.Printf("%d. %s\n", i+1, table)
		}
	}
	fmt.Println()

	// Check specifically for ps_upload_signatures
	var sigTableExists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'ps_upload_signatures')").Scan(&sigTableExists)

	if sigTableExists {
		fmt.Println("✅ ps_upload_signatures table EXISTS!")

		// Count records
		var count int64
		db.Raw("SELECT COUNT(*) FROM ps_upload_signatures").Scan(&count)
		fmt.Printf("Record count: %d\n", count)

		if count > 0 {
			fmt.Println("\n=== Recent Upload Signatures ===")
			type SignatureInfo struct {
				ID        string `json:"id"`
				ShareId   string `json:"share_id"`
				Signature string `json:"signature"`
				IsUsed    bool   `json:"is_used"`
				Expiry    string `json:"expiry"`
			}

			var signatures []SignatureInfo
			db.Raw("SELECT id, share_id, LEFT(signature, 50) || '...' as signature, is_used, expiry FROM ps_upload_signatures ORDER BY created_at DESC LIMIT 3").Scan(&signatures)

			for i, sig := range signatures {
				fmt.Printf("%d. ID: %s\n", i+1, sig.ID)
				fmt.Printf("   Share: %s\n", sig.ShareId)
				fmt.Printf("   Signature: %s\n", sig.Signature)
				fmt.Printf("   Used: %t\n", sig.IsUsed)
				fmt.Printf("   Expiry: %s\n", sig.Expiry)
				fmt.Println()
			}
		}
	} else {
		fmt.Println("❌ ps_upload_signatures table does NOT exist")
	}
}
