package models

import (
	"time"

	"gorm.io/gorm"
)

type Exchange struct {
	gorm.Model
	InitiatorID     uint       `json:"initiator_id"`
	RecipientID     uint       `json:"recipient_id"`
	InitiatorBookID uint       `json:"initiator_book_id"`
	RecipientBookID uint       `json:"recipient_book_id"`
	Status          string     `json:"status" gorm:"enum:pending,accepted,completed,cancelled"`
	CompletedAt     *time.Time `json:"completed_at"`

	Initiator *User `json:"initiator" gorm:"foreignKey:InitiatorID"`
	Recipient *User `json:"recipient" gorm:"foreignKey:RecipientID"`

	InitiatorBook *Book `json:"initiator_book" gorm:"foreignKey:InitiatorBookID"`
	RecipientBook *Book `json:"recipient_book" gorm:"foreignKey:RecipientBookID"`
}
