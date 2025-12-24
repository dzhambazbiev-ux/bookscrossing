package models

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	AISummary   string `json:"aisummary"`
	Status      string `json:"status" gorm:"enum:available,reserved"`
	UserID      uint   `json:"user_id"`

	User   *User   `json:"user" gorm:"foreignKey:UserID"`
	Genres []Genre `json:"genres" gorm:"many2many:book_genres"`
}
