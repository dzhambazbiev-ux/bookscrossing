package models

import (
	"time"

	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	AuthorID     uint
	TargetUserID uint
	TargetBookID uint
	Text         string
	Rating       int
	CreatedAt    time.Time

	Author     *User `gorm:"foreignKey:AuthorID"`
	TargetUser *User `gorm:"foreignKey:TargetUserID"`
	TargetBook *Book `gorm:"foreignKey:TargetBookID"`
}
