package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string
	Email        string
	PasswordHash string
	City         string
	Address      string
	RegisteredAt time.Time
}
