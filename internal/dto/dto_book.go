package dto

type CreateBookRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	AISummary   string `json:"ai_summary"`
	GenreIDs    []uint `json:"genre_ids"` // для привязки жанров
}

type UpdateBookRequest struct {
	Description *string `json:"description"`
}
