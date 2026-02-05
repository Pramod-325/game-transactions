package handlers

import (
	"game-wallet-demo/prisma/db"
)

// Handler holds dependencies for your HTTP controllers
type Handler struct {
	Client *db.PrismaClient
}

// NewHandler creates a new Handler instance
func NewHandler(client *db.PrismaClient) *Handler {
	return &Handler{Client: client}
}