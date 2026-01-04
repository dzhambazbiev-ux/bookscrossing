package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	redisTimeout = 80 * time.Millisecond
	listTTL      = 10 * time.Second
	searchTTL    = 10 * time.Second
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
	bookRepo repository.BookRepository
	log      *slog.Logger
	rdb      *redis.Client
}

func NewServiceBook(bookRepo repository.BookRepository, log *slog.Logger, rdb *redis.Client) BookService {
	svc := &bookService{
		bookRepo: bookRepo,
		log:      log,
		rdb:      rdb,
	}

	return svc
}

func (s *bookService) invalidateListCache() {
	if s.rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := s.rdb.Incr(ctx, "books:list:ver").Err(); err != nil {
		s.log.Error("cache invalidation failed", "error", err)
		return
	}
	s.log.Info("cache invalidated", "cache", "books:list", "method", "version bump")
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

	s.invalidateListCache()
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
	//  Готовим контекст с таймаутом для Redis
	redisCtx, redisCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer redisCancel()

	// Читаем "версию" списка (если ключа нет — считаем 0)
	ver, err := s.rdb.Get(redisCtx, "books:list:ver").Int64()
	if err != nil && err != redis.Nil {
		s.log.Error("redis get version failed", "error", err)
		ver = 0
	}

	s.log.Info("list cache version", "ver", ver)

	// Ключ кэша включает версию + limit/offset
	key := fmt.Sprintf("books:list:v=%d:l=%d:o=%d", ver, limit, offset)

	// Пробуем взять из кэша
	cached, err := s.rdb.Get(redisCtx, key).Bytes()
	if err == nil {
		var list []models.Book
		if err := json.Unmarshal(cached, &list); err == nil {
			s.log.Info("cache hit", "key", key)
			return list, nil
		}
		s.log.Error("cache unmarshal failed", "key", key, "error", err)
		// если JSON битый — идём в БД как будто кэша нет
	} else if err == redis.Nil {
		//s.log.Error("cache get failed", "key", key, "error", err)
		s.log.Info("cache miss", "key", key)
		// если Redis глючит — идём в БД как будто кэша нет
	} else {
		s.log.Warn("cache get failed", "key", key, "error", err)
		//s.log.Info("cache miss", "key", key)
	}

	// Идем в БД
	list, err := s.bookRepo.GetList(limit, offset)
	if err != nil {
		return nil, err
	}

	// Пытаемся сохранить в кэш (best-effort)
	// b, err := json.Marshal(list)
	// if err == nil {
	// 	_ = s.rdb.Set(redisCtx, key, b, listTTL).Err()
	// } else {
	// 	s.log.Error("cache marshal failed", "error", err)
	// }

	b, err := json.Marshal(list)
	if err != nil {
		s.log.Error("cache marshal failed", "error", err)
		return list, nil
	}

	setCtx, setCancel := context.WithTimeout(context.Background(), redisTimeout)
	defer setCancel()

	if err := s.rdb.Set(setCtx, key, b, listTTL).Err(); err != nil {
		s.log.Warn("cache set failed", "key", key, "error", err)
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

	s.invalidateListCache()
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

	if err := s.bookRepo.Delete(bookID); err != nil {
		return err
	}

	s.invalidateListCache()
	return nil
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

	var cacheKey string

	if s.rdb != nil {
		redisCtx, redisCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer redisCancel()

		ver, err := s.rdb.Get(redisCtx, "books:list:ver").Int64()
		if err != nil && err != redis.Nil {
			s.log.Error("redis get version failed", "error", err)
			ver = 0
		}

		// Делаем короткий стабильный ключ:
		// 1) JSON от query (после нормализации)
		// 2) sha256 хэш (чтобы ключ не был длинным)
		qb, err := json.Marshal(query)
		if err != nil {
			s.log.Warn("cache key marshal failed", "error", err)
		} else {
			sum := sha256.Sum256(qb)
			qHash := hex.EncodeToString(sum[:])
			cacheKey = fmt.Sprintf("books:search:v=%d:%s", ver, qHash)

			// Пробуем взять из кэша
			cached, err := s.rdb.Get(redisCtx, cacheKey).Bytes()
			if err == nil {
				// В кэше держим и список книг, и total (общее количество)
				var payload struct {
					Books []models.Book `json:"books"`
					Total int64         `json:"total"`
				}

				if err := json.Unmarshal(cached, &payload); err == nil {
					s.log.Info("cache hit", "key", cacheKey)
					return payload.Books, payload.Total, nil
				}

				s.log.Error("cache unmarshal failed", "key", cacheKey, "error", err)

				// если JSON битый — просто идём в БД
			} else if err == redis.Nil {
				s.log.Info("cache miss", "key", cached)
			} else {
				s.log.Warn("cache get failed", "key", cacheKey, "error", err)
			}
		}
	}

	// Идем в БД
	books, total, err := s.bookRepo.Search(query)
	if err != nil {
		return nil, 0, err
	}

	// Пытаемся сохранить в кэш (best-effort)
	if s.rdb != nil && cacheKey != "" {
		payload := struct {
			Books []models.Book `json:"books"`
			Total int64         `json:"total"`
		}{
			Books: books,
			Total: total,
		}

		b, err := json.Marshal(payload)
		if err != nil {
			s.log.Error("cache marshal failed", "key", cacheKey, "error", err)
			return books, total, nil
		}

		setCtx, setCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer setCancel()

		if err := s.rdb.Set(setCtx, cacheKey, b, searchTTL).Err(); err != nil {
			s.log.Warn("cache set failed", "key", cacheKey, "error", err)
		}
	}

	return books, total, nil
}

func (s *bookService) GetBooksByUserID(userID uint, status string) ([]models.Book, error) {
	return s.bookRepo.GetByUserID(userID, status)
}

func (s *bookService) GetAvailableBooks(city string) ([]models.Book, error) {
	return s.bookRepo.GetAvailable(city)
}