package main

import (
	"alkaukaba-backend/config"
	"alkaukaba-backend/controllers"
	"alkaukaba-backend/database"
	"alkaukaba-backend/models"
	"alkaukaba-backend/routes"
	"alkaukaba-backend/utils"
)

func main() {
	cfg := config.LoadConfig()

	// DB
	database.ConnectDB(cfg)
	database.DB.AutoMigrate(&models.User{})

	// JWT
	utils.InitJWT(cfg.JWTSecret)

	// Init OAuth providers used by controllers
	controllers.InitAuth(cfg)

	r := routes.SetupRouter()
	r.Run(":" + cfg.Port)
}