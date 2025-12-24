package dto

type UserCreateRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	City     string `json:"city"`
	Address  string `json:"address"`
}

type UserUpdateRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	City     *string `json:"city"`
	Address  *string `json:"address"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserProfileResponse struct {
	ID                       uint   `json:"id"`
	Name                     string `json:"name"`
	City                     string `json:"city"`
	BooksCount               int64  `json:"books_count"`
	SuccessfulExchangesCount int64  `json:"successful_exchanges_count"`
}
