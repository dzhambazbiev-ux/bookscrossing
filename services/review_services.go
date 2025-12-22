package services

import (
	"errors"
	"strings"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/repository"
)

type ReviewService interface {
	Create(authorID uint, req dto.CreateReviewRequest) error
	GetByUserID(userID uint) ([]models.Review, error)
	GetByBookID(bookID uint) ([]models.Review, error)
	Delete(reviewID uint, authorID uint) error
}

type reviewService struct {
	repo repository.ReviewRepository
}

func NewReviewService(repo repository.ReviewRepository) ReviewService {
	return &reviewService{repo: repo}
}

func (s *reviewService) Create(authorID uint, req dto.CreateReviewRequest) error {
	trimmedText := strings.TrimSpace(req.Text)

	// if trimmedText == "" {
	// 	return errors.New("review text is required")
	// }

	length := len([]rune(trimmedText))
	if length < 10 || length > 150 {
		return errors.New("review text must be between 10 and 150 characters")
	}

	if strings.TrimSpace(req.Text) == "" {
		return errors.New("review text is request")
	}

	if req.Rating < 1 || req.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	if req.TargetUserID == authorID {
		return errors.New("cannot leave review to yourself")
	}

	review := models.Review{
		AuthorID:     authorID,
		TargetUserID: req.TargetUserID,
		TargetBookID: req.TargetBookID,
		Text:         req.Text,
		Rating:       req.Rating,
	}
	return s.repo.Create(&review)
}

func (s *reviewService) GetByUserID(userID uint) ([]models.Review, error) {
	return s.repo.GetByUserID(userID)
}

func (s *reviewService) GetByBookID(bookID uint) ([]models.Review, error) {
	return s.repo.GetByBookID(bookID)
}

func (s *reviewService) Delete(reviewID uint, authorID uint) error {
	review, err := s.repo.GetByID(reviewID)
	if err != nil {
		return err
	}

	if review.AuthorID != authorID {
		return errors.New("you are not allowed to delete this review")
	}
	return s.repo.Delete(reviewID)
}
