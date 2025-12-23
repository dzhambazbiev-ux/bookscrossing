package transport

import (
	"net/http"
	"strconv"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/services"
	"github.com/gin-gonic/gin"
)

type GenreHandler struct {
	service services.GenreService
}

func NewGenreHandler(service services.GenreService) *GenreHandler {
	return &GenreHandler{service: service}
}

func (h *GenreHandler) RegisterGenreRoutes(r *gin.RouterGroup) {
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
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"genre created": genre,
	})
}

func (h *GenreHandler) List(c *gin.Context) {
	genres, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get genres",
		})
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

	genre, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "genre not found",
		})
		return
	}

	c.JSON(http.StatusOK, genre)
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete genre",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "genre deleted",
	})
}
