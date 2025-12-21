package models

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	Title       string
	Author      string
	Description string
	AISummary   string
	Status      string
	UserID      uint

	User   *User   `gorm:"foreignKey:UserID"`
	Genres []Genre `gorm:"many2many:book_genres"`
}
