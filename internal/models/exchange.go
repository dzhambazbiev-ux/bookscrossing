package models

import (
	"time"

	"gorm.io/gorm"
)

type Exchange struct {
	gorm.Model
	InitiatorID     uint
	RecipientID     uint
	InitiatorBookID uint
	RecipientBookID uint
	Status          string
	CompletedAt     *time.Time

	Initiator *User `gorm:"foreignKey:InitiatorID"`
	Recipient *User `gorm:"foreignKey:RecipientID"`

	InitiatorBook *Book `gorm:"foreignKey:InitiatorBookID"`
	RecipientBook *Book `gorm:"foreignKey:RecipientBookID"`
}
