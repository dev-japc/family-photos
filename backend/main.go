package main

import (
	"time"
	"backend/config"
	"backend/controllers"
	"backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {
	//DB init
	config.ConnectDatabase()
	
	// create the router
	r := gin.Default()

	// --- CONFIGURACIÓN DE CORS ---
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4321"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Enable static file serving
	// This allows accessing the 'uploads' folder via HTTP
	r.Static("/uploads", "./uploads")

	// defining base routes for testing
	r.GET("/", func(c *gin.Context){
		c.JSON(200, gin.H{
			"message": "API working fine with Gin Gonic",
		})
	})

	// Admin routes
	adminRoutes := r.Group("/api/admin")
	adminRoutes.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adminRoutes.DELETE("/users/:id", controllers.AdminDeleteUser)
		adminRoutes.GET("/whitelist", controllers.GetWhitelist)
		adminRoutes.POST("/whitelist", controllers.AddToWhitelist)
		adminRoutes.DELETE("/whitelist/:national_id_number", controllers.DeleteFromWhitelist)
	}

	//grouping auth routes
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/register", controllers.Register)
		authRoutes.POST("/login", controllers.Login)
		authRoutes.POST("/logout", controllers.Logout)
	}

	//grouping photo routes
	photoRoutes := r.Group("/api/photos")
	photoRoutes.Use(middleware.AuthMiddleware())
	{
		photoRoutes.GET("", controllers.GetPhotos)

		photoRoutes.POST("/upload", middleware.AdminMiddleware(), controllers.UploadPhoto)
		photoRoutes.PUT("/:id", middleware.AdminMiddleware(), controllers.UpdatePhoto)
		photoRoutes.DELETE("/:id", middleware.AdminMiddleware(), controllers.DeletePhoto)
	}

	// user routes
	userRoutes := r.Group("/api/user")
	userRoutes.Use(middleware.AuthMiddleware())
	{
		userRoutes.PUT("", controllers.UpdateUser)    
		userRoutes.DELETE("", controllers.DeleteUser) 
	}

	//run the server
	r.Run(":8080")
}