package main

import (
	"backend/config"
	"backend/controllers"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	//DB init
	config.ConnectDatabase()
	
	// create the router
	r := gin.Default()

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