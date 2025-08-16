package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser            string
	DBPass            string
	DBHost            string
	DBName            string
	Port              string
	JWTSecret         string
	GoogleClientID    string
	GoogleClientSecret string
	GoogleRedirectURL string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	cfg := Config{
		DBUser:             os.Getenv("DB_USER"),
		DBPass:             os.Getenv("DB_PASS"),
		DBHost:             os.Getenv("DB_HOST"),
		DBName:             os.Getenv("DB_NAME"),
		Port:               os.Getenv("PORT"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
		log.Println("PORT not set, defaulting to 8080")
	}

	return cfg
}