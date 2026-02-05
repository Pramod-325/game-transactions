package handlers

import (
	"context"
	"game-wallet-demo/internal/services"
	"game-wallet-demo/prisma/db"
	"github.com/gin-gonic/gin"
)

// GetBalance (Read-Only, no changes needed)
func (h *Handler) GetBalance(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	
	user, err := h.Client.User.FindUnique(db.User.ID.Equals(userId)).With(
		db.User.Inventory.Fetch(),
		db.User.Account.Fetch(),
	).Exec(context.Background())

	if err != nil {
		c.JSON(500, gin.H{"error": "Fetch failed"})
		return
	}

	account, okAccount := user.Account()
	inventory, okInventory := user.Inventory()

	balance := 0
	if okAccount { balance = account.CachedBalance }
	
	coins, boxes := 0, 0
	if okInventory {
		coins = inventory.GoldCoins
		boxes = inventory.TreasureBoxes
	}

	c.JSON(200, gin.H{
		"username": user.Username,
		"balance":  balance,
		"referralCode": user.ReferralCode,
		"inventory": gin.H{"goldCoins": coins, "treasureBoxes": boxes},
	})
}

// TopUp -> Uses Async Service
func (h *Handler) TopUp(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	
	// Calls the new Hybrid service
	// Result: User gets money INSTANTLY. Treasury records it ~500ms later.
	err := services.ProcessTopUp(context.Background(), h.Client, userId, 100)
	
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "Topup Successful"})
}

// Purchase -> Uses Async Service
func (h *Handler) Purchase(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	var req struct { Item string `json:"item"` }
	c.BindJSON(&req)

	cost := 0
	if req.Item == "gold_coin" { cost = 10 }
	if req.Item == "treasure_box" { cost = 50 }
	
	if cost == 0 {
		c.JSON(400, gin.H{"error": "Invalid item"})
		return
	}

	// Calls the new Hybrid service
	// Result: User loses money INSTANTLY (Atomic). Treasury receives it ~500ms later.
	err := services.ProcessPurchase(context.Background(), h.Client, userId, req.Item, cost)
	
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "Purchase Successful"})
}