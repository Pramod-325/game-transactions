package config

import (
	"os"
)

// Default fallback (only for local dev safety)
var JwtSecret = []byte("unsafe-default-secret")
var FrontendURL string

const TreasuryID = "SYSTEM_TREASURY_ID"

func LoadConfig() {
	secret := os.Getenv("JWT_SECRET")
	FrontendURL = os.Getenv("FRONTEND_URL")
	if secret != "" {
		JwtSecret = []byte(secret)
	} else {
		panic("JWT_SECRET environment variable is not set")
	}
}