package main

import (
	"game-wallet-demo/config"
	"github.com/joho/godotenv"
	"context"
	"fmt"
	"log"
	"game-wallet-demo/prisma/db"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using system environment variables")
	}

	config.LoadConfig()
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Prisma.Disconnect()

	ctx := context.Background()
	treasuryID := "SYSTEM_TREASURY_ID"

	// Check if Treasury exists
	_, err := client.Account.FindUnique(
		db.Account.ID.Equals(treasuryID),
	).Exec(ctx)

	if err != nil {
		// If error is "ErrNotFound" (simplified check), create it
		fmt.Println("⚙️  Creating System Treasury...")

		_, err := client.Account.CreateOne(
			// We don't link a User because it's the System
			db.Account.ID.Set(treasuryID),
			db.Account.Name.Set("System Treasury"),
			db.Account.CachedBalance.Set(2000000000), // 2 Billion Initial Supply
			db.Account.Version.Set(1),
		).Exec(ctx)

		if err != nil {
			log.Fatalf("Failed to seed treasury: %v", err)
		}
		fmt.Println("✅ System Treasury Created!")
	} else {
		fmt.Println("ℹ️  System Treasury already exists.")
	}
}
