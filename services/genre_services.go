package services

import (
	"errors"
	"strings"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/repository"
)

type GenreService interface {
	Create(req dto.GenreCreateRequest) (*models.Genre, error)
	GetByID(id uint) (*models.Genre, error)
	List() ([]models.Genre, error)
	Delete(id uint) error
}

type genreService struct {
	repo repository.GenreRepository
}

func NewGenreService(repo repository.GenreRepository) GenreService {
	return &genreService{repo: repo}
}

func (s *genreService) Create(req dto.GenreCreateRequest) (*models.Genre, error) {
	name := strings.TrimSpace(req.Name)

	if name == "" {
		return nil, errors.New("genre name is required")
	}

	genre := &models.Genre{
		Name: name,
	}

	if err := s.repo.Create(genre); err != nil {
		return nil, err
	}

	return genre, nil
}

func (s *genreService) GetByID(id uint) (*models.Genre, error) {
	return s.repo.GetByID(id)
}

func (s *genreService) List() ([]models.Genre, error) {
	return s.repo.List()
}

func (s *genreService) Delete(id uint) error {

	return s.repo.Delete(id)
}
