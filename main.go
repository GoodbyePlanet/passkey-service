package main

import (
	"log"
	"log/slog"
	"os"
	"passkey-service/config"
	"passkey-service/models"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := godotenv.Load(".env.development"); err != nil {
		logger.Warn("Warning: could not load .env.development (%v)", err)
	}

	config.ConnectDatabase()
	err := config.DB.AutoMigrate(&models.User{}, &models.Credential{})
	if err != nil {
		return
	}
	logger.Info("Database initialized!")

	InitWebAuthn()
	r := SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	err = r.Run(":" + port)
	if err != nil {
		return
	}
}
