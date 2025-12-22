package dto

type CreateBookRequest struct {
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	AISummary   string   `json:"ai_summary"`
	GenreIDs    []uint   `json:"genre_ids"` // для привязки жанров
}


type UpdateBookRequest struct {
	Title       *string  `json:"title,omitempty"`
	Author      *string  `json:"author,omitempty"`
	Description *string  `json:"description,omitempty"`
	AISummary   *string  `json:"ai_summary,omitempty"`
	GenreIDs    []uint   `json:"genre_ids,omitempty"` // можно менять жанры
}
