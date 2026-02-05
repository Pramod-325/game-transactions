package handlers

import (
	"context"
	"fmt"
	"game-wallet-demo/config"
	"game-wallet-demo/internal/services"
	"game-wallet-demo/internal/utils"
	"game-wallet-demo/prisma/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (h *Handler) Signup(c *gin.Context) {
	var req struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		ReferralCode string `json:"referralCode"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid Data"})
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	ownRefCode := fmt.Sprintf("REF-%s", utils.GenerateRandomString(6))

	ctx := context.Background()
	var optionalParams []db.UserSetParam
	if req.ReferralCode != "" {
		optionalParams = append(optionalParams, db.User.ReferredBy.Set(req.ReferralCode))
	}

	// 1. Create User (Standard DB Call)
	newUser, err := h.Client.User.CreateOne(
		db.User.Username.Set(req.Username),
		db.User.Password.Set(string(hashedPwd)),
		db.User.ReferralCode.Set(ownRefCode),
		optionalParams...,
	).Exec(ctx)

	if err != nil {
		c.JSON(400, gin.H{"error": "Signup failed (User likely exists)"})
		return
	}

	// 2. Create Assets
	h.Client.Inventory.CreateOne(db.Inventory.User.Link(db.User.ID.Equals(newUser.ID))).Exec(ctx)
	h.Client.Account.CreateOne(db.Account.User.Link(db.User.ID.Equals(newUser.ID)), db.Account.Name.Set("Main Wallet")).Exec(ctx)

	// 3. Referral Bonus (NOW ASYNC / NON-BLOCKING)
	// Even if 10,000 users signup, this will not lock the Treasury row.
	if req.ReferralCode != "" {
		if referrer, err := h.Client.User.FindUnique(db.User.ReferralCode.Equals(req.ReferralCode)).Exec(ctx); err == nil {
			
			// Bonus for New User
			services.ProcessBonus(ctx, h.Client, newUser.ID, 50, "Welcome Bonus")
			
			// Bonus for Referrer
			services.ProcessBonus(ctx, h.Client, referrer.ID, 50, "Referral Reward")
		}
	}

	c.JSON(200, gin.H{"userId": newUser.ID, "status": "Created"})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil { return }

	user, err := h.Client.User.FindUnique(db.User.Username.Equals(req.Username)).Exec(context.Background())
	if err != nil {
		c.JSON(401, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Wrong password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString(config.JwtSecret)

	c.JSON(200, gin.H{"token": tokenString})
}