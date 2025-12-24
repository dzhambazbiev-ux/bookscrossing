package dto

import "time"

type UserPublicResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
}

type GenreResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type BookResponse struct {
	ID          uint              `json:"id"`
	Title       string            `json:"title"`
	Author      string            `json:"author"`
	Description string            `json:"description"`
	AISummary   string            `json:"ai_summary"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	Owner       UserPublicResponse `json:"owner"`
	Genres      []GenreResponse   `json:"genres"`
}

type BookListResponse struct {
	Data       []BookResponse `json:"data"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int            `json:"total"`
	TotalPages int            `json:"total_pages"`
}
