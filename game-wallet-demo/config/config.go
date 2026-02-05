package config

import (
	"os"
)

// Default fallback (only for local dev safety)
var JwtSecret = []byte("unsafe-default-secret")

const TreasuryID = "SYSTEM_TREASURY_ID"

func LoadConfig() {
	secret := os.Getenv("JWT_SECRET")
	if secret != "" {
		JwtSecret = []byte(secret)
	} else {
		panic("JWT_SECRET environment variable is not set")
	}
}