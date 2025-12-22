package repository

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(req *models.Review) error
	GetByTargetUserID(id uint) ([]models.Review, error)
	GetByTargetBookID(id uint) ([]models.Review, error)
}

type reviewRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewReviewRepository(db *gorm.DB, log *slog.Logger) ReviewRepository {
	return &reviewRepository{
		db:  db,
		log: log,
	}
}

func (r *reviewRepository) Create(req *models.Review) error {
	if req == nil {
		r.log.Error("error in Create function review_repository.go")
		return errors.New("error create category in db")
	}

	return r.db.Create(req).Error
}

func (r *reviewRepository) GetByTargetUserID(id uint) ([]models.Review, error) {
	var list []models.Review
	if err := r.db.
		Where("target_user_id = ?", id).
		Preload("Author").
		Preload("TargetBook").
		Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (r *reviewRepository) GetByTargetBookID(id uint) ([]models.Review, error) {
	var list []models.Review
	if err := r.db.
		Where("target_book_id = ?", id).
		Preload("Author").
		Preload("TargetUser").
		Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}
