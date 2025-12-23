package transport

import (
	"net/http"

	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/services"
	"github.com/gin-gonic/gin"
)

type ExchangeHandler struct {
	exchangeService services.ExchangeService
}

func NewExchangeHandler(exchangeService services.ExchangeService) *ExchangeHandler {
	return &ExchangeHandler{exchangeService: exchangeService}
}

func (h *ExchangeHandler) CreateExchange(c *gin.Context) {
	var req models.Exchange
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.exchangeService.CreateExchange(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "Exchange created successfully"})
}
