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
			auth.PUT("/update", middlewares.AuthRequired(), controllers.UpdateUser)
			auth.PUT("/update-password", middlewares.AuthRequired(), controllers.UpdatePassword)
		}

		// Protected example
		api.GET("/me", middlewares.AuthRequired(), controllers.Me)
		api.GET("/health", controllers.HealthCheck)

	}

	return r
}
