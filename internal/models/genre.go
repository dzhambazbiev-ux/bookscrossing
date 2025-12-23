package models

import "gorm.io/gorm"

type Genre struct {
	gorm.Model
	Name  string `json:"name"`
	Books []Book `json:"books" gorm:"many2many:book_genres"`
}
