package dto

import "time"

type UserCreateRequest struct {
	Name         string
	Email        string
	PasswordHash string
	City         string
	Address      string
	RegisteredAt time.Time
}

type UserUpdateRequest struct {
	Name         *string
	Email        *string
	PasswordHash *string
	City         *string
	Address      *string
}
