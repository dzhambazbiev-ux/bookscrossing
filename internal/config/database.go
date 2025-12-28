package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Connect(logger *slog.Logger) *gorm.DB {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbMode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v",
		dbHost, dbUser, dbPass, dbName, dbPort, dbMode)

	level := gormlogger.Warn
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		level = gormlogger.Info
	}

	gormLogger := gormlogger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})

	if err != nil {
		logger.Error("failed to connect", "error", err)
		os.Exit(1)
	}

	logger.Info("connected to database")

	return db
}
