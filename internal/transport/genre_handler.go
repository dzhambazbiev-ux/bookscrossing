package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/services"
	"github.com/gin-gonic/gin"
)

type GenreHandler struct {
	service services.GenreService
}

func NewGenreHandler(service services.GenreService) *GenreHandler {
	return &GenreHandler{service: service}
}

func (h *GenreHandler) RegisterGenreRoutes(r *gin.Engine) {
	r.POST("/genres", h.Create)
	r.GET("/genres", h.List)
	r.GET("/genres/:id", h.GetByID)
	r.DELETE("/genres/:id", h.Delete)
}

func (h *GenreHandler) Create(c *gin.Context) {
	var req dto.GenreCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}
	genre, err := h.service.Create(req)
	if err != nil {
		// map repository/service errors to HTTP codes
		switch {
		case errors.Is(err, dto.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		case errors.Is(err, dto.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create genre"})
			return
		}
	}

	// return created resource directly
	c.JSON(http.StatusCreated, genre)
}

func (h *GenreHandler) List(c *gin.Context) {
	genres, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get genres",
		})
		// logging can be done inside repo/service; avoid undefined handler logger here
		return
	}

	c.JSON(http.StatusOK, genres)
}

func (h *GenreHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid genre id",
		})
		return
	}

	g, err := h.service.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
			return
		} else if errors.Is(err, dto.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		} else if errors.Is(err, dto.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, g)
}

func (h *GenreHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid genre id",
		})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete genre"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "genre deleted",
	})
}
