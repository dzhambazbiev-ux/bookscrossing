package transport

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/middleware"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/services"
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
		books.POST("", middleware.JWTAuth(), h.CreateBook)
		books.GET("", h.Search)
		books.GET("/available", h.GetAvailable)
		books.GET("/list", h.GetBookList)
		books.GET("/:id", h.GetBookByID)
		books.PATCH("/:id", middleware.JWTAuth(), h.UpdateBook)
		books.DELETE("/:id", middleware.JWTAuth(), h.DeleteBook)
	}
	r.GET("/users/:id/books", h.GetByUserID)
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

	ctx.JSON(http.StatusCreated, mapBookToResponse(*book))
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	userID := ctx.GetUint("user_id") // 🔥 из JWT

	var req dto.UpdateBookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := h.service.Update(uint(bookID), userID, req)
	if err != nil {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, mapBookToResponse(*book))
}

func (h *BookHandler) DeleteBook(ctx *gin.Context) {
	bookID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	userID := ctx.GetUint("user_id")

	if err := h.service.Delete(uint(bookID), userID); err != nil {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *BookHandler) Search(ctx *gin.Context) {
	var query dto.BookListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорркетные параметры"})
		return
	}

	query.Author = strings.TrimSpace(query.Author)
	query.City = strings.TrimSpace(query.City)
	query.Status = strings.TrimSpace(query.Status)
	query.SortBy = strings.TrimSpace(query.SortBy)
	query.SortOrder = strings.TrimSpace(query.SortOrder)
	query.Title = strings.TrimSpace(query.Title)

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
	owner := dto.UserPublicResponse{}
	if b.User != nil {
		owner = dto.UserPublicResponse{
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

func (h *BookHandler) GetByUserID(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	status := ctx.Query("status")

	books, err := h.service.GetBooksByUserID(uint(userID), status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBook := make([]dto.BookResponse, 0, len(books))
	for _, b := range books {
		respBook = append(respBook, mapBookToResponse(b))
	}

	ctx.JSON(http.StatusOK, respBook)
}

func (h *BookHandler) GetAvailable(ctx *gin.Context) {
	city := ctx.Query("city")

	books, err := h.service.GetAvailableBooks(city)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBook := make([]dto.BookResponse, 0, len(books))
	for _, b := range books {
		respBook = append(respBook, mapBookToResponse(b))
	}

	ctx.JSON(http.StatusOK, respBook)
}
