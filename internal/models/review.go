package models

import (
	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	AuthorID     uint   `json:"author_id"`
	TargetUserID uint   `json:"target_user_id"`
	TargetBookID uint   `json:"target_book_id"`
	Text         string `json:"text"`
	Rating       int    `json:"rating"`

	Author     *User `json:"author" gorm:"foreignKey:AuthorID"`
	TargetUser *User `json:"target_user" gorm:"foreignKey:TargetUserID"`
	TargetBook *Book `json:"target_book" gorm:"foreignKey:TargetBookID"`
}
