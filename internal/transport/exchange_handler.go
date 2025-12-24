package transport

import (
	"net/http"
	"strconv"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
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
	exchange, err := h.exchangeService.CreateExchange(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapExchangeToResponse(*exchange))
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

func (h *ExchangeHandler) GetByID(c *gin.Context) {
	exchangeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	exchange, err := h.exchangeService.GetByID(uint(exchangeID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exchange)
}

func (h *ExchangeHandler) GetAll(c *gin.Context) {
	exchanges, err := h.exchangeService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.ExchangeResponse, 0, len(exchanges))
	for _, exchange := range exchanges {
		response = append(response, mapExchangeToResponse(exchange))
	}

	c.JSON(http.StatusOK, response)
}

func mapExchangeToResponse(e models.Exchange) dto.ExchangeResponse {
	return dto.ExchangeResponse{
		ID:              e.ID,
		InitiatorID:     e.InitiatorID,
		RecipientID:     e.RecipientID,
		InitiatorBookID: e.InitiatorBookID,
		RecipientBookID: e.RecipientBookID,
		Status:          e.Status,
		CompletedAt:     e.CompletedAt,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}
