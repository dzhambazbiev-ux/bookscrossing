package repository

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type GenreRepository interface {
	Create(req *models.Genre) error
	GetByID(id uint) (*models.Genre, error)
	GetByName(name string) (*models.Genre, error)
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
		return dto.ErrInvalidInput
	}
	if existing, _ := r.GetByName(req.Name); existing != nil {
		return dto.ErrConflict
	}
	return r.db.Create(req).Error
}

func (r *genreRepository) GetByID(id uint) (*models.Genre, error) {
	var genre models.Genre

	if err := r.db.First(&genre, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		r.log.Error("error in GetByID genre", "id", id, "err", err)
		return nil, err
	}

	return &genre, nil
}

func (r *genreRepository) GetByName(name string) (*models.Genre, error) {
	var genre models.Genre

	if err := r.db.Where("name = ?", name).First(&genre).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		r.log.Error("error in GetByName genre", "name", name, "err", err)
		return nil, err
	}

	return &genre, nil
}

func (r *genreRepository) List() ([]models.Genre, error) {
	var genres []models.Genre

	if err := r.db.Find(&genres).Error; err != nil {
		r.log.Error("error in List genre", "err", err)
		return nil, err
	}

	return genres, nil
}

func (r *genreRepository) Delete(id uint) error {
	res := r.db.Delete(&models.Genre{}, id)
	if res.Error != nil {
		r.log.Error("error in Delete genre", "id", id, "err", res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		return dto.ErrNotFound
	}
	return nil
}
