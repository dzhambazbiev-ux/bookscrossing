package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dasler-fw/bookcrossing/internal/cache"
	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/repository"
)

type BookService interface {
	CreateBook(userID uint, ras dto.CreateBookRequest) (*models.Book, error)
	GetByID(id uint) (*models.Book, error)
	GetList(limit, offset int) ([]models.Book, error)
	Update(bookID uint, userID uint, req dto.UpdateBookRequest) (*models.Book, error)
	Delete(bookID uint, userID uint) error
	SearchBooks(query dto.BookListQuery) ([]models.Book, int64, error)
	GetBooksByUserID(userID uint, status string) ([]models.Book, error)
	GetAvailableBooks(city string) ([]models.Book, error)
}

type bookService struct {
	bookRepo  repository.BookRepository
	log       *slog.Logger
	listCache *cache.TTLCache[string, []models.Book]
}

func NewServiceBook(bookRepo repository.BookRepository, log *slog.Logger) BookService {
	svc := &bookService{
		bookRepo:  bookRepo,
		log:       log,
		listCache: cache.NewTTLCache[string, []models.Book](10 * time.Second),
	}

	svc.listCache.StartJanitor(context.Background(), 2*time.Second, func(removed int, size int) {
		log.Info("cache sweep", "removed", removed, "size", size)
	})
	return svc
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

func (s *bookService) GetList(limit, offset int) ([]models.Book, error) {
	key := fmt.Sprintf("books:list:l=%d:o=%d", limit, offset)
	if v, ok := s.listCache.Get(key); ok {
		s.log.Info("cache hit", "key", key)
		return v, nil
	}

	s.log.Info("cache miss", "key", key)

	list, err := s.bookRepo.GetList(limit, offset)
	if err != nil {
		return nil, err
	}

	s.listCache.Set(key, list)
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
