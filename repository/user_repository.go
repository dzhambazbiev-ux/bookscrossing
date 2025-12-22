package repository

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	Update(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	List() ([]models.User, error)
	Delete(id uint) error
}

type userRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewUserRepository(db *gorm.DB, log *slog.Logger) UserRepository {
	return &userRepository{
		db:  db,
		log: log,
	}
}

func (r *userRepository) Create(user *models.User) error {
	if user == nil {
		r.log.Error("ошибка создания профиля")
		return errors.New("пустой пользователь")
	}
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User

	if err := r.db.First(&user, id).Error; err != nil {
		r.log.Error("ошибка получения профиля по ID")
		return nil, errors.New("пользователь не найден")
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	if user == nil {
		r.log.Error("ошибка обновления профиля")
		return errors.New("пустой пользователь")
	}
	return r.db.Save(user).Error
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		r.log.Error("ошибка получения профиля по Email")
		return nil, errors.New("пользователь не найден")
	}
	return &user, nil
}

func (r *userRepository) List() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		r.log.Error("ошибка получения списка пользователей")
		return nil, errors.New("не удалось получить список пользователей")
	}
	return users, nil
}
func (r *userRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.User{}, id).Error; err != nil {
		r.log.Error("ошибка удаления профиля")
		return errors.New("не удалось удалить пользователя")
	}
	return nil
}
