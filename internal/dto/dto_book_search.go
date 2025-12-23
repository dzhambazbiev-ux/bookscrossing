package dto

type BookListQuery struct {
	// Фильтры
	GenreID *uint  `form:"genre_id"`
	City    string `form:"city"`
	Author  string `form:"author"`
	Status  string `form:"status"`
	Title   string `form:"title"`

	// Пагинация
	Page  int `form:"page"`
	Limit int `form:"limit"`

	// Сортировка
	// sort_by: created_at | title
	// sort_order: asc | desc
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)
