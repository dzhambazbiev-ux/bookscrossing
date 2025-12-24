package transport

import (
	"net/http"
	"strconv"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/services"
	"github.com/gin-gonic/gin"
)

type ExchangeHandler struct {
	exchangeService services.ExchangeService
}

func NewExchangeHandler(exchangeService services.ExchangeService) *ExchangeHandler {
	return &ExchangeHandler{exchangeService: exchangeService}
}

func (h *ExchangeHandler) RegisterExchangeRoutes(router *gin.Engine) {
	router.POST("/exchanges", h.CreateExchange)
	router.PUT("/exchanges/:id/accept", h.AcceptExchange)
	router.PUT("/exchanges/:id/complete", h.CompleteExchange)
	router.PUT("/exchanges/:id/cancel", h.CancelExchange)
}

func (h *ExchangeHandler) CancelExchange(c *gin.Context) {
	exchangeID := c.Param("id")
	exchangeIDInt, err := strconv.Atoi(exchangeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.exchangeService.CancelExchange(uint(exchangeIDInt)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Exchange cancelled successfully"})
}

func (h *ExchangeHandler) CompleteExchange(c *gin.Context) {
	exchangeID := c.Param("id")
	exchangeIDInt, err := strconv.Atoi(exchangeID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.exchangeService.CompleteExchange(uint(exchangeIDInt)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange completed successfully"})
}

func (h *ExchangeHandler) CreateExchange(c *gin.Context) {
	var req dto.CreateExchangeRequest
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

func (h *ExchangeHandler) AcceptExchange(c *gin.Context) {
	exchangeID := c.Param("id")
	exchangeIDInt, err := strconv.Atoi(exchangeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.exchangeService.AcceptExchange(uint(exchangeIDInt)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Exchange accepted successfully"})
}
