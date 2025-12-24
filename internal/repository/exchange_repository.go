package repository

import (
	"errors"
	"log/slog"
	"time"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type ExchangeRepository interface {
	CreateExchange(req *models.Exchange) error
	CompleteExchange(req *models.Exchange) error
	CancelExchange(req *models.Exchange) error
	Update(req *models.Exchange) error
	GetByID(id uint) (*models.Exchange, error)
	GetAll() ([]models.Exchange, error)
}

type exchangeRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewExchangeRepository(db *gorm.DB, log *slog.Logger) ExchangeRepository {
	return &exchangeRepository{
		db:  db,
		log: log,
	}
}

func (r *exchangeRepository) CancelExchange(req *models.Exchange) error {
	if req == nil {
		r.log.Error("error in CancelExchange function exchange_repository.go")
		return dto.ErrExchangeCancelFailed
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Book{}).Where("id = ?", req.InitiatorBookID).Update("status", "available").Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Book{}).Where("id = ?", req.RecipientBookID).Update("status", "available").Error; err != nil {
			return err
		}

		req.Status = "cancelled"
		req.CompletedAt = nil
		if err := tx.Save(req).Error; err != nil {
			r.log.Error("error in CancelExchange function exchange_repository.go", "error", err)
			return err
		}
		return nil
	})
}
func (r *exchangeRepository) CompleteExchange(req *models.Exchange) error {
	if req == nil {
		r.log.Error("error in CompleteExchange function exchange_repository.go")
		return dto.ErrExchangeCompleteFailed
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if req.CompletedAt == nil {
			completedAt := time.Now()
			req.CompletedAt = &completedAt
		}

		if err := tx.Model(&models.Book{}).Where("id = ?", req.InitiatorBookID).Updates(map[string]interface{}{
			"status":  "available",
			"user_id": req.RecipientID,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Book{}).Where("id = ?", req.RecipientBookID).Updates(map[string]interface{}{
			"status":  "available",
			"user_id": req.InitiatorID,
		}).Error; err != nil {
			return err
		}

		if err := tx.Save(req).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *exchangeRepository) GetByID(id uint) (*models.Exchange, error) {
	if id == 0 {
		r.log.Error("error in GetByID function exchange_repository.go")
		return nil, dto.ErrExchangeGetFailed
	}

	var exchange models.Exchange
	if err := r.db.Where("id = ?", id).First(&exchange).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Error("error in GetByID function exchange_repository.go", "error", err)
			return nil, dto.ErrExchangeGetFailed
		}

		r.log.Error("error in GetByID function exchange_repository.go", "error", err)
		return nil, dto.ErrExchangeGetFailed
	}

	return &exchange, nil
}

func (r *exchangeRepository) CreateExchange(req *models.Exchange) error {
	if req == nil {
		r.log.Error("error in Create function exchange_repository.go")
		return dto.ErrExchangeCreateFailed
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(req).Error; err != nil {
			r.log.Error("error in CreateExchange function exchange_repository.go", "error", err)
			return err
		}
		if err := tx.Model(&models.Book{}).Where("id = ?", req.InitiatorBookID).Update("status", "reserved").Error; err != nil {
			r.log.Error("error in CreateExchange function exchange_repository.go", "error", err)
			return err
		}
		if err := tx.Model(&models.Book{}).Where("id = ?", req.RecipientBookID).Update("status", "reserved").Error; err != nil {
			r.log.Error("error in CreateExchange function exchange_repository.go", "error", err)
			return err
		}
		return nil
	})
}

func (r *exchangeRepository) Update(req *models.Exchange) error {
	if req == nil {
		r.log.Error("error in Update function book_repository.go")
		return dto.ErrExchangeUpdateFailed
	}

	return r.db.Save(req).Error
}

func (r *exchangeRepository) GetAll() ([]models.Exchange, error) {
	var exchanges []models.Exchange
	if err := r.db.Find(&exchanges).Error; err != nil {
		r.log.Error("error in GetAll function exchange_repository.go", "error", err)
		return nil, errors.New("error get exchanges from db")
	}
	return exchanges, nil
}
