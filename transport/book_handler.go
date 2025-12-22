package transport

import (
	"math"
	"net/http"
	"strings"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/services"
	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	service services.BookService
}

func NewBookHandler(service services.BookService) *BookHandler {
	return &BookHandler{service: service}
}

func (h *BookHandler) RegisterRoutes(r *gin.Engine) {
	books := r.Group("/books")
	{
		books.GET("", h.Search)
	}
}

func (h *BookHandler) Search(ctx *gin.Context) {
	var query dto.BookListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорретный JSON"})
		return
	}

	query.Author = strings.TrimSpace(query.Author)
	query.City = strings.TrimSpace(query.City)
	query.Status = strings.TrimSpace(query.Status)
	query.SortBy = strings.TrimSpace(query.SortBy)
	query.SortOrder = strings.TrimSpace(query.SortOrder)
	query.Title = strings.TrimSpace(query.Title)

	if query.Page <= 0 {
		query.Page = dto.DefaultPage
	}

	if query.Limit <= 0 {
		query.Limit = dto.DefaultLimit
	}

	if query.Limit > dto.MaxLimit {
		query.Limit = dto.MaxLimit
	}

	books, total, err := h.service.SearchBooks(query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBooks := make([]dto.BookResponse, 0, len(books))
	for _, b := range books {
		respBooks = append(respBooks, mapBookToResponse(b))
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.Limit)))
	if query.Limit <= 0 {
		totalPages = 0
	}

	ctx.JSON(http.StatusOK, dto.BookListResponse{
		Data:       respBooks,
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      int(total),
		TotalPages: totalPages,
	})
}

func mapBookToResponse(b models.Book) dto.BookResponse {
	owner := dto.UserPublicReponse{}
	if b.User != nil {
		owner = dto.UserPublicReponse{
			ID:   b.User.ID,
			Name: b.User.Name,
			City: b.User.City,
		}
	}
	genres := make([]dto.GenreResponse, 0, len(b.Genres))

	for _, g := range b.Genres {
		genres = append(genres, dto.GenreResponse{ID: g.ID, Name: g.Name})
	}

	return dto.BookResponse{
		ID:          b.ID,
		Title:       b.Title,
		Author:      b.Author,
		Description: b.Description,
		AISummary:   b.AISummary,
		Status:      b.Status,
		CreatedAt:   b.CreatedAt,
		Owner:       owner,
		Genres:      genres,
	}
}
