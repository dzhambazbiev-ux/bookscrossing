package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/jwtutil"
	"github.com/dasler-fw/bookcrossing/internal/models"
	"github.com/dasler-fw/bookcrossing/internal/repository"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	Register(req dto.UserCreateRequest) (string, error)
	Login(req dto.LoginRequest) (string, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(id uint, req dto.UserUpdateRequest) (*models.User, error)
	ListUsers(limit, offset int) ([]models.User, error)
	DeleteUser(id uint) error
	GetProfile(userID uint) (*dto.UserProfileResponse, error)
	UpdateProfile(userID uint, req dto.UserUpdateRequest) error
	GetUserExchanges(userID uint, status string) ([]models.Exchange, error)
}

type userService struct {
	db       *gorm.DB
	userRepo repository.UserRepository
	bookRepo repository.BookRepository
	log      *slog.Logger
	rdb      *redis.Client
}

func NewServiceUser(db *gorm.DB, userRepo repository.UserRepository, bookRepo repository.BookRepository, log *slog.Logger, rdb *redis.Client) UserService {
	return &userService{
		db:       db,
		userRepo: userRepo,
		bookRepo: bookRepo,
		log:      log,
		rdb:      rdb,
	}
}

func (s *userService) InvalidateUserList() {
	if s.rdb == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := s.rdb.Incr(ctx, "users:list:ver").Err(); err != nil {
		s.log.Error("user cache invalidation failed", "error", err)
		return
	}

	s.log.Info("user cache invalidated", "cache", "users:list", "method", "version bump")
}

func (s *userService) Register(req dto.UserCreateRequest) (string, error) {

	_, err := s.userRepo.GetByEmail(req.Email)
	if err == nil {
		return "", dto.ErrEmailAlreadyUsed
	}

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
	}

	if err := s.userRepo.Create(user); err != nil {
		return "", err
	}

	return jwtutil.GenerateToken(user.ID)
}

func (s *userService) Login(req dto.LoginRequest) (string, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return "", dto.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return "", dto.ErrInvalidCredentials
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
			return nil, dto.ErrUserPasswordHashFailed
		}
		user.PasswordHash = string(hash)
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, dto.ErrUserUpdateFailed
	}

	s.InvalidateUserList()
	return user, nil
}

func (s *userService) ListUsers(limit, offset int) ([]models.User, error) {

	if limit <= 0 {
		limit = dto.DefaultLimit
	}

	if limit > dto.MaxLimit {
		limit = dto.MaxLimit
	}

	if offset <= 0 {
		offset = 0
	}

	redisCtx, redisCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer redisCancel()

	ver, err := s.rdb.Get(redisCtx, "users:list:ver").Int64()
	if err != nil && err != redis.Nil {
		s.log.Error("redis get version in users list failed", "error", err)
		ver = 0
	}

	s.log.Info("user list cache version", "ver", ver)

	key := fmt.Sprintf("users:list:v=%d:l=%d:o:=%d", ver, limit, offset)

	cached, err := s.rdb.Get(redisCtx, key).Bytes()
	if err == nil {
		var userList []models.User
		if err := json.Unmarshal(cached, &userList); err == nil {
			s.log.Info("user cache hit", "key", key)
			return userList, nil
		}
		s.log.Error("user cache unmarshal failed", "key", key, "error", err)
	} else if err == redis.Nil {
		s.log.Info("user cache miss", "key", key)
	} else {
		s.log.Warn("user cache get failed", "key", key, "error", err)
	}

	users, err := s.userRepo.List(limit, offset)
	if err != nil {
		return nil, dto.ErrUserListFailed
	}

	b, err := json.Marshal(users)
	if err != nil {
		s.log.Error("user cache marshal failed", "error", err)
		return users, nil
	}

	setCtx, setCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer setCancel()

	if err := s.rdb.Set(setCtx, key, b, 10*time.Second); err != nil {
		s.log.Warn("user cache set failed", "key", key, "error", err)
	}

	return users, nil
}

func (s *userService) DeleteUser(id uint) error {
	if err := s.userRepo.Delete(id); err != nil {
		return dto.ErrUserDeleteFailed
	}
	return nil
}

func (s *userService) GetProfile(userID uint) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, repository.ErrUserNotFound
	}

	books, err := s.bookRepo.GetByUserID(userID, "")
	if err != nil {
		return nil, dto.ErrUserProfileFailed
	}

	var successfulExchanges int64
	if err := s.db.Model(&models.Exchange{}).
		Where("(initiator_id = ? OR recipient_id = ?) AND status = ?", userID, userID, "completed").
		Count(&successfulExchanges).Error; err != nil {
		return nil, dto.ErrUserProfileStatsFailed
	}
	return &dto.UserProfileResponse{
		ID:                       user.ID,
		Name:                     user.Name,
		City:                     user.City,
		BooksCount:               int64(len(books)),
		SuccessfulExchangesCount: successfulExchanges,
	}, nil
}

func (s *userService) UpdateProfile(userID uint, req dto.UserUpdateRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return repository.ErrUserNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.City != nil {
		user.City = *req.City
	}
	if req.Address != nil {
		user.Address = *req.Address
	}
	if err := s.userRepo.Update(user); err != nil {
		return dto.ErrUserProfileUpdateFailed
	}
	return nil
}

func (s *userService) GetUserExchanges(userID uint, status string) ([]models.Exchange, error) {
	var exchanges []models.Exchange

	q := s.db.Model(&models.Exchange{}).
		Where("initiator_id = ? OR recipient_id = ?", userID, userID)

	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Order("created_at desc").Find(&exchanges).Error; err != nil {
		return nil, dto.ErrUserExchangesFailed
	}

	return exchanges, nil
}
