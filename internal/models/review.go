package models

import (
	"time"

	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	AuthorID     uint
	TargetUserID uint `json:"targer_user_id"`
	TargetBookID uint `json:"targer_book_id"`
	Text         string
	Rating       int
	CreatedAt    time.Time

	Author     *User `gorm:"foreignKey:AuthorID"`
	TargetUser *User `gorm:"foreignKey:TargetUserID"`
	TargetBook *Book `gorm:"foreignKey:TargetBookID"`
}
