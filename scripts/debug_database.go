package main

import (
	"fmt"
	"log"

	"planarcomputer/pss-fs/config"
	"planarcomputer/pss-fs/database"
)

func main() {
	// Load config and initialize database
	cfg := config.Load()

	fmt.Println("=== Database Configuration ===")
	fmt.Printf("Host: %s\n", cfg.Database.Host)
	fmt.Printf("User: %s\n", cfg.Database.User)
	fmt.Printf("Database: %s\n", cfg.Database.Name)
	fmt.Printf("Port: %s\n", cfg.Database.Port)
	fmt.Printf("SSLMode: %s\n", cfg.Database.SSLMode)
	fmt.Printf("TimeZone: %s\n", cfg.Database.TimeZone)
	fmt.Println()

	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Get the current database name
	var currentDB string
	database.DB.Raw("SELECT current_database()").Scan(&currentDB)
	fmt.Printf("=== Currently Connected Database ===\n")
	fmt.Printf("Database Name: %s\n", currentDB)
	fmt.Println()

	// List all tables in the current database
	fmt.Println("=== Tables in Current Database ===")
	var tables []string
	database.DB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename").Scan(&tables)

	if len(tables) == 0 {
		fmt.Println("No tables found in public schema!")
	} else {
		for i, table := range tables {
			fmt.Printf("%d. %s\n", i+1, table)
		}
	}
	fmt.Println()

	// Check if our expected tables exist
	expectedTables := []string{"ps_upload_signatures", "ps_shares", "ps_users", "ps_files"}
	fmt.Println("=== Expected Tables Check ===")
	for _, expectedTable := range expectedTables {
		var exists bool
		database.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", expectedTable).Scan(&exists)

		status := "❌ MISSING"
		if exists {
			status = "✅ EXISTS"

			// Count records if table exists
			var count int64
			database.DB.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", expectedTable)).Scan(&count)
			status += fmt.Sprintf(" (%d records)", count)
		}

		fmt.Printf("%s: %s\n", expectedTable, status)
	}
	fmt.Println()

	// If ps_upload_signatures exists, show some records
	var sigTableExists bool
	database.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'ps_upload_signatures')").Scan(&sigTableExists)

	if sigTableExists {
		fmt.Println("=== Recent Upload Signatures ===")
		type SignatureInfo struct {
			ID        string `json:"id"`
			ShareId   string `json:"share_id"`
			Signature string `json:"signature"`
			IsUsed    bool   `json:"is_used"`
			Expiry    string `json:"expiry"`
		}

		var signatures []SignatureInfo
		database.DB.Raw("SELECT id, share_id, LEFT(signature, 50) || '...' as signature, is_used, expiry FROM ps_upload_signatures ORDER BY created_at DESC LIMIT 5").Scan(&signatures)

		if len(signatures) == 0 {
			fmt.Println("No upload signatures found!")
		} else {
			for i, sig := range signatures {
				fmt.Printf("%d. ID: %s\n", i+1, sig.ID)
				fmt.Printf("   Share: %s\n", sig.ShareId)
				fmt.Printf("   Signature: %s\n", sig.Signature)
				fmt.Printf("   Used: %t\n", sig.IsUsed)
				fmt.Printf("   Expiry: %s\n", sig.Expiry)
				fmt.Println()
			}
		}
	}

	// Show connection string (with password masked)
	dsn := fmt.Sprintf("host=%s user=%s password=*** dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
		cfg.Database.TimeZone)
	fmt.Printf("=== Connection String ===\n")
	fmt.Printf("%s\n", dsn)
}
