package dto

type UserCreateRequest struct {
	Name     string
	Email    string
	Password string
	City     string
	Address  string
}

type UserUpdateRequest struct {
	Name     *string
	Email    *string
	Password *string
	City     *string
	Address  *string
}
