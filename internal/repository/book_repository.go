package repository

import (
	"log/slog"
	"strings"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type BookRepository interface {
	Create(req *models.Book) error
	GetList() ([]models.Book, error)
	GetByID(id uint) (*models.Book, error)
	Update(book *models.Book) error
	Delete(id uint) error
	Search(query dto.BookListQuery) ([]models.Book, int64, error)
	AttachGenres(bookID uint, genreIDs []uint) error
	GetByUserID(userID uint, status string) ([]models.Book, error)
	GetAvailable(city string) ([]models.Book, error)
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
		return dto.ErrBookCreateFailed
	}

	return r.db.Create(req).Error
}

func (r *bookRepository) GetByID(id uint) (*models.Book, error) {
	var book models.Book
	if err := r.db.Preload("Genres").Preload("User").First(&book, id).Error; err != nil {
		r.log.Error("error in GetByID book_repository.go", "id", id, "err", err)
		return nil, dto.ErrBookGetFailed
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
		return dto.ErrBookUpdateFailed
	}

	return r.db.Save(book).Error
}

func (r *bookRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Book{}, id).Error; err != nil {
		r.log.Error("error in Delete function book_repository.go")
		return dto.ErrBookDeleteFailed
	}

	return nil
}

func (r *bookRepository) Search(query dto.BookListQuery) ([]models.Book, int64, error) {
	db := r.db.Model(&models.Book{})

	if query.GenreID != nil {
		db = db.Joins("JOIN book_genres bg ON bg.book_id = books.id").
			Where("bg.genre_id = ?", *query.GenreID)
	}

	if query.City != "" {
		db = db.Joins("JOIN users u ON u.id = books.user_id").
			Where("u.city ILIKE ?", "%"+query.City+"%")
	}

	if query.Author != "" {
		db = db.Where("books.author ILIKE ?", "%"+query.Author+"%")
	}

	if query.Title != "" {
		db = db.Where("books.title ILIKE ?", "%"+query.Title+"%")
	}

	if query.Status != "" {
		db = db.Where("books.status = ?", query.Status)
	}

	var total int64
	countQuery := db.Session(&gorm.Session{}).
		Select("COUNT(DISTINCT books.id)")

	if err := countQuery.Scan(&total).Error; err != nil {
		r.log.Error("ошибка считывании книг", "err", err)
		return nil, 0, err
	}

	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	sortOrder := strings.ToLower(strings.TrimSpace(query.SortOrder))

	validSortFields := map[string]string{
		"title":      "books.title",
		"created_at": "books.created_at",
	}

	sortField, ok := validSortFields[sortBy]
	if !ok {
		sortField = "books.created_at"
	}

	validOrders := map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
	}

	order, ok := validOrders[sortOrder]
	if !ok {
		order = "DESC"
	}

	offset := (query.Page - 1) * query.Limit

	var books []models.Book

	subQuery := db.Session(&gorm.Session{}).
		Select("books.id", sortField).
		Distinct("books.id", sortField).
		Order(sortField + " " + order).
		Limit(query.Limit).
		Offset(offset)

	if err := db.Where("books.id IN (SELECT books.id FROM (?) AS sorted_books)", subQuery).
		Preload("Genres").
		Preload("User").
		Order(sortField + " " + order).
		Find(&books).Error; err != nil {
		r.log.Error("ошибка при поиске книг", "err", err)
		return nil, 0, err
	}

	return books, total, nil
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

func (r *bookRepository) GetByUserID(userID uint, status string) ([]models.Book, error) {
	var books []models.Book

	db := r.db.Model(&models.Book{}).Where("user_id = ?", userID)

	if status != "" {
		db = db.Where("status = ?", strings.TrimSpace(status))
	}

	if err := db.Preload("Genres").
		Preload("User").
		Order("created_at DESC").
		Find(&books).Error; err != nil {
		r.log.Error("Ошибка в функции GetByUserID book_repository.go", "err", err)
		return nil, err
	}

	return books, nil
}

func (r *bookRepository) GetAvailable(city string) ([]models.Book, error) {
	var books []models.Book

	db := r.db.Model(&models.Book{}).
		Where("books.status = ?", "available")

	city = strings.TrimSpace(city)
	if city != "" {
		db = db.Joins("JOIN users u ON u.id = books.user_id").
			Where("u.city ILIKE ?", "%"+city+"%")
	}

	if err := db.Preload("Genres").
		Preload("User").
		Order("created_at DESC").
		Find(&books).Error; err != nil {
		r.log.Error("Ошибка в функции GetAvailable book_repository.go", "err", err)
		return nil, err
	}

	return books, nil
}
