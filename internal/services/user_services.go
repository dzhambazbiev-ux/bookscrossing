package services

import (
	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
)

type UserService interface {
	CreateUser(req dto.UserCreateRequest) (*models.User, error)

	GetUserByID(id uint) (*models.User, error)

	UpdateUser(id uint, req dto.UserUpdateRequest) (*models.User, error)

	ListUsers() ([]models.User, error)

	DeleteUser(id uint) error
}
