package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/repository"
)

type BookService interface {
	CreateBook(userID uint, ras dto.CreateBookRequest) (*models.Book, error)
	GetByID(id uint) (*models.Book, error)
	GetList() ([]models.Book, error)
	Update(bookID uint, userID uint, req dto.UpdateBookRequest) (*models.Book, error)
	Delete(id uint) error
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

	return book, err
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
        return nil, errors.New("только владелец может редактировать книгу")
    }

    if req.Title != nil {
        book.Title = *req.Title
    }
    if req.Description != nil {
        book.Description = *req.Description
    }
    if req.AISummary != nil {
        book.AISummary = *req.AISummary
    }

    if len(req.GenreIDs) > 0 {
        if err := s.bookRepo.AttachGenres(book.ID, req.GenreIDs); err != nil {
            return nil, err
        }
    }

    if err := s.bookRepo.Update(book); err != nil {
        return nil, err
    }

    return book, nil
}


func (s *bookService) Delete(id uint) error {
	book, err := s.bookRepo.GetByID(id)
	if err != nil {
		return err
	}
	if book.Status == "pending" || book.Status == "accepted" {
		return errors.New("нельзя удалить книгу, участвующую в обмене")
	}

	return  nil
}

func GenerateAISummary(description string) (string, error) {
	apiKey := "GROK_API_KEY"

	payload := map[string]string{
		"prompt": "Сделай краткое резюме книги: " + description,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.grok.ai/v1/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// пример, как получить текст из ответа
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if text, ok := choice["text"].(string); ok {
				return text, nil
			}
		}
	}

	return "", nil
}
