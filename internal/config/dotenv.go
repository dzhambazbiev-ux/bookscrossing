package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func SetEnv(logger *slog.Logger) {
	err := godotenv.Load(".env")

	if err != nil {
		logger.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}

	logger.Info("environment variables loaded successfully")
}
