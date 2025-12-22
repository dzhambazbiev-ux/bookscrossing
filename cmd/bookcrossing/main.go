package main

import (
	"log/slog"
	"os"

	"github.com/dasler-fw/bookcrossing/internal/config"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/repository"
	"github.com/dasler-fw/bookcrossing/services"
	"github.com/dasler-fw/bookcrossing/transport"
	"github.com/gin-gonic/gin"
)

func main() {
	log := config.InitLogger()

	config.SetEnv(log)

	db := config.Connect(log)

	if err := db.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Genre{},
		&models.Exchange{},
		&models.Review{},
	); err != nil {
		log.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	reviewRepo := repository.NewReviewRepository(db, log)
	reviewService := services.NewReviewService(reviewRepo)

	log.Info("migrations completed")

	httpServer := gin.Default()

	reviewHandler := transport.NewReviewHandler(reviewService)
	reviewHandler.RegisterReviewRoutes(httpServer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := httpServer.Run(":" + port); err != nil {
		log.Error("не удалось запустить сервер", slog.Any("error", err))
	}
}
