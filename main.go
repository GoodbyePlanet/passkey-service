package main

import (
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
}
