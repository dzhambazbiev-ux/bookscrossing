package repository

import "github.com/dasler-fw/bookcrossing/internal/models"

type GenreRepository interface {
	Create(genre *models.Genre) error

	GetByID(id uint) (*models.Genre, error)

	List() ([]models.Genre, error)

	Delete(id uint) error
}
