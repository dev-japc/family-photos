package main

import (
	"backend/config"
	"backend/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	//DB init
	config.ConnectDatabase()
	
	// create the router
	r := gin.Default()

	// defining base routes for testing
	r.GET("/", func(c *gin.Context){
		c.JSON(200, gin.H{
			"message": "API working fine with Gin Gonic",
		})
	})

	//grouping auth routes
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/register", controllers.Register)
		authRoutes.POST("/login", controllers.Login)
	}

	//run the server
	r.Run(":8080")
}