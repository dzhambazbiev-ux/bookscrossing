package dto

type CreateBookRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	AISummary   string `json:"aisummary"`
}

type UpdateBookRequest struct {
	Description *string `json:"description"`
}
