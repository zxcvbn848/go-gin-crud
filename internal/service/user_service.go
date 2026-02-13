package service

import (
	"context"
	"errors"
	"go-gin-crud/internal/cache"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/logger"
	"go-gin-crud/internal/repository"
	"math"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(id uint) (*dto.UserResponse, error)
	UpdateUser(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(id uint) error
	GetUsersWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error)
}

type userService struct {
	userRepo  repository.UserRepository
	userCache cache.UserCache
}

func NewUserService(userRepo repository.UserRepository, userCache cache.UserCache) UserService {
	return &userService{
		userRepo:  userRepo,
		userCache: userCache,
	}
}

func (s *userService) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	// 檢查 Email 是否已存在
	_, err := s.userRepo.FindByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email 已存在")
	}

	// 加密密碼
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		return nil, err
	}

	// 設定 role，如果沒有提供則使用預設值 "user"
	role := req.Role
	if role == "" {
		role = "user"
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hashed),
		Role:     role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	resp := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if s.userCache != nil {
		_ = s.userCache.SetUser(context.Background(), user.ID, resp)
		logger.Log.WithField("user_id", user.ID).Info("User 快取已寫入（Create）")
	}
	return resp, nil
}

func (s *userService) GetUserByID(id uint) (*dto.UserResponse, error) {
	ctx := context.Background()
	if s.userCache != nil {
		if cached, err := s.userCache.GetUser(ctx, id); err == nil && cached != nil {
			logger.Log.WithField("user_id", id).Info("User 從 Redis 快取取得")
			return cached, nil
		}
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	resp := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if s.userCache != nil {
		_ = s.userCache.SetUser(ctx, id, resp)
		logger.Log.WithField("user_id", id).Info("User 從 DB 取得並寫入快取")
	} else {
		logger.Log.WithField("user_id", id).Info("User 從 DB 取得（未啟用快取）")
	}
	return resp, nil
}

func (s *userService) UpdateUser(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 如果更新 email，檢查是否已存在
	if req.Email != "" && req.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(req.Email)
		if err == nil && existingUser.ID != user.ID {
			return nil, errors.New("email 已存在")
		}
		user.Email = req.Email
	}

	// 如果更新密碼，加密新密碼
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashed)
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	resp := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if s.userCache != nil {
		_ = s.userCache.SetUser(context.Background(), user.ID, resp)
		logger.Log.WithField("user_id", user.ID).Info("User 快取已更新（Update）")
	}
	return resp, nil
}

func (s *userService) DeleteUser(id uint) error {
	if err := s.userRepo.Delete(id); err != nil {
		return err
	}
	if s.userCache != nil {
		_ = s.userCache.DeleteUser(context.Background(), id)
		logger.Log.WithField("user_id", id).Info("User 快取已刪除（Delete）")
	}
	return nil
}

func (s *userService) GetUsersWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error) {
	// 設定預設值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	users, total, err := s.userRepo.FindAllWithPagination(req.Page, req.PageSize, req.Search)
	if err != nil {
		return nil, err
	}

	// 轉換為 UserResponse（不包含密碼）
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dto.PaginationResponse{
		Data:       userResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
