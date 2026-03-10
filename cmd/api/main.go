package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/susapa/Inventory_API/docs" // Swagger docs
	"github.com/susapa/Inventory_API/internal/auth"
	"github.com/susapa/Inventory_API/internal/config"
	"github.com/susapa/Inventory_API/internal/database"
	"github.com/susapa/Inventory_API/internal/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Inventory API
// @version         1.0
// @description     This is an inventory management API server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// 0. Load Configuration
	config.LoadConfig()

	// 1. Init DB
	database.InitDB()

	// 2. Setup Gin
	r := gin.Default()

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 3. Public Routes
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", auth.Register)
		authRoutes.POST("/login", auth.Login)
		authRoutes.POST("/logout", auth.Logout)
	}

	// 4. Protected Routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", auth.GetProfile)
	}

	// 5. Run Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s in %s mode", port, config.GetEnv())
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
