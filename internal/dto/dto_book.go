package dto

type CreateBookRequest struct {
	Title       string
	Author      string
	Description string
	AISummary   string
}

type UpdateBookRequest struct {
	Description *string
}