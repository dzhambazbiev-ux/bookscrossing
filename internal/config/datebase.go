package config

import (
	"fmt"
	"log/slog"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(logger *slog.Logger) *gorm.DB {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbMode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%v user=%v dbname=%v port=%v sslmode=%v",
		dbHost, dbUser, dbName, dbPort, dbMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		logger.Error("failed to connect", "error", err)
		os.Exit(1)
	}

	logger.Info("connected to database")

	return db
}
