package services

import (
	"errors"
	"log/slog"
	"time"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/jwtutil"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(req dto.UserCreateRequest) (string, error)
	Login(req dto.LoginRequest) (string, error)
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

func (s *userService) Register(req dto.UserCreateRequest) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		City:         req.City,
		Address:      req.Address,
		RegisteredAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return "", err
	}

	return jwtutil.GenerateToken(user.ID)
}

func (s *userService) Login(req dto.LoginRequest) (string, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return "", errors.New("invalid credentials")
	}

	return jwtutil.GenerateToken(user.ID)
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
