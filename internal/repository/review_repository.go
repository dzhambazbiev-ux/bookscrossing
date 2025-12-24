package repository

import (
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(req *models.Review) error
	GetByID(id uint) (*models.Review, error)
	Delete(id uint) error
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
		r.log.Error("error in create review")
		return dto.ErrReviewCreateFail
	}
	return r.db.Create(req).Error
}

func (r *reviewRepository) GetByID(id uint) (*models.Review, error) {
	var reviews models.Review

	if err := r.db.First(&reviews, id).Error; err != nil {
		r.log.Error("error in GetByID review")
		return nil, dto.ErrReviewNotFound
	}
	return &reviews, nil
}

func (r *reviewRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Review{}, id).Error; err != nil {
		r.log.Error("error in Delete review")
		return dto.ErrReviewDeleteFail
	}

	return nil
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
