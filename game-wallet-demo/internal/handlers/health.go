package handlers

import (
	"context"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"game-wallet-demo/prisma/db" 
)

// HealthCheck returns a Gin handler that has access to the DB client
func HealthCheck(client *db.PrismaClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic Response Structure
		status := gin.H{
			"status":    "UP",
			"service":   "game-wallet",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		// Database Check with Timeout
		// We use a 2-second timeout so the health check doesn't hang if DB is slow
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Run a raw SQL query "SELECT 1" to verify connection
		_, err := client.Prisma.QueryRaw("SELECT 1").Exec(ctx)

		if err != nil {
			status["database"] = "DOWN"
			status["error"] = err.Error()
			// Return 503 Service Unavailable if DB is down
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}

		status["database"] = "CONNECTED"
		c.JSON(http.StatusOK, status)
	}
}