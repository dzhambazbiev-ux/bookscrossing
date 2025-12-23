package repository

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type ExchangeRepository interface {
	CreateExchange(req *models.Exchange) error
	Update(req *models.Exchange) error
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

func (r *exchangeRepository) CreateExchange(req *models.Exchange) error {
	if req == nil {
		r.log.Error("error in Create function exchange_repository.go")
		return errors.New("error create category in db")
	}

	return r.db.Create(req).Error
}

func (r *exchangeRepository) Update(req *models.Exchange) error {
	if req == nil {
		r.log.Error("error in Update function book_repository.go")
		return errors.New("error update in db")
	}

	return r.db.Save(req).Error
}
