package main

import (
	graphql2 "github.com/99designs/gqlgen/graphql"
	"log"
	"net/http"
	"payment-service-iae/config"
	"payment-service-iae/database"
	"payment-service-iae/graphql"
	"payment-service-iae/middleware"
	"payment-service-iae/services"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize services
	authClient := services.NewAuthClient(cfg)
	userClient := services.NewUserClient(cfg.UserServiceURL)
	midtransService := services.NewMidtransService(cfg)
	paymentService := services.NewPaymentService(db, midtransService, userClient)

	// Initialize GraphQL resolver
	resolver := graphql.NewResolver(paymentService)

	// Create GraphQL handler
	srv := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "payment",
			"version": "1.0.0",
		})
	})

	// GraphQL playground (development only)
	r.GET("/", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))

	// GraphQL endpoint with authentication
	r.POST("/graphql", middleware.AuthMiddleware(authClient), gin.WrapH(srv))

	// REST API endpoints
	api := r.Group("/api/v1")
	{
		// Public endpoints
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})

		// Protected endpoints
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(authClient))
		{
			// Admin endpoints with specific permissions
			admin := protected.Group("/admin")
			admin.Use(middleware.RequirePermission(authClient, "payment", "read_all"))
			{
				admin.GET("/payments", func(c *gin.Context) {
					// Get all payments with pagination
					c.JSON(http.StatusOK, gin.H{"message": "Admin payments endpoint"})
				})

				admin.GET("/stats", func(c *gin.Context) {
					// Get payment statistics
					c.JSON(http.StatusOK, gin.H{"message": "Payment stats endpoint"})
				})
			}
		}
	}

	// Webhook endpoint for Midtrans notifications (no auth required)
	r.POST("/webhook/midtrans", func(c *gin.Context) {
		var notification map[string]interface{}
		if err := c.ShouldBindJSON(&notification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		log.Printf("Received Midtrans notification: %+v", notification)

		if err := paymentService.HandleNotification(notification); err != nil {
			log.Printf("Error handling notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process notification"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	log.Printf("Payment service starting on port %s", cfg.Port)
	log.Printf("GraphQL playground available at http://localhost:%s/", cfg.Port)
	log.Printf("GraphQL endpoint at http://localhost:%s/graphql", cfg.Port)
	log.Printf("Webhook endpoint at http://localhost:%s/webhook/midtrans", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Placeholder for gqlgen generated code
func NewExecutableSchema(cfg Config) graphql2.ExecutableSchema {
	return nil
}

type Config struct {
	Resolvers interface{}
}
