package repository

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type GenreRepository interface {
	Create(req *models.Genre) error
	GetByID(id uint) (*models.Genre, error)
	List() ([]models.Genre, error)
	Delete(id uint) error
}

type genreRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewGenreRepository(db *gorm.DB, log *slog.Logger) GenreRepository {
	return &genreRepository{
		db:  db,
		log: log,
	}
}

func (r *genreRepository) Create(req *models.Genre) error {
	if req == nil {
		r.log.Error("genre is nil in Create")
		return errors.New("genre is nil")
	}
	return r.db.Create(req).Error
}

func (r *genreRepository) GetByID(id uint) (*models.Genre, error) {
	var genre models.Genre

	if err := r.db.First(&genre, id).Error; err != nil {
		r.log.Error("error in GetByID genre", "id", id)
		return nil, err
	}

	return &genre, nil
}

func (r *genreRepository) List() ([]models.Genre, error) {
	var genres []models.Genre

	if err := r.db.Find(&genres).Error; err != nil {
		r.log.Error("error in List genre")
		return nil, err
	}

	return genres, nil
}

func (r *genreRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Genre{}, id).Error; err != nil {
		r.log.Error("error in Delete genre", "id", id)
		return err
	}

	return nil
}
