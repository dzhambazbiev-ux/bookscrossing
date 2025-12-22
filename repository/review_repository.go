package repository

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(req *models.Review) error
	GetByID(id uint) (*models.Review, error)
	GetByUserID(userID uint) ([]models.Review, error)
	GetByBookID(bookID uint) ([]models.Review, error)
	Delete(id uint) error
}

type reviewRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewReviewRepository(db *gorm.DB, log *slog.Logger) ReviewRepository {
	return &reviewRepository{db: db, log: log}
}

func (r *reviewRepository) Create(req *models.Review) error {
	if req == nil {
		r.log.Error("error in create review")
		return errors.New("error create review")
	}
	return r.db.Create(req).Error
}

func (r *reviewRepository) GetByID(id uint) (*models.Review, error) {
	var reviews models.Review

	if err := r.db.First(&reviews, id).Error; err != nil {
		r.log.Error("error in GetByID review")
		return nil, errors.New("review not found")
	}
	return &reviews, nil
}

func (r *reviewRepository) GetByUserID(userID uint) ([]models.Review, error) {
	var reviews []models.Review

	if err := r.db.
		Preload("Author").
		Preload("TargetBook").
		Where("target_user_id = ?", userID).
		Find(&reviews).Error; err != nil {

		r.log.Error("error in GetByUserID review")
		return nil, err
	}
	return reviews, nil
}

func (r *reviewRepository) GetByBookID(bookID uint) ([]models.Review, error) {
	var reviews []models.Review

	if err := r.db.
		Preload("Author").
		Where("target_book_id = ?", bookID).
		Find(&reviews).Error; err != nil {

		r.log.Error("error in GetBYBookID review")
		return nil, err
	}
	return reviews, nil
}

func (r *reviewRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Review{}, id).Error; err != nil {
		r.log.Error("error in Delete review")
		return errors.New("error delete review")
	}
	return nil
}
