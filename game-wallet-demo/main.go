package main

import (
	"fmt"
	"game-wallet-demo/config"
	"game-wallet-demo/internal/worker"
	"game-wallet-demo/internal/handlers"
	"game-wallet-demo/internal/middleware"
	"game-wallet-demo/prisma/db"
	"log"
	"time"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using system environment variables")
	}

	config.LoadConfig()

	allowedOrigins := []string{
        "http://localhost:5173", 
    }

    if config.FrontendURL != "" {
        allowedOrigins = append(allowedOrigins, config.FrontendURL)
    }

	// Database Setup
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}
	defer client.Prisma.Disconnect()

	// Initialize Handlers with DB
	worker.InitTreasuryBatcher(client)

	// Handlers ...
	h := handlers.NewHandler(client)

	// Router Setup
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true, // This is now safe because we are NOT using AllowAllOrigins
        MaxAge:           12 * time.Hour,
	}))

	// Define Routes
	r.POST("/signup", h.Signup)
	r.POST("/login", h.Login)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/balance", h.GetBalance)
		protected.POST("/top-up", h.TopUp)
		protected.POST("/purchase", h.Purchase)
	}

	fmt.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}