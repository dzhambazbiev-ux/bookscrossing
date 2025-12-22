package services

import (
	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
)

type GenreService interface {
	CreateGenre(req dto.GenreCreateRequest) (*models.Genre, error)

	GetGenreByID(id uint) (*models.Genre, error)

	ListGenres() ([]models.Genre, error)

	DeleteGenre(id uint) error
}
