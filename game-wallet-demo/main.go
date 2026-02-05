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

	// 1. Database Setup
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}
	defer client.Prisma.Disconnect()

	// 2. Initialize Handlers with DB
	worker.InitTreasuryBatcher(client)

	// 4. Handlers ...
	h := handlers.NewHandler(client)

	// 3. Router Setup
	r := gin.Default()

	// 4. CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	r.GET("/health", func(c *gin.Context) {
		// Basic Server Check
		status := gin.H{
			"status": "UP",
			"service": "game-wallet",
		}

		ctx := context.Background()
		_, err := client.Prisma.QueryRaw("SELECT 1").Exec(ctx)
		
		if err != nil {
			status["database"] = "DOWN"
			status["error"] = err.Error()
			c.JSON(503, status)
			return
		}

		status["database"] = "CONNECTED"
		c.JSON(200, status)
	})

	// 5. Define Routes
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