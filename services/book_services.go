package services

import (
	"strings"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/repository"
)

type BookService interface {
	SearchBooks(query dto.BookListQuery) ([]models.Book, int64, error)
}

type bookService struct {
	bookRepo repository.BookRepository
}

func NewBookService(bookRepo repository.BookRepository) BookService {
	return &bookService{bookRepo: bookRepo}
}

func (s *bookService) SearchBooks(query dto.BookListQuery) ([]models.Book, int64, error) {
	query.SortBy = strings.ToLower(strings.TrimSpace(query.SortBy))
	query.SortOrder = strings.ToLower(strings.TrimSpace(query.SortOrder))

	if query.SortBy == "" {
		query.SortBy = "created_at"
	}

	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	return s.bookRepo.Search(query)
}
