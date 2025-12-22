package dto

type GenreCreateRequest struct {
	Name string
}

type GenreUpdateRequest struct {
	Name *string
}
