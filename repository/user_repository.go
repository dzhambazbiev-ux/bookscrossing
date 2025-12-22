package repository

import (
	"github.com/dasler-fw/bookcrossing/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error

	GetByID(id uint) (*models.User, error)

	Update(user *models.User) error

	GetByEmail(email string) (*models.User, error)

	List() ([]models.User, error)

	Delete(id uint) error
}
