package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/dasler-fw/bookcrossing/internal/config"
	"github.com/dasler-fw/bookcrossing/internal/middleware"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/repository"
	"github.com/dasler-fw/bookcrossing/internal/services"
	"github.com/dasler-fw/bookcrossing/internal/transport"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	log := config.InitLogger()

	config.SetEnv(log)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("redis ping failed", "err", err)
	} else {
		log.Info("redis ping ok")
	}

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

	log.Info("migrations completed")

	reviewRepo := repository.NewReviewRepository(db, log)
	exchangeRepo := repository.NewExchangeRepository(db, log)
	bookRepo := repository.NewBookRepository(db, log)
	userRepo := repository.NewUserRepository(db, log)
	genreRepo := repository.NewGenreRepository(db, log)

	exchangeService := services.NewExchangeService(exchangeRepo, bookRepo, log)
	reviewService := services.NewReviewService(reviewRepo)
	bookService := services.NewServiceBook(bookRepo, log, rdb)
	userService := services.NewServiceUser(db, userRepo, bookRepo, log)
	genreService := services.NewGenreService(genreRepo)

	httpServer := gin.New()
	httpServer.Use(gin.Recovery())
	httpServer.Use(middleware.RequestLogger(log))

	transport.RegisterRoutes(
		httpServer,
		log,
		bookService,
		exchangeService,
		genreService,
		reviewService,
		userService,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := httpServer.Run(":" + port); err != nil {
		log.Error("не удалось запустить сервер", slog.Any("error", err))
	}
}
