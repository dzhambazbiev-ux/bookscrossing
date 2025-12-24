package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/repository"
)

type BookService interface {
	CreateBook(userID uint, ras dto.CreateBookRequest) (*models.Book, error)
	GetByID(id uint) (*models.Book, error)
	GetList() ([]models.Book, error)
	Update(bookID uint, userID uint, req dto.UpdateBookRequest) (*models.Book, error)
	Delete(bookID uint, userID uint) error
	SearchBooks(query dto.BookListQuery) ([]models.Book, int64, error)
	GetBooksByUserID(userID uint, status string) ([]models.Book, error)
	GetAvailableBooks(city string) ([]models.Book, error)
}

type bookService struct {
	bookRepo repository.BookRepository
	log      *slog.Logger
}

func NewServiceBook(bookRepo repository.BookRepository, log *slog.Logger) BookService {
	return &bookService{
		bookRepo: bookRepo,
		log:      log,
	}
}

func (s *bookService) CreateBook(userID uint, req dto.CreateBookRequest) (*models.Book, error) {
	book := &models.Book{
		Title:       req.Title,
		Author:      req.Author,
		Description: req.Description,
		Status:      "available",
		UserID:      userID,
	}

	// Если AISummary пустой, генерируем через Grok AI
	if req.AISummary == "" {
		summary, err := GenerateAISummary(req.Description)
		if err != nil {
			return nil, err
		}
		book.AISummary = summary
	} else {
		book.AISummary = req.AISummary
	}

	// Сохраняем книгу
	if err := s.bookRepo.Create(book); err != nil {
		return nil, err
	}

	// Привязываем жанры
	if len(req.GenreIDs) > 0 {
		if err := s.bookRepo.AttachGenres(book.ID, req.GenreIDs); err != nil {
			return nil, err
		}
	}

	return book, nil
}

func (s *bookService) GetByID(id uint) (*models.Book, error) {
	book, err := s.bookRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (s *bookService) GetList() ([]models.Book, error) {
	list, err := s.bookRepo.GetList()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *bookService) Update(bookID uint, userID uint, req dto.UpdateBookRequest) (*models.Book, error) {
	book, err := s.bookRepo.GetByID(bookID)
	if err != nil {
		return nil, err
	}

	if book.UserID != userID {
		return nil, dto.ErrBookForbidden
	}

	if req.Description != nil {
		book.Description = *req.Description
	}

	if err := s.bookRepo.Update(book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *bookService) Delete(bookID uint, userID uint) error {
	book, err := s.bookRepo.GetByID(bookID)
	if err != nil {
		return err
	}

	if book.UserID != userID {
		return dto.ErrBookForbidden
	}

	if book.Status == "pending" || book.Status == "accepted" {
		return dto.ErrBookInExchange
	}

	return s.bookRepo.Delete(bookID)
}

func GenerateAISummary(description string) (string, error) {
	apiKey := os.Getenv("GROK_API_KEY")
	if apiKey == "" {
		return "", dto.ErrAISummaryFailed
	}

	payload := map[string]string{
		"prompt": "Сделай краткое резюме книги: " + description,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://api.grok.ai/v1/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if err := resp.Body.Close(); err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", errors.New("grok api error: " + string(b))
	}

	var result map[string]interface{}
	dec := json.NewDecoder(io.LimitReader(resp.Body, 10*1024)) // limit 10KB
	if err := dec.Decode(&result); err != nil {
		return "", err
	}

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if text, ok := choice["text"].(string); ok {
				return text, nil
			}
		}
	}

	// fallback: если структура другая — попытаться найти "text" или "message"
	if t, ok := result["text"].(string); ok && t != "" {
		return t, nil
	}

	return "", errors.New("empty summary from AI")
}

func (s *bookService) SearchBooks(query dto.BookListQuery) ([]models.Book, int64, error) {
	if query.Page <= 0 {
		query.Page = dto.DefaultPage
	}

	if query.Limit <= 0 {
		query.Limit = dto.DefaultLimit
	}

	if query.Limit > dto.MaxLimit {
		query.Limit = dto.MaxLimit
	}
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

func (s *bookService) GetBooksByUserID(userID uint, status string) ([]models.Book, error) {
	return s.bookRepo.GetByUserID(userID, status)
}

func (s *bookService) GetAvailableBooks(city string) ([]models.Book, error) {
	return s.bookRepo.GetAvailable(city)
}
