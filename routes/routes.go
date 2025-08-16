package routes

import (
	"alkaukaba-backend/controllers"
	"alkaukaba-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/")
	{
		// Auth
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.GET("/google", controllers.GoogleLogin)
			auth.GET("/google/callback", controllers.GoogleCallback)
		}

		// Protected example
		api.GET("/me", middlewares.AuthRequired(), controllers.Me)
		api.GET("/health", controllers.HealthCheck)

	}

	return r
}