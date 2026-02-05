package services

import (
	"context"
	"errors"
	"game-wallet-demo/config"
	"game-wallet-demo/internal/worker"
	"game-wallet-demo/prisma/db"
)

// ProcessTransactionHybrid handles User (Sync) + Treasury (Async)
func ProcessTransactionHybrid(ctx context.Context, client *db.PrismaClient, fromUser string, toUser string, amount int, typeStr string, desc string) error {
	
	// We assume one side is ALWAYS the Treasury in this demo flow (TopUp or Purchase).
	// If User-to-User transfer, we would use the old synchronous method.

	isPurchase := toUser == config.TreasuryID // User spending money
	isTopUp := fromUser == config.TreasuryID  // User getting money

	var userAccountId string
	var balanceChange int // Negative for purchase, Positive for topup

	// Identify the "User" account (The one that needs strict locking)
	if isPurchase {
		// Fetch User Account ID
		acc, err := client.Account.FindFirst(db.Account.UserID.Equals(fromUser)).Exec(ctx)
		if err != nil { return err }
		userAccountId = acc.ID
		
		// Check Balance Sync
		if acc.CachedBalance < amount {
			return errors.New("insufficient funds")
		}
		balanceChange = -amount
	} else if isTopUp {
		acc, err := client.Account.FindFirst(db.Account.UserID.Equals(toUser)).Exec(ctx)
		if err != nil { return err }
		userAccountId = acc.ID
		balanceChange = amount
	} else {
		return errors.New("unsupported transaction type for hybrid flow")
	}

	// --- START SYNC DB TRANSACTION (USER SIDE ONLY) ---
	// We only touch the User's data here. We do NOT touch Treasury rows.
	
	// 1. Create Journal
	journal, err := client.Journal.CreateOne(
		db.Journal.Description.Set(desc),
		db.Journal.Type.Set(typeStr),
	).Exec(ctx)
	if err != nil { return err }

	// 2. Create User's Ledger Entry
	_, err = client.LedgerEntry.CreateOne(
		db.LedgerEntry.Amount.Set(balanceChange),
		db.LedgerEntry.Account.Link(db.Account.ID.Equals(userAccountId)),
		db.LedgerEntry.Journal.Link(db.Journal.ID.Equals(journal.ID)),
	).Exec(ctx)
	if err != nil { return err }

	// 3. Update User Balance (Atomic)
	_, err = client.Account.FindUnique(db.Account.ID.Equals(userAccountId)).Update(
		db.Account.CachedBalance.Increment(balanceChange),
	).Exec(ctx)
	if err != nil { return err }

	// --- END SYNC TRANSACTION ---

	// --- START ASYNC (TREASURY SIDE) ---
	// We fire this off to the worker. The User request returns immediately.
	// The ledger is technically "out of balance" for 500ms, which is acceptable for the Treasury.
	
	treasuryAmount := -balanceChange // If User -10, Treasury +10

	worker.GlobalBatcher.Submit(worker.TreasuryTask{
		Amount:    treasuryAmount,
		JournalID: journal.ID, // Link to the SAME journal
		Desc:      desc,
	})

	return nil
}

// Helper wrappers
func ProcessPurchase(ctx context.Context, client *db.PrismaClient, userId string, item string, cost int) error {
	// 1. Financial Transaction
	err := ProcessTransactionHybrid(ctx, client, userId, config.TreasuryID, cost, "PURCHASE", "Bought "+item)
	if err != nil { return err }

	// 2. Inventory (Ideally this should be in the Sync Transaction above, but kept separate for modularity in this demo)
	update := client.Inventory.FindUnique(db.Inventory.UserID.Equals(userId))
	if item == "gold_coin" {
		update.Update(db.Inventory.GoldCoins.Increment(1)).Exec(ctx)
	} else if item == "treasure_box" {
		update.Update(db.Inventory.TreasureBoxes.Increment(1)).Exec(ctx)
	}
	return nil
}

func ProcessTopUp(ctx context.Context, client *db.PrismaClient, userId string, amount int) error {
	return ProcessTransactionHybrid(ctx, client, config.TreasuryID, userId, amount, "TOPUP", "Bank Deposit")
}

func ProcessBonus(ctx context.Context, client *db.PrismaClient, userId string, amount int, desc string) error {
    // This uses the Hybrid flow: 
    // 1. Instantly credits the User (Sync)
    // 2. Queues the Treasury deduction (Async)
    return ProcessTransactionHybrid(ctx, client, config.TreasuryID, userId, amount, "BONUS", desc)
}