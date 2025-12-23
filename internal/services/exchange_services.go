package services

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/repository"
)

type ExchangeService interface {
	CreateExchange(req *models.Exchange) error
}

type exchangeService struct {
	exchangeRepo repository.ExchangeRepository
	log          *slog.Logger
}

func NewExchangeService(exchangeRepo repository.ExchangeRepository, log *slog.Logger) ExchangeService {
	return &exchangeService{exchangeRepo: exchangeRepo, log: log}
}

func (s *exchangeService) CreateExchange(req *models.Exchange) error {
	if req == nil {
		s.log.Error("error in CreateExchange function exchange_services.go")
		return errors.New("error create exchange in db")
	}

	return s.exchangeRepo.CreateExchange(req)
}
