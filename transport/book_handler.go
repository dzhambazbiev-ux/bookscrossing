package transport

import (
	"math"
	"net/http"
	"strconv"
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
		books.POST("", h.CreateBook)
		books.GET("/:id", h.GetBookByID)
		books.GET("", h.GetBookList)
		books.PATCH("/:id", h.UpdateBook)
		books.DELETE("/:id", h.DeleteBook)
		books.GET("", h.Search)
	}
}

func (h *BookHandler) CreateBook(ctx *gin.Context) {
	var input dto.CreateBookRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")

	book, err := h.service.CreateBook(userID, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(201, book)
}


func (h *BookHandler) GetBookByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := h.service.GetByID(uint(id))
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	ctx.IndentedJSON(http.StatusOK, book)
}

func (h *BookHandler) GetBookList(ctx *gin.Context) {
	list, err := h.service.GetList()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get books"})
		return
	}

	ctx.IndentedJSON(http.StatusOK, list)
}

func (h *BookHandler) UpdateBook(ctx *gin.Context) {
	bookID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "invalid book id"})
		return
	}

	userID := ctx.GetUint("user_id") // 🔥 из JWT

	var req dto.UpdateBookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	book, err := h.service.Update(uint(bookID), userID, req)
	if err != nil {
		ctx.JSON(403, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, mapBookToResponse(*book))
}


func (h *BookHandler) DeleteBook(ctx *gin.Context) {
	bookID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	userID := ctx.GetUint("user_id")

	book, err := h.service.GetByID(uint(bookID))
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	if book.UserID != userID {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"error": "нельзя удалить чужую книгу"})
		return
	}

	if err := h.service.Delete(uint(bookID), userID); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"deleted": true})
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
