package services

import (
	"errors"
	"log/slog"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(req dto.UserCreateRequest) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(id uint, req dto.UserUpdateRequest) (*models.User, error)
	ListUsers() ([]models.User, error)
	DeleteUser(id uint) error
}

type userService struct {
	userRepo repository.UserRepository
	log      *slog.Logger
}

func NewServiceUser(userRepo repository.UserRepository, log *slog.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		log:      log,
	}
}

func (s *userService) CreateUser(req dto.UserCreateRequest) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("ошибка хеширования пароля", "email", req.Email, "err", err)
		return nil, errors.New("ошибка обработки пароля")
	}
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		City:         req.City,
		Address:      req.Address,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("ошибка создания пользователя")

	}
	return user, nil
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (s *userService) UpdateUser(id uint, req dto.UserUpdateRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, repository.ErrUserNotFound
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}

	if req.City != nil {
		user.City = *req.City
	}

	if req.Address != nil {
		user.Address = *req.Address
	}
	if req.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.log.Error("ошибка хеширования пароля", "id", id, "err", err)
			return nil, errors.New("ошибка обработки пароля")
		}
		user.PasswordHash = string(hash)
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("ошибка обновления пользователя")
	}
	return user, nil
}

func (s *userService) ListUsers() ([]models.User, error) {
	users, err := s.userRepo.List()
	if err != nil {
		return nil, errors.New("ошибка получения списка пользователей")
	}
	return users, nil
}

func (s *userService) DeleteUser(id uint) error {
	if err := s.userRepo.Delete(id); err != nil {
		return errors.New("ошибка удаления пользователя")
	}
	return nil
}
