package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	City         string    `json:"city"`
	Address      string    `json:"address"`
	RegisteredAt time.Time `json:"registered_at"`
}
