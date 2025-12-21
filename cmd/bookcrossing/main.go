package main

import (
	"log/slog"
	"os"

	"github.com/dasler-fw/bookcrossing/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	log := config.InitLogger()

	config.SetEnv(log)

	db := config.Connect(log)

	if err := db.AutoMigrate(); err != nil {
		log.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	log.Info("migrations completed")

	httpServer := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := httpServer.Run(":" + port); err != nil {
		log.Error("не удалось запустить сервер", slog.Any("error", err))
	}
}
