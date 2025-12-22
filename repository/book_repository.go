package repository

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type BookRepository interface {
	Create(req *models.Book) error
	GetList() ([]models.Book, error)
	GetByID(id uint) (*models.Book, error)
	Update(book *models.Book) error
	Delete(id uint) error
	AttachGenres(bookID uint, genreIDs []uint) error
}

type bookRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewBookRepository(db *gorm.DB, log *slog.Logger) BookRepository {
	return &bookRepository{
		db:  db,
		log: log,
	}
}

func (r *bookRepository) Create(req *models.Book) error {
	if req == nil {
		r.log.Error("error in Create function book_repository.go")
		return errors.New("error create category in db")
	}

	return r.db.Create(req).Error
}

func (r *bookRepository) GetByID(id uint) (*models.Book, error) {
	var book models.Book
	if err := r.db.Preload("Genres").Preload("User").First(&book, id).Error; err != nil {
		r.log.Error("error in Delete function book_repository.go")
		return nil, errors.New("error delete in db")
	}

	return &book, nil
}

func (r *bookRepository) GetList() ([]models.Book, error) {
	var list []models.Book
	if err := r.db.Preload("Genres").Find(&list).Error; err != nil {
		r.log.Error("error in List function book_repository.go")
		return nil, err
	}

	return list, nil
}

func (r *bookRepository) Update(book *models.Book) error {
	if book == nil {
		r.log.Error("error in Update function book_repository.go")
		return errors.New("error update in db")
	}

	return r.db.Save(book).Error
}

func (r *bookRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Book{}, id).Error; err != nil {
		r.log.Error("error in Delete function book_repository.go")
		return errors.New("error delete in db")
	}

	return nil
}

func (r *bookRepository) AttachGenres(bookID uint, genreIDs []uint) error {
	var book models.Book
	if err := r.db.First(&book, bookID).Error; err != nil {
		return err
	}

	var genres []models.Genre
	if err := r.db.Where("id IN ?", genreIDs).Find(&genres).Error; err != nil {
		return err
	}

	// Привязываем жанры к книге
	if err := r.db.Model(&book).Association("Genres").Replace(genres); err != nil {
		return err
	}

	return nil
}
