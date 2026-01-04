package config

import (
	"log/slog"

	"github.com/joho/godotenv"
)

func SetEnv(logger *slog.Logger) {
	if err := godotenv.Load(".env"); err != nil {
		logger.Warn(".env not found, using OS environment variables", "error", err)
		return
	}
	logger.Info("environment variables loaded successfully")
}
